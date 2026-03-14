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

	"github.com/dkarczmarski/go-kweb-lang/github/internal/throttle"
)

const (
	defaultBaseURL        = "https://api.github.com"
	maxHTTPRetries        = 10
	defaultRetryWait      = time.Minute
	retryResetSafetyDelay = 3 * time.Second
)

var (
	ErrUnexpectedHTTPStatus = errors.New("unexpected http status")
	ErrRateLimitExceeded    = errors.New("api rate limit exceeded")
	ErrSecondaryRateLimit   = errors.New("secondary api rate limit exceeded")
	ErrInvalidBaseURL       = errors.New("invalid github base url")
	ErrNoCommitsFound       = errors.New("no commits found")
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

//nolint:tagliatelle
type PRSearchResult struct {
	Items      []PRItem `json:"items"`
	TotalCount int      `json:"total_count"`
}

//nolint:tagliatelle
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
		config.BaseURL = defaultBaseURL
		//nolint:exhaustruct
		config.HTTPClient = &http.Client{
			Timeout: time.Minute,
		}
	}
}

func WithAuthorization(token, userAgent string) func(*Config) {
	return func(config *Config) {
		if len(token) == 0 {
			return
		}

		config.HTTPClient.Transport = &authorizationTransport{
			Token:     token,
			UserAgent: userAgent,
		}
	}
}

func WithThrottle(interval time.Duration) func(*Config) {
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

	var throttlerInstance *throttle.Throttler
	if config.ThrottleInterval > 0 {
		throttlerInstance = throttle.NewThrottler(config.ThrottleInterval)
	}

	return &GitHub{
		baseURL:    config.BaseURL,
		httpClient: config.HTTPClient,
		throttler:  throttlerInstance,
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
		return nil, fmt.Errorf("decode latest commit response: %w", err)
	}

	if len(commits) == 0 {
		return nil, ErrNoCommitsFound
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

func (gh *GitHub) PRSearch(
	ctx context.Context,
	filter PRSearchFilter,
	page PageRequest,
) (*PRSearchResult, error) {
	filter.LangCode = toShortLangCode(filter.LangCode)

	urlStr, err := gh.buildPRSearchURL(filter, page)
	if err != nil {
		return nil, fmt.Errorf("build PR search URL: %w", err)
	}

	resp, err := gh.httpGetWithRetry(ctx, urlStr)
	if err != nil {
		return nil, err
	}

	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("read PR search response body: %w", err)
	}

	var result PRSearchResult
	if err := json.Unmarshal(body, &result); err != nil {
		return nil, fmt.Errorf("parse PR search JSON response: %w", err)
	}

	return &result, nil
}

// toShortLangCode shortens compound language codes, for example zh-cn -> zh, pt-br -> pt.
func toShortLangCode(langCode string) string {
	if len(langCode) > 2 && langCode[2] == '-' {
		langCode = langCode[:2]
	}

	return langCode
}

func (gh *GitHub) buildPRSearchURL(filter PRSearchFilter, page PageRequest) (string, error) {
	baseURL := fmt.Sprintf("%v/search/issues", gh.baseURL)

	queryParts := []string{
		"repo:kubernetes/website",
		"is:pr",
	}

	if filter.OnlyOpen {
		queryParts = append(queryParts, "state:open")
	}

	if len(filter.LangCode) > 0 {
		queryParts = append(queryParts, "label:language/"+filter.LangCode)
	}

	if len(filter.UpdatedFrom) > 0 {
		queryParts = append(queryParts, "updated:>"+filter.UpdatedFrom)
	}

	queryText := strings.Join(queryParts, "+")

	queryValues := url.Values{}

	if len(page.Sort) > 0 {
		queryValues.Set("sort", page.Sort)
	}

	if len(page.Order) > 0 {
		queryValues.Set("order", page.Order)
	}

	if page.PerPage > 0 {
		queryValues.Set("per_page", strconv.Itoa(page.PerPage))
	}

	queryValues.Set("page", "1")

	parsedURL, err := url.Parse(baseURL)
	if err != nil {
		return "", fmt.Errorf("%w: %w", ErrInvalidBaseURL, err)
	}

	parsedURL.RawQuery = fmt.Sprintf("q=%s&%s", queryText, queryValues.Encode())

	return parsedURL.String(), nil
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
		return nil, fmt.Errorf("decode PR commits JSON: %w", err)
	}

	commitIDs := make([]string, 0, len(commits))
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
		return nil, fmt.Errorf("decode commit files JSON: %w", err)
	}

	files := make([]string, 0, len(detail.Files))
	for _, file := range detail.Files {
		files = append(files, file.Filename)
	}

	return &CommitFiles{
		CommitID: detail.SHA,
		Files:    files,
	}, nil
}

type commitModel struct {
	SHA   string `json:"sha"`
	Files []struct {
		Filename string `json:"filename"`
	} `json:"files"`
}

func (gh *GitHub) httpGetWithRetry(ctx context.Context, urlStr string) (*http.Response, error) {
	var (
		resp *http.Response
		err  error
	)

	for i := range maxHTTPRetries {
		if i > 0 {
			log.Printf("[%d/%d] retry http GET %s", i, maxHTTPRetries, urlStr)
		}

		resp, err = gh.httpGet(ctx, urlStr)
		if err == nil {
			break
		}

		var retryableErr *retryError
		if !errors.As(err, &retryableErr) {
			break
		}

		log.Printf("connection error: %+v", retryableErr)

		if waitErr := waitForRetry(ctx, retryableErr); waitErr != nil {
			return nil, waitErr
		}
	}

	return resp, err
}

func waitForRetry(ctx context.Context, retryableErr *retryError) error {
	switch {
	case len(retryableErr.retryAfterStr) > 0:
		return waitForRetryAfterHeader(ctx, retryableErr)
	case len(retryableErr.resetStr) > 0:
		return waitForRateLimitReset(ctx, retryableErr)
	default:
		log.Printf("wait for %v", defaultRetryWait)

		return sleepCtx(ctx, defaultRetryWait)
	}
}

func waitForRetryAfterHeader(ctx context.Context, retryableErr *retryError) error {
	seconds, err := strconv.Atoi(retryableErr.retryAfterStr)
	if err != nil {
		log.Printf("invalid Retry-After header value: %q", retryableErr.retryAfterStr)

		return sleepCtx(ctx, defaultRetryWait)
	}

	log.Printf(
		"received http code %d with Retry-After header. wait for %d seconds",
		retryableErr.statusCode,
		seconds,
	)

	return sleepCtx(ctx, time.Duration(seconds)*time.Second)
}

func waitForRateLimitReset(ctx context.Context, retryableErr *retryError) error {
	resetUnix, err := strconv.ParseInt(retryableErr.resetStr, 10, 64)
	if err != nil || resetUnix <= 0 {
		log.Printf("invalid X-RateLimit-Reset header value: %q", retryableErr.resetStr)

		return sleepCtx(ctx, defaultRetryWait)
	}

	resetTime := time.Unix(resetUnix, 0)
	waitDuration := time.Until(resetTime) + retryResetSafetyDelay

	if waitDuration <= 0 {
		return nil
	}

	log.Printf(
		"received http code %d with X-RateLimit-Reset header. wait for %v seconds until reset at %v",
		retryableErr.statusCode,
		waitDuration.Seconds(),
		resetTime,
	)

	return sleepCtx(ctx, waitDuration)
}

func sleepCtx(ctx context.Context, d time.Duration) error {
	timer := time.NewTimer(d)
	defer timer.Stop()

	select {
	case <-ctx.Done():
		return fmt.Errorf("sleep interrupted by context: %w", ctx.Err())
	case <-timer.C:
		return nil
	}
}

func (gh *GitHub) httpGet(ctx context.Context, urlStr string) (*http.Response, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, urlStr, nil)
	if err != nil {
		return nil, fmt.Errorf("create request: %w", err)
	}

	if gh.throttler != nil {
		if err := gh.throttler.Throttle(ctx); err != nil {
			return nil, fmt.Errorf("github throttling failed: %w", err)
		}
	}

	resp, err := gh.httpClient.Do(req)
	if err != nil {
		if isTimeoutErr(err) {
			//nolint:exhaustruct
			err = &retryError{err: err}
		}

		return nil, fmt.Errorf("send request: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		defer resp.Body.Close()

		body, _ := io.ReadAll(resp.Body)

		if ok, rateLimitErr := isRateLimitError(resp, body); ok {
			return nil, rateLimitErr
		}

		return nil, fmt.Errorf(
			"%w: status=%s body=%s",
			ErrUnexpectedHTTPStatus,
			resp.Status,
			string(body),
		)
	}

	return resp, nil
}

func isTimeoutErr(err error) bool {
	var netErr net.Error

	return errors.As(err, &netErr) && netErr.Timeout()
}

func isRateLimitError(resp *http.Response, body []byte) (bool, error) {
	if resp == nil {
		return false, nil
	}

	if resp.StatusCode != http.StatusForbidden && resp.StatusCode != http.StatusTooManyRequests {
		return false, nil
	}

	// GitHub API uses:
	// - X-RateLimit-Remaining
	// - X-RateLimit-Reset
	remainingStr := resp.Header.Get("X-Ratelimit-Remaining")
	resetStr := resp.Header.Get("X-Ratelimit-Reset")
	retryAfterStr := resp.Header.Get("Retry-After")

	switch {
	case len(body) > 0 && bytes.Contains(body, []byte("rate limit exceeded")):
		return true, &retryError{
			err:           ErrRateLimitExceeded,
			statusCode:    resp.StatusCode,
			remainingStr:  remainingStr,
			resetStr:      resetStr,
			retryAfterStr: retryAfterStr,
		}
	case len(body) > 0 && bytes.Contains(body, []byte("secondary rate limit")):
		return true, &retryError{
			err:           ErrSecondaryRateLimit,
			statusCode:    resp.StatusCode,
			remainingStr:  remainingStr,
			resetStr:      resetStr,
			retryAfterStr: retryAfterStr,
		}
	}

	return false, nil
}

type retryError struct {
	err error

	statusCode    int
	remainingStr  string
	resetStr      string
	retryAfterStr string
}

func (e *retryError) Error() string {
	return e.err.Error()
}

func (e *retryError) Unwrap() error {
	return e.err
}

func (e *retryError) IsRetryable() bool {
	return true
}
