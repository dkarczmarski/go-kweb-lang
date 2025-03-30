// Package gitcache is a cached wrapper for git repository.
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

type GitRepoCache struct {
	gitRepo  git.Repo
	cacheDir string
}

func New(gitRepo git.Repo, cacheDir string) *GitRepoCache {
	return &GitRepoCache{
		gitRepo:  gitRepo,
		cacheDir: cacheDir,
	}
}

func (c *GitRepoCache) Create(ctx context.Context, url string) error {
	return c.gitRepo.Create(ctx, url)
}

func (c *GitRepoCache) FileExists(path string) (bool, error) {
	return c.gitRepo.FileExists(path)
}

func (c *GitRepoCache) ListFiles(path string) ([]string, error) {
	return c.gitRepo.ListFiles(path)
}

func (c *GitRepoCache) FindFileLastCommit(ctx context.Context, path string) (git.CommitInfo, error) {
	key := path
	return proxycache.Get(
		ctx,
		c.cacheDir,
		CategoryLastCommit,
		key,
		nil,
		func(ctx context.Context) (git.CommitInfo, error) {
			return c.gitRepo.FindFileLastCommit(ctx, path)
		},
	)
}

func (c *GitRepoCache) FindFileCommitsAfter(ctx context.Context, path string, commitIDFrom string) ([]git.CommitInfo, error) {
	key := path
	return proxycache.Get(
		ctx,
		c.cacheDir,
		CategoryUpdates,
		key,
		nil,
		func(ctx context.Context) ([]git.CommitInfo, error) {
			return c.gitRepo.FindFileCommitsAfter(ctx, path, commitIDFrom)
		},
	)
}

func (c *GitRepoCache) FindMergePoints(ctx context.Context, commitID string) ([]git.CommitInfo, error) {
	key := commitID
	return proxycache.Get(
		ctx,
		c.cacheDir,
		CategoryMergePoints,
		key,
		nil,
		func(ctx context.Context) ([]git.CommitInfo, error) {
			return c.gitRepo.FindMergePoints(ctx, commitID)
		},
	)
}

func (c *GitRepoCache) Fetch(ctx context.Context) error {
	return c.gitRepo.Fetch(ctx)
}

func (c *GitRepoCache) FreshCommits(ctx context.Context) ([]git.CommitInfo, error) {
	return c.gitRepo.FreshCommits(ctx)
}

func (c *GitRepoCache) Pull(ctx context.Context) error {
	return c.gitRepo.Pull(ctx)
}

func (c *GitRepoCache) CommitFiles(ctx context.Context, commitID string) ([]string, error) {
	// todo: do we need to cache it? probably not
	return c.gitRepo.CommitFiles(ctx, commitID)
}

func (c *GitRepoCache) InvalidatePath(path string) error {
	for _, category := range []string{
		CategoryLastCommit,
		CategoryUpdates,
	} {
		key := path
		if err := proxycache.InvalidateKey(c.cacheDir, category, key); err != nil {
			return fmt.Errorf("error while invalidataing cache key %v: %w", key, err)
		}
	}

	return nil
}

func (c *GitRepoCache) PullRefresh(ctx context.Context) error {
	if err := c.gitRepo.Fetch(ctx); err != nil {
		return fmt.Errorf("git fetch error: %w", err)
	}

	freshCommits, err := c.gitRepo.FreshCommits(ctx)
	if err != nil {
		return fmt.Errorf("git list fresh commits error: %w", err)
	}

	for _, fc := range freshCommits {
		commitFiles, err := c.gitRepo.CommitFiles(ctx, fc.CommitID)
		if err != nil {
			return fmt.Errorf("git list files of commit %s error: %w", fc.CommitID, err)
		}

		for _, f := range commitFiles {
			if err := c.InvalidatePath(f); err != nil {
				return fmt.Errorf("git cache invalidate path %s error: %w", f, err)
			}
		}
	}
	if err := c.gitRepo.Pull(ctx); err != nil {
		return fmt.Errorf("git pull error: %w", err)
	}
	return nil
}
