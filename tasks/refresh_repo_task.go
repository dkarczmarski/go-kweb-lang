package tasks

import (
	"context"
	"fmt"
	"go-kweb-lang/gitpc"
)

type RefreshRepoTask struct {
	gitRepoCache *gitpc.GitRepoCache
}

func NewRefreshRepoTask(gitRepoCache *gitpc.GitRepoCache) *RefreshRepoTask {
	return &RefreshRepoTask{
		gitRepoCache: gitRepoCache,
	}
}

func (t *RefreshRepoTask) Run(ctx context.Context) error {
	if err := t.gitRepoCache.PullRefresh(ctx); err != nil {
		return fmt.Errorf("git cache pull refresh error: %w", err)
	}

	return nil
}
