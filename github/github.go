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
	"strings"
	"time"
)

type Config struct {
	BaseURL string
	Token   string
}

type GitHub struct {
	BaseURL    string
	HTTPClient *http.Client
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

func NewGitHub(opts ...func(*Config)) *GitHub {
	config := Config{
		BaseURL: "https://api.github.com",
	}

	for _, opt := range opts {
		opt(&config)
	}

	var transport http.RoundTripper

	if len(config.Token) > 0 {
		transport = &authorizationTransport{
			Token: config.Token,
		}
	}

	httpClient := &http.Client{
		Timeout:   time.Minute,
		Transport: transport,
	}

	return &GitHub{
		BaseURL:    config.BaseURL,
		HTTPClient: httpClient,
	}
}

func (gh *GitHub) GetLatestCommit(ctx context.Context) (*CommitInfo, error) {
	urlStr := fmt.Sprintf("%v/repos/kubernetes/website/commits?per_page=1", gh.BaseURL)

	resp, err := gh.httpGet(ctx, urlStr)
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

	resp, err := gh.httpGet(ctx, urlStr)
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
	baseURL := fmt.Sprintf("%v/search/issues", gh.BaseURL)

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
	urlStr := fmt.Sprintf("%v/repos/kubernetes/website/pulls/%d/commits", gh.BaseURL, prNumber)

	resp, err := gh.httpGet(ctx, urlStr)
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
	urlStr := fmt.Sprintf("%v/repos/kubernetes/website/commits/%s", gh.BaseURL, commitID)

	resp, err := gh.httpGet(ctx, urlStr)
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

func (gh *GitHub) httpGet(ctx context.Context, urlStr string) (*http.Response, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, urlStr, nil)
	if err != nil {
		return nil, fmt.Errorf("error while creating request: %w", err)
	}

	resp, err := gh.HTTPClient.Do(req)
	if err != nil {
		if isTimeoutErr(err) {
			err = wrapRetryableErr(err)
		}

		return nil, fmt.Errorf("error while sending request: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		defer resp.Body.Close()

		body, _ := io.ReadAll(resp.Body)

		if isRateLimitError(resp, body) {
			err = wrapRetryableErr(errors.New("API rate limit exceeded"))
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

func isRateLimitError(resp *http.Response, body []byte) bool {
	if resp == nil {
		return false
	}

	if resp.StatusCode == http.StatusForbidden || resp.StatusCode == http.StatusTooManyRequests {
		if remaining := resp.Header.Get("X-RateLimit-Remaining"); remaining == "0" {
			return true
		}

		if len(body) > 0 && bytes.Contains(body, []byte("rate limit exceeded")) {
			return true
		}
	}

	return false
}

type retryErr struct {
	err error
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

func wrapRetryableErr(err error) error {
	return &retryErr{err}
}
