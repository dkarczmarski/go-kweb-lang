package github

//go:generate mockgen -typed -source=github.go -destination=../mocks/mock_github.go -package=mocks

import (
	"context"
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

	PRSearch(ctx context.Context, filter PRSearchFilter, page PageRequest) (*PRSearchResult, error)

	GetPRCommits(ctx context.Context, prNumber int) ([]string, error)

	GetCommitFiles(ctx context.Context, commitID string) (*CommitFiles, error)
}

func New(opts ...func(*ClientConfig)) GitHub {
	return NewClient(opts...)
}
