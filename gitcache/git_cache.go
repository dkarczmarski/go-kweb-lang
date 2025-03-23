// Package gitcache is a cached wrapper for git repository.
package gitcache

import (
	"context"
	"fmt"
	"go-kweb-lang/filecache"
	"go-kweb-lang/git"
	"os"
	"path/filepath"
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
	return filecache.CacheWrapperCtx(ctx, FileLastCommitDir(c.cacheDir), key, func() (git.CommitInfo, error) {
		return c.gitRepo.FindFileLastCommit(ctx, path)
	})
}

func (c *GitRepoCache) FindFileCommitsAfter(ctx context.Context, path string, commitIDFrom string) ([]git.CommitInfo, error) {
	key := path
	return filecache.CacheWrapperCtx(ctx, FileUpdatesDir(c.cacheDir), key, func() ([]git.CommitInfo, error) {
		return c.gitRepo.FindFileCommitsAfter(ctx, path, commitIDFrom)
	})
}

func (c *GitRepoCache) FindMergePoints(ctx context.Context, commitID string) ([]git.CommitInfo, error) {
	key := commitID
	return filecache.CacheWrapperCtx(ctx, MergePointsDir(c.cacheDir), key, func() ([]git.CommitInfo, error) {
		return c.gitRepo.FindMergePoints(ctx, commitID)
	})
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
	for _, cacheFile := range []string{
		filepath.Join(FileLastCommitDir(c.cacheDir), filecache.KeyFile(filecache.KeyHash(path))),
		filepath.Join(FileUpdatesDir(c.cacheDir), filecache.KeyFile(filecache.KeyHash(path))),
	} {
		if err := removeFile(cacheFile); err != nil {
			return err
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

func removeFile(path string) error {
	_, err := os.Stat(path)
	if err != nil {
		if os.IsNotExist(err) {
			// skip silently
			return nil
		}
		return fmt.Errorf("failed to check file: %w", err)
	}

	err = os.Remove(path)
	if err != nil {
		return fmt.Errorf("failed to remove file %s: %w", path, err)
	}

	return nil
}
