package tasks

import (
	"context"
	"fmt"

	"go-kweb-lang/githist"
)

type RefreshRepoTask struct {
	gitRepoHist *githist.GitHist
}

func NewRefreshRepoTask(gitRepoHist *githist.GitHist) *RefreshRepoTask {
	return &RefreshRepoTask{
		gitRepoHist: gitRepoHist,
	}
}

func (t *RefreshRepoTask) Run(ctx context.Context) error {
	if err := t.gitRepoHist.PullRefresh(ctx); err != nil {
		return fmt.Errorf("git cache pull refresh error: %w", err)
	}

	return nil
}
