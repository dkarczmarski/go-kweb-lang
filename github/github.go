package github

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"
)

type GitHub struct {
	httpClient *http.Client
}

func New() *GitHub {
	httpClient := &http.Client{
		Timeout: time.Minute,
	}

	return &GitHub{
		httpClient: httpClient,
	}
}

type CommitInfo struct {
	CommitID string
	DateTime string
}

func (gh *GitHub) GetLatestCommitAfter(ctx context.Context, dateTimeAfter string) (*CommitInfo, error) {
	return gh.getCommit(ctx, "https://api.github.com/repos/kubernetes/website/commits?per_page=1&since="+dateTimeAfter)
}

func (gh *GitHub) GetLatestCommit(ctx context.Context) (*CommitInfo, error) {
	return gh.getCommit(ctx, "https://api.github.com/repos/kubernetes/website/commits?per_page=1")
}

func (gh *GitHub) getCommit(ctx context.Context, url string) (*CommitInfo, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, fmt.Errorf("request creation error: %w", err)
	}

	resp, err := gh.httpClient.Do(req)
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
