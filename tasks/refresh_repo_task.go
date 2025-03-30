package tasks

import (
	"context"
	"fmt"

	"go-kweb-lang/gitpc"
)

type RefreshRepoTask struct {
	gitRepoProxyCache *gitpc.ProxyCache
}

func NewRefreshRepoTask(gitRepoProxyCache *gitpc.ProxyCache) *RefreshRepoTask {
	return &RefreshRepoTask{
		gitRepoProxyCache: gitRepoProxyCache,
	}
}

func (t *RefreshRepoTask) Run(ctx context.Context) error {
	if err := t.gitRepoProxyCache.PullRefresh(ctx); err != nil {
		return fmt.Errorf("git cache pull refresh error: %w", err)
	}

	return nil
}
