// Package gitpc is a cache proxy for git repository.
// The name stands for git proxy cache.
package gitpc

import (
	"context"
	"fmt"

	"go-kweb-lang/git"
	"go-kweb-lang/proxycache"
)

// todo: should it be private ?
const (
	CategoryLastCommit        = "git-file-last-commit"
	CategoryUpdates           = "git-file-updates"
	CategoryMergePoints       = "git-merge-points"
	CategoryAncestorCommits   = "git-ancestor-commits"
	CategoryMainBranchCommits = "git-main-branch-commits"
)

type ProxyCache struct {
	gitRepo  git.Repo
	cacheDir string
}

func New(gitRepo git.Repo, cacheDir string) *ProxyCache {
	return &ProxyCache{
		gitRepo:  gitRepo,
		cacheDir: cacheDir,
	}
}

// Create function is a plain proxy wrapper to git.Repo.
func (pc *ProxyCache) Create(ctx context.Context, url string) error {
	return pc.gitRepo.Create(ctx, url)
}

// Checkout function is a plain proxy wrapper to git.Repo.
func (pc *ProxyCache) Checkout(ctx context.Context, commitID string) error {
	return pc.gitRepo.Checkout(ctx, commitID)
}

// ListMainBranchCommits function is a cache proxy wrapper to git.Repo.
func (pc *ProxyCache) ListMainBranchCommits(ctx context.Context) ([]git.CommitInfo, error) {
	return proxycache.Get(
		ctx,
		pc.cacheDir,
		CategoryMainBranchCommits,
		"",
		nil,
		func(ctx context.Context) ([]git.CommitInfo, error) {
			return pc.gitRepo.ListMainBranchCommits(ctx)
		},
	)
}

func (pc *ProxyCache) invalidateMainBranchCommits() error {
	return proxycache.InvalidateKey(pc.cacheDir, CategoryMainBranchCommits, "")
}

// FileExists function is a plain proxy wrapper to git.Repo.
func (pc *ProxyCache) FileExists(path string) (bool, error) {
	return pc.gitRepo.FileExists(path)
}

// ListFiles function is a plain proxy wrapper to git.Repo.
func (pc *ProxyCache) ListFiles(path string) ([]string, error) {
	return pc.gitRepo.ListFiles(path)
}

// FindFileLastCommit function is a cache proxy wrapper to git.Repo.
func (pc *ProxyCache) FindFileLastCommit(ctx context.Context, path string) (git.CommitInfo, error) {
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

// FindFileCommitsAfter function is a cache proxy wrapper to git.Repo.
func (pc *ProxyCache) FindFileCommitsAfter(ctx context.Context, path string, commitIDFrom string) ([]git.CommitInfo, error) {
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

// ListMergePoints function is a cache proxy wrapper to git.Repo.
func (pc *ProxyCache) ListMergePoints(ctx context.Context, commitID string) ([]git.CommitInfo, error) {
	key := commitID

	return proxycache.Get(
		ctx,
		pc.cacheDir,
		CategoryMergePoints,
		key,
		nil,
		func(ctx context.Context) ([]git.CommitInfo, error) {
			return pc.gitRepo.ListMergePoints(ctx, commitID)
		},
	)
}

// Fetch function is a cache proxy wrapper to git.Repo.
func (pc *ProxyCache) Fetch(ctx context.Context) error {
	return pc.gitRepo.Fetch(ctx)
}

// ListFreshCommits function is a cache proxy wrapper to git.Repo.
func (pc *ProxyCache) ListFreshCommits(ctx context.Context) ([]git.CommitInfo, error) {
	return pc.gitRepo.ListFreshCommits(ctx)
}

// Pull function is a cache proxy wrapper to git.Repo.
func (pc *ProxyCache) Pull(ctx context.Context) error {
	return pc.gitRepo.Pull(ctx)
}

// ListFilesInCommit function is a cache proxy wrapper to git.Repo.
func (pc *ProxyCache) ListFilesInCommit(ctx context.Context, commitID string) ([]string, error) {
	return pc.gitRepo.ListFilesInCommit(ctx, commitID)
}

// todo: should it be private ?
func (pc *ProxyCache) InvalidatePath(path string) error {
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

func (pc *ProxyCache) PullRefresh(ctx context.Context) error {
	if err := pc.gitRepo.Fetch(ctx); err != nil {
		return fmt.Errorf("git fetch error: %w", err)
	}

	freshCommits, err := pc.gitRepo.ListFreshCommits(ctx)
	if err != nil {
		return fmt.Errorf("git list fresh commits error: %w", err)
	}

	for _, fc := range freshCommits {
		commitFiles, err := pc.gitRepo.ListFilesInCommit(ctx, fc.CommitID)
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

	if len(freshCommits) > 0 {
		if err := pc.invalidateMainBranchCommits(); err != nil {
			return fmt.Errorf("error while invalidating main branch commits: %w", err)
		}
	}

	return nil
}

// ListAncestorCommits function is a cache proxy wrapper to git.Repo.
func (pc *ProxyCache) ListAncestorCommits(ctx context.Context, commitID string) ([]git.CommitInfo, error) {
	key := commitID

	return proxycache.Get(
		ctx,
		pc.cacheDir,
		CategoryAncestorCommits,
		key,
		nil,
		func(ctx context.Context) ([]git.CommitInfo, error) {
			return pc.gitRepo.ListAncestorCommits(ctx, commitID)
		},
	)
}
