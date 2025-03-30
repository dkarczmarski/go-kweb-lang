// Package gitpc is a cache proxy for git repository.
// The name stands for git proxy cache.
package gitpc

import (
	"context"
	"fmt"

	"go-kweb-lang/git"
	"go-kweb-lang/proxycache"
)

const (
	CategoryLastCommit  = "git-file-last-commit"
	CategoryUpdates     = "git-file-updates"
	CategoryMergePoints = "git-merge-points"
)

type GitRepoProxyCache struct {
	gitRepo  git.Repo
	cacheDir string
}

func New(gitRepo git.Repo, cacheDir string) *GitRepoProxyCache {
	return &GitRepoProxyCache{
		gitRepo:  gitRepo,
		cacheDir: cacheDir,
	}
}

func (pc *GitRepoProxyCache) Create(ctx context.Context, url string) error {
	return pc.gitRepo.Create(ctx, url)
}

func (pc *GitRepoProxyCache) FileExists(path string) (bool, error) {
	return pc.gitRepo.FileExists(path)
}

func (pc *GitRepoProxyCache) ListFiles(path string) ([]string, error) {
	return pc.gitRepo.ListFiles(path)
}

func (pc *GitRepoProxyCache) FindFileLastCommit(ctx context.Context, path string) (git.CommitInfo, error) {
	key := path
	return proxycache.Get(
		ctx,
		pc.cacheDir,
		CategoryLastCommit,
		key,
		nil,
		func(ctx context.Context) (git.CommitInfo, error) {
			return pc.gitRepo.FindFileLastCommit(ctx, path)
		},
	)
}

func (pc *GitRepoProxyCache) FindFileCommitsAfter(ctx context.Context, path string, commitIDFrom string) ([]git.CommitInfo, error) {
	key := path
	return proxycache.Get(
		ctx,
		pc.cacheDir,
		CategoryUpdates,
		key,
		nil,
		func(ctx context.Context) ([]git.CommitInfo, error) {
			return pc.gitRepo.FindFileCommitsAfter(ctx, path, commitIDFrom)
		},
	)
}

func (pc *GitRepoProxyCache) FindMergePoints(ctx context.Context, commitID string) ([]git.CommitInfo, error) {
	key := commitID
	return proxycache.Get(
		ctx,
		pc.cacheDir,
		CategoryMergePoints,
		key,
		nil,
		func(ctx context.Context) ([]git.CommitInfo, error) {
			return pc.gitRepo.FindMergePoints(ctx, commitID)
		},
	)
}

func (pc *GitRepoProxyCache) Fetch(ctx context.Context) error {
	return pc.gitRepo.Fetch(ctx)
}

func (pc *GitRepoProxyCache) FreshCommits(ctx context.Context) ([]git.CommitInfo, error) {
	return pc.gitRepo.FreshCommits(ctx)
}

func (pc *GitRepoProxyCache) Pull(ctx context.Context) error {
	return pc.gitRepo.Pull(ctx)
}

func (pc *GitRepoProxyCache) CommitFiles(ctx context.Context, commitID string) ([]string, error) {
	return pc.gitRepo.CommitFiles(ctx, commitID)
}

func (pc *GitRepoProxyCache) InvalidatePath(path string) error {
	for _, category := range []string{
		CategoryLastCommit,
		CategoryUpdates,
	} {
		key := path
		if err := proxycache.InvalidateKey(pc.cacheDir, category, key); err != nil {
			return fmt.Errorf("error while invalidataing cache key %v: %w", key, err)
		}
	}

	return nil
}

func (pc *GitRepoProxyCache) PullRefresh(ctx context.Context) error {
	if err := pc.gitRepo.Fetch(ctx); err != nil {
		return fmt.Errorf("git fetch error: %w", err)
	}

	freshCommits, err := pc.gitRepo.FreshCommits(ctx)
	if err != nil {
		return fmt.Errorf("git list fresh commits error: %w", err)
	}

	for _, fc := range freshCommits {
		commitFiles, err := pc.gitRepo.CommitFiles(ctx, fc.CommitID)
		if err != nil {
			return fmt.Errorf("git list files of commit %s error: %w", fc.CommitID, err)
		}

		for _, f := range commitFiles {
			if err := pc.InvalidatePath(f); err != nil {
				return fmt.Errorf("git cache invalidate path %s error: %w", f, err)
			}
		}
	}
	if err := pc.gitRepo.Pull(ctx); err != nil {
		return fmt.Errorf("git pull error: %w", err)
	}
	return nil
}
