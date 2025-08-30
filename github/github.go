// Package github provides information about the Kubernetes GitHub repository.
package github

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"net"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"

	"go-kweb-lang/github/internal/throttle"
)

type Config struct {
	BaseURL          string
	HTTPClient       *http.Client
	Token            string
	UserAgent        string
	ThrottleInterval time.Duration
}

type GitHub struct {
	baseURL    string
	httpClient *http.Client
	throttler  *throttle.Throttler
}

type CommitInfo struct {
	CommitID string
	DateTime string
}

type PRSearchFilter struct {
	OnlyOpen    bool
	LangCode    string
	UpdatedFrom string
}

type PageRequest struct {
	Sort    string
	Order   string
	Page    int
	PerPage int
}

type PRSearchResult struct {
	Items      []PRItem `json:"items"`
	TotalCount int      `json:"total_count"`
}

type PRItem struct {
	Number    int    `json:"number"`
	UpdatedAt string `json:"updated_at"`
}

type CommitFiles struct {
	CommitID string
	Files    []string
}

func WithDefaults() func(*Config) {
	return func(config *Config) {
		config.BaseURL = "https://api.github.com"
		config.HTTPClient = &http.Client{
			Timeout: time.Minute,
		}
	}
}

func WithAuthorization(token, userAgent string) func(config *Config) {
	return func(config *Config) {
		var transport http.RoundTripper

		if len(token) > 0 {
			transport = &authorizationTransport{
				Token:     token,
				UserAgent: userAgent,
			}

			config.HTTPClient.Transport = transport
		}
	}
}

func WithThrottle(interval time.Duration) func(config *Config) {
	return func(config *Config) {
		if interval > 0 {
			config.ThrottleInterval = interval
		}
	}
}

func NewGitHub(opts ...func(*Config)) *GitHub {
	var config Config

	for _, opt := range opts {
		opt(&config)
	}

	var throttler *throttle.Throttler
	if config.ThrottleInterval > 0 {
		throttler = throttle.NewThrottler(config.ThrottleInterval)
	}

	return &GitHub{
		baseURL:    config.BaseURL,
		httpClient: config.HTTPClient,
		throttler:  throttler,
	}
}

func (gh *GitHub) GetLatestCommit(ctx context.Context) (*CommitInfo, error) {
	urlStr := fmt.Sprintf("%v/repos/kubernetes/website/commits?per_page=1", gh.baseURL)

	resp, err := gh.httpGetWithRetry(ctx, urlStr)
	if err != nil {
		return nil, err
	}

	defer resp.Body.Close()

	var commits []commitResponse
	if err := json.NewDecoder(resp.Body).Decode(&commits); err != nil {
		return nil, fmt.Errorf("json error: %w", err)
	}

	if len(commits) == 0 {
		return nil, nil
	}

	commitInfo := CommitInfo{
		CommitID: commits[0].SHA,
		DateTime: commits[0].Commit.Committer.Date,
	}

	return &commitInfo, nil
}

type commitResponse struct {
	SHA    string `json:"sha"`
	Commit struct {
		Committer struct {
			Date string `json:"date"`
		} `json:"author"`
	} `json:"commit"`
}

func (gh *GitHub) PRSearch(ctx context.Context, filter PRSearchFilter, page PageRequest) (*PRSearchResult, error) {
	urlStr := gh.buildPRSearchURL(filter, page)

	resp, err := gh.httpGetWithRetry(ctx, urlStr)
	if err != nil {
		return nil, err
	}

	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("error while reading response body: %w", err)
	}

	var result PRSearchResult
	if err := json.Unmarshal(body, &result); err != nil {
		return nil, fmt.Errorf("error while parsing JSON response: %w", err)
	}

	return &result, nil
}

func (gh *GitHub) buildPRSearchURL(filter PRSearchFilter, page PageRequest) string {
	baseURL := fmt.Sprintf("%v/search/issues", gh.baseURL)

	qParts := []string{
		"repo:kubernetes/website",
		"is:pr",
	}

	if filter.OnlyOpen {
		qParts = append(qParts, "state:open")
	}

	if len(filter.LangCode) > 0 {
		qParts = append(qParts, "label:language/"+filter.LangCode)
	}

	if len(filter.UpdatedFrom) > 0 {
		// format: updated:>2024-12-01
		qParts = append(qParts, "updated:>"+filter.UpdatedFrom)
	}

	q := strings.Join(qParts, "+")

	query := url.Values{}

	if len(page.Sort) > 0 {
		query.Set("sort", page.Sort)
	}

	if len(page.Order) > 0 {
		query.Set("order", page.Order)
	}

	if page.PerPage > 0 {
		query.Set("per_page", fmt.Sprintf("%v", page.PerPage))
	}

	query.Set("page", "1")

	u, err := url.Parse(baseURL)
	if err != nil {
		log.Fatal(fmt.Errorf("error while parsing base URL: %w", err))
	}

	u.RawQuery = fmt.Sprintf("q=%s&%s", q, query.Encode())

	return u.String()
}

func (gh *GitHub) GetPRCommits(ctx context.Context, prNumber int) ([]string, error) {
	urlStr := fmt.Sprintf("%v/repos/kubernetes/website/pulls/%d/commits", gh.baseURL, prNumber)

	resp, err := gh.httpGetWithRetry(ctx, urlStr)
	if err != nil {
		return nil, err
	}

	defer resp.Body.Close()

	var commits []commitItem
	if err := json.NewDecoder(resp.Body).Decode(&commits); err != nil {
		return nil, fmt.Errorf("error while decoding JSON: %w", err)
	}

	var commitIDs []string
	for _, commit := range commits {
		commitIDs = append(commitIDs, commit.SHA)
	}

	return commitIDs, nil
}

type commitItem struct {
	SHA string `json:"sha"`
}

func (gh *GitHub) GetCommitFiles(ctx context.Context, commitID string) (*CommitFiles, error) {
	urlStr := fmt.Sprintf("%v/repos/kubernetes/website/commits/%s", gh.baseURL, commitID)

	resp, err := gh.httpGetWithRetry(ctx, urlStr)
	if err != nil {
		return nil, err
	}

	defer resp.Body.Close()

	var detail commitModel
	if err := json.NewDecoder(resp.Body).Decode(&detail); err != nil {
		return nil, fmt.Errorf("error while decoding JSON: %w", err)
	}

	var files []string
	for _, f := range detail.Files {
		files = append(files, f.Filename)
	}

	return &CommitFiles{
		detail.SHA,
		files,
	}, nil
}

type commitModel struct {
	SHA   string `json:"sha"`
	Files []struct {
		Filename string `json:"filename"`
	} `json:"files"`
}

func (gh *GitHub) httpGetWithRetry(ctx context.Context, urlStr string) (*http.Response, error) {
	var resp *http.Response
	var err error

	maxRetries := 10 // magic number
	for i := 0; i < maxRetries; i++ {
		if i > 0 {
			log.Printf("[%d/%d] retry http GET %s", i, maxRetries, urlStr)
		}

		resp, err = gh.httpGet(ctx, urlStr)
		if err == nil {
			break
		}

		var terr *retryErr
		if errors.As(err, &terr) {
			if len(terr.retryAfterStr) > 0 {
				if seconds, err := strconv.Atoi(terr.retryAfterStr); err == nil {
					log.Printf("received http code %d with Retry-After header. wait for %d seconds",
						terr.statusCode, seconds)
					if err := sleepCtx(ctx, time.Duration(seconds)*time.Second); err != nil {
						return nil, err
					}
				}
			} else if len(terr.remainingStr) > 0 && len(terr.resetStr) > 0 {
				remaining, _ := strconv.Atoi(terr.remainingStr)
				resetUnix, _ := strconv.ParseInt(terr.resetStr, 10, 64)

				if remaining == 0 && resetUnix > 0 {
					resetTime := time.Unix(resetUnix, 0)
					wait := time.Until(resetTime)
					if wait > 0 {
						log.Printf("received http code %d with X-RateLimit-Reset header. wait for %v seconds until reset at %v",
							terr.statusCode, wait.Seconds(), resetTime)
						if err := sleepCtx(ctx, wait); err != nil {
							return nil, err
						}
					}
				}
			} else {
				log.Printf("wait for 1 minute. connection error: %v", err)
				if err := sleepCtx(ctx, time.Minute); err != nil {
					return nil, err
				}
			}
		} else {
			break
		}
	}

	return resp, err
}

func sleepCtx(ctx context.Context, d time.Duration) error {
	timer := time.NewTimer(d)
	defer timer.Stop()

	select {
	case <-ctx.Done():
		return ctx.Err()
	case <-timer.C:
		return nil
	}
}

func (gh *GitHub) httpGet(ctx context.Context, urlStr string) (*http.Response, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, urlStr, nil)
	if err != nil {
		return nil, fmt.Errorf("error while creating request: %w", err)
	}

	if gh.throttler != nil {
		// small delay to avoid secondary rate limit
		if err := gh.throttler.Throttle(ctx); err != nil {
			return nil, err
		}
	}

	resp, err := gh.httpClient.Do(req)
	if err != nil {
		if isTimeoutErr(err) {
			err = &retryErr{err: err}
		}

		return nil, fmt.Errorf("error while sending request: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		defer resp.Body.Close()

		body, _ := io.ReadAll(resp.Body)

		if ok, err := isRateLimitError(resp, body); ok {
			return nil, err
		}

		return nil, fmt.Errorf("error while reading response: %w\nstatus: %s\nbody: %s",
			err, resp.Status, string(body))
	}

	return resp, err
}

func isTimeoutErr(err error) bool {
	var netErr net.Error
	return errors.As(err, &netErr) && netErr.Timeout()
}

func isRateLimitError(resp *http.Response, body []byte) (bool, error) {
	if resp == nil {
		return false, nil
	}

	if resp.StatusCode == http.StatusForbidden || resp.StatusCode == http.StatusTooManyRequests {
		remainingStr := resp.Header.Get("X-RateLimit-Remaining")
		resetStr := resp.Header.Get("X-RateLimit-Reset")
		retryAfterStr := resp.Header.Get("Retry-After")

		if len(body) > 0 && bytes.Contains(body, []byte("rate limit exceeded")) {
			return true, &retryErr{
				err:           errors.New("API rate limit exceeded"),
				statusCode:    resp.StatusCode,
				remainingStr:  remainingStr,
				resetStr:      resetStr,
				retryAfterStr: retryAfterStr,
			}
		}
	}

	return false, nil
}

type retryErr struct {
	err error

	statusCode    int
	remainingStr  string
	resetStr      string
	retryAfterStr string
}

func (e *retryErr) Error() string {
	return e.err.Error()
}

func (e *retryErr) Unwrap() error {
	return e.err
}

func (e *retryErr) IsRetryable() bool {
	return true
}
