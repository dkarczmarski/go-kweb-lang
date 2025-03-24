package github

//go:generate mockgen -typed -source=github.go -destination=../mocks/mock_github.go -package=mocks

import (
	"context"
	"net/http"
	"time"
)

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

type GitHub interface {
	GetLatestCommit(ctx context.Context) (*CommitInfo, error)

	PRSearch(filter PRSearchFilter, page PageRequest) (*PRSearchResult, error)

	GetPRCommits(prNumber int) ([]string, error)

	GetCommitFiles(commitID string) (*CommitFiles, error)
}

func New() GitHub {
	httpClient := &http.Client{
		Timeout: time.Minute,
	}

	return &Client{
		BaseURL:    "https://api.github.com",
		HTTPClient: httpClient,
	}
}
