package gitcache

import (
	"fmt"
	"go-kweb-lang/git"
	"go-kweb-lang/gitcache/internal"
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

func (c *GitRepoCache) Create(url string) error {
	return c.gitRepo.Create(url)
}

func (c *GitRepoCache) FileExists(path string) (bool, error) {
	return c.gitRepo.FileExists(path)
}

func (c *GitRepoCache) ListFiles(path string) ([]string, error) {
	return c.gitRepo.ListFiles(path)
}

func (c *GitRepoCache) FindFileLastCommit(path string) (git.CommitInfo, error) {
	key := path
	return cacheWrapper(internal.FileLastCommitDir(c.cacheDir), key, func() (git.CommitInfo, error) {
		return c.gitRepo.FindFileLastCommit(path)
	})
}

func (c *GitRepoCache) FindFileCommitsAfter(path string, commitIdFrom string) ([]git.CommitInfo, error) {
	key := path
	return cacheWrapper(internal.FileUpdatesDir(c.cacheDir), key, func() ([]git.CommitInfo, error) {
		return c.gitRepo.FindFileCommitsAfter(path, commitIdFrom)
	})
}

func (c *GitRepoCache) FindMergePoints(commitId string) ([]git.CommitInfo, error) {
	key := commitId
	return cacheWrapper(internal.MergePointsDir(c.cacheDir), key, func() ([]git.CommitInfo, error) {
		return c.gitRepo.FindMergePoints(commitId)
	})
}

func (c *GitRepoCache) Fetch() error {
	return c.gitRepo.Fetch()
}

func (c *GitRepoCache) FreshCommits() ([]git.CommitInfo, error) {
	return c.gitRepo.FreshCommits()
}

func (c *GitRepoCache) Pull() error {
	return c.gitRepo.Pull()
}

func (c *GitRepoCache) CommitFiles(commitId string) ([]string, error) {
	// todo: do we need to cache it? probably not
	return c.gitRepo.CommitFiles(commitId)
}

func (c *GitRepoCache) InvalidatePath(path string) error {
	for _, cacheFile := range []string{
		filepath.Join(internal.FileLastCommitDir(c.cacheDir), internal.KeyFile(internal.KeyHash(path))),
		filepath.Join(internal.FileUpdatesDir(c.cacheDir), internal.KeyFile(internal.KeyHash(path))),
	} {
		if err := removeFile(cacheFile); err != nil {
			return err
		}
	}

	return nil
}

func (c *GitRepoCache) PullRefresh() error {
	if err := c.gitRepo.Fetch(); err != nil {
		return fmt.Errorf("git fetch error: %w", err)
	}
	freshCommits, err := c.gitRepo.FreshCommits()
	if err != nil {
		return fmt.Errorf("git list fresh commits error: %w", err)
	}
	for _, fc := range freshCommits {
		commitFiles, err := c.gitRepo.CommitFiles(fc.CommitId)
		if err != nil {
			return fmt.Errorf("git list files of commit %s error: %w", fc.CommitId, err)
		}
		for _, f := range commitFiles {
			if err := c.InvalidatePath(f); err != nil {
				return fmt.Errorf("git cache invalidate path %s error: %w", f, err)
			}
		}
	}
	if err := c.gitRepo.Pull(); err != nil {
		return fmt.Errorf("git pull error: %w", err)
	}
	return nil
}

func cacheWrapper[T any](cacheDir string, key string, block func() (T, error)) (T, error) {
	if err := internal.EnsureDir(cacheDir); err != nil {
		var zero T
		return zero, err
	}

	hash := internal.KeyHash(key)
	cachePath := filepath.Join(cacheDir, internal.KeyFile(hash))
	if internal.FileExists(cachePath) {
		var buff T
		if err := internal.ReadJsonFromFile(cachePath, &buff); err != nil {
			return buff, err
		}

		return buff, nil
	}

	result, err := block()
	if err != nil {
		return result, err
	}

	if err := internal.WriteJsonToFile(cachePath, result); err != nil {
		var zero T
		return zero, err
	}

	return result, nil
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
