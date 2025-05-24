// Package gitpc is a proxy cache for git repository.
// The name stands for git proxy cache.
package gitpc

import (
	"context"
	"fmt"
	"log"

	"go-kweb-lang/git"
	"go-kweb-lang/proxycache"
)

const (
	categoryLastCommit        = "git-file-last-commit"
	categoryUpdates           = "git-file-updates"
	categoryForkCommit        = "git-fork-commit"
	categoryMergeCommit       = "git-merge-commit"
	categoryMainBranchCommits = "git-main-branch-commits"
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

// FindFileLastCommit function is a cache proxy wrapper to git.Repo.
// The cached result should be invalidated when a new commit occurs for the given path.
// The result should be invalidated when the given path exists in at least one commit in git pull.
func (pc *ProxyCache) FindFileLastCommit(ctx context.Context, path string) (git.CommitInfo, error) {
	key := path

	return proxycache.Get(
		ctx,
		pc.cacheDir,
		categoryLastCommit,
		key,
		nil,
		func(ctx context.Context) (git.CommitInfo, error) {
			return pc.gitRepo.FindFileLastCommit(ctx, path)
		},
	)
}

// FindFileCommitsAfter function is a cache proxy wrapper to git.Repo.
// The cached result should be invalidated when a new commit occurs for the given path.
func (pc *ProxyCache) FindFileCommitsAfter(ctx context.Context, path string, commitIDFrom string) ([]git.CommitInfo, error) {
	key := path

	return proxycache.Get(
		ctx,
		pc.cacheDir,
		categoryUpdates,
		key,
		nil,
		func(ctx context.Context) ([]git.CommitInfo, error) {
			return pc.gitRepo.FindFileCommitsAfter(ctx, path, commitIDFrom)
		},
	)
}

// FindForkCommit is a cache proxy wrapper around a function that returns the fork commit
// (on the main branch) for the given commitID. The commitID can point to a commit on the main branch,
// or to a commit from another branch.
// The result is never invalidated.
func (pc *ProxyCache) FindForkCommit(ctx context.Context, commitID string) (*git.CommitInfo, error) {
	key := commitID

	return proxycache.Get(
		ctx,
		pc.cacheDir,
		categoryForkCommit,
		key,
		nil,
		func(ctx context.Context) (*git.CommitInfo, error) {
			mainBranchCommits, err := pc.listMainBranchCommits(ctx)
			if err != nil {
				return nil, err
			}

			return findCommitFunc(ctx, mainBranchCommits, commitID, pc.gitRepo.ListAncestorCommits)
		},
	)
}

// FindMergeCommit is a cache proxy wrapper around a function that returns the merge commit
// with the main branch for the given commitID. The commitID can point to a commit on the main branch,
// or to a commit from another branch.
// When the result is not nil, the cache never needs to be invalidated.
func (pc *ProxyCache) FindMergeCommit(ctx context.Context, commitID string) (*git.CommitInfo, error) {
	key := commitID

	return proxycache.Get(
		ctx,
		pc.cacheDir,
		categoryMergeCommit,
		key,
		func(commitInfo *git.CommitInfo) bool {
			return commitInfo == nil
		},
		func(ctx context.Context) (*git.CommitInfo, error) {
			mainBranchCommits, err := pc.listMainBranchCommits(ctx)
			if err != nil {
				return nil, err
			}

			return findCommitFunc(ctx, mainBranchCommits, commitID, pc.gitRepo.ListMergePoints)
		},
	)
}

func (pc *ProxyCache) listMainBranchCommits(ctx context.Context) ([]git.CommitInfo, error) {
	return proxycache.Get(
		ctx,
		pc.cacheDir,
		categoryMainBranchCommits,
		"",
		nil,
		func(ctx context.Context) ([]git.CommitInfo, error) {
			return pc.gitRepo.ListMainBranchCommits(ctx)
		},
	)
}

func findCommitFunc(
	ctx context.Context,
	mainBranchCommits []git.CommitInfo,
	commitID string,
	listFunc func(ctx context.Context, commitID string) ([]git.CommitInfo, error),
) (*git.CommitInfo, error) {
	var commitInfo *git.CommitInfo

	if !containsCommit(mainBranchCommits, commitID) {
		commits, err := listFunc(ctx, commitID)
		if err != nil {
			return nil, err
		}

		commitInfo = findFirstCommit(mainBranchCommits, commits)
	}

	return commitInfo, nil
}

func containsCommit(list []git.CommitInfo, commitID string) bool {
	for i := range list {
		if list[i].CommitID == commitID {
			return true
		}
	}

	return false
}

func findFirstCommit(mainBranchCommits []git.CommitInfo, commits []git.CommitInfo) *git.CommitInfo {
	commitsLen := len(commits)
	if commitsLen == 0 {
		return nil
	}

	for i := 0; i < commitsLen; i++ {
		commit := commits[i]
		if containsCommit(mainBranchCommits, commit.CommitID) {
			return &commit
		}
	}

	// todo: move it
	log.Fatal("unexpected state: this should never happen")

	return nil
}

// PullRefresh performs a git fetch to retrieve fresh data, detects any changes,
// invalidates relevant cache keys, and finally runs git pull.
func (pc *ProxyCache) PullRefresh(ctx context.Context) error {
	if err := pc.gitRepo.Fetch(ctx); err != nil {
		return fmt.Errorf("git fetch error: %w", err)
	}

	freshCommits, err := pc.gitRepo.ListFreshCommits(ctx)
	if err != nil {
		return fmt.Errorf("git list fresh commits error: %w", err)
	}

	var mergeCommits []git.CommitInfo
	for _, fc := range freshCommits {
		commitFiles, err := pc.gitRepo.ListFilesInCommit(ctx, fc.CommitID)
		if err != nil {
			return fmt.Errorf("git list files of commit %s error: %w", fc.CommitID, err)
		}

		if len(commitFiles) == 0 {
			// it is a merge commit
			mergeCommits = append(mergeCommits, fc)
		}

		for _, f := range commitFiles {
			if err := pc.invalidatePath(f); err != nil {
				return fmt.Errorf("git cache invalidate path %s error: %w", f, err)
			}
		}
	}

	for _, mc := range mergeCommits {
		// todo: handle files from merge commit branch
		_ = mc
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

func (pc *ProxyCache) invalidatePath(path string) error {
	for _, category := range []string{
		categoryLastCommit,
		categoryUpdates,
	} {
		key := path
		if err := proxycache.InvalidateKey(pc.cacheDir, category, key); err != nil {
			return fmt.Errorf("error while invalidataing cache key %v: %w", key, err)
		}
	}

	return nil
}

func (pc *ProxyCache) invalidateMainBranchCommits() error {
	return proxycache.InvalidateKey(pc.cacheDir, categoryMainBranchCommits, "")
}

// IsMainBranchCommit checks whether the given commit ID is part of the main branch.
func (pc *ProxyCache) IsMainBranchCommit(ctx context.Context, commitID string) (bool, error) {
	mainBranchCommits, err := pc.listMainBranchCommits(ctx)
	if err != nil {
		return false, err
	}

	return containsCommit(mainBranchCommits, commitID), nil
}
