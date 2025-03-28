package github

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"strings"
	"time"
)

type ClientConfig struct {
	BaseURL string
	Token   string
}

type Client struct {
	BaseURL    string
	HTTPClient *http.Client
}

func NewClient(opts ...func(*ClientConfig)) *Client {
	config := ClientConfig{
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

	return &Client{
		BaseURL:    config.BaseURL,
		HTTPClient: httpClient,
	}
}

func (gh *Client) GetLatestCommit(ctx context.Context) (*CommitInfo, error) {
	return gh.getCommit(ctx, fmt.Sprintf("%v/repos/kubernetes/website/commits?per_page=1", gh.BaseURL))
}

func (gh *Client) getCommit(ctx context.Context, url string) (*CommitInfo, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, fmt.Errorf("request creation error: %w", err)
	}

	resp, err := gh.HTTPClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("HTTP request error: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("GitHub API http code: %d", resp.StatusCode)
	}

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

func (gh *Client) PRSearch(filter PRSearchFilter, page PageRequest) (*PRSearchResult, error) {
	urlStr := gh.buildPRSearchURL(filter, page)

	resp, err := gh.HTTPClient.Get(urlStr)
	if err != nil {
		return nil, fmt.Errorf("error while sending request to GitHub API: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("error while reading GitHub API response: status %s", resp.Status)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("error while reading response body: %v", err)
	}

	var result PRSearchResult
	if err := json.Unmarshal(body, &result); err != nil {
		return nil, fmt.Errorf("error while parsing JSON response: %v", err)
	}

	return &result, nil
}

func (gh *Client) buildPRSearchURL(filter PRSearchFilter, page PageRequest) string {
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
		log.Fatal(fmt.Errorf("error while parsing base URL: %v", err))
	}

	u.RawQuery = fmt.Sprintf("q=%s&%s", q, query.Encode())

	return u.String()
}

func (gh *Client) GetPRCommits(prNumber int) ([]string, error) {
	urlStr := fmt.Sprintf("%v/repos/kubernetes/website/pulls/%d/commits", gh.BaseURL, prNumber)

	req, err := http.NewRequest("GET", urlStr, nil)
	if err != nil {
		return nil, fmt.Errorf("error while creating request: %v", err)
	}

	resp, err := gh.HTTPClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("error while sending request: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("error while reading response: status %s\nBody: %s", resp.Status, string(body))
	}

	var commits []commitItem
	if err := json.NewDecoder(resp.Body).Decode(&commits); err != nil {
		return nil, fmt.Errorf("error while decoding JSON: %v", err)
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

func (gh *Client) GetCommitFiles(commitID string) (*CommitFiles, error) {
	urlStr := fmt.Sprintf("%v/repos/kubernetes/website/commits/%s", gh.BaseURL, commitID)

	req, err := http.NewRequest("GET", urlStr, nil)
	if err != nil {
		return nil, fmt.Errorf("error while creating request: %v", err)
	}

	resp, err := gh.HTTPClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("error while sending request: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("error while reading response: status %s\nbody: %s", resp.Status, string(body))
	}

	var detail commitModel
	if err := json.NewDecoder(resp.Body).Decode(&detail); err != nil {
		return nil, fmt.Errorf("error while decoding JSON: %v", err)
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
