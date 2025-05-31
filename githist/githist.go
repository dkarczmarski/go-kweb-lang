// Package githist provides information about commits in a git repository.
package githist

//go:generate mockgen -typed -source=githist.go -destination=./internal/mocks/mocks.go -package=mocks

import (
	"context"
	"fmt"
	"log"

	"go-kweb-lang/git"
	"go-kweb-lang/proxycache"
)

const (
	bucketMainBranchCommits = "git-main-branch-commits"
)

// GitRepo is an interface used to decouple this package from the concrete git implementation
type GitRepo interface {
	ListMainBranchCommits(ctx context.Context) ([]git.CommitInfo, error)

	ListMergePoints(ctx context.Context, commitID string) ([]git.CommitInfo, error)

	Fetch(ctx context.Context) error

	ListFreshCommits(ctx context.Context) ([]git.CommitInfo, error)

	Pull(ctx context.Context) error

	ListFilesInCommit(ctx context.Context, commitID string) ([]string, error)

	ListAncestorCommits(ctx context.Context, commitID string) ([]git.CommitInfo, error)

	ListCommitParents(ctx context.Context, commitID string) ([]string, error)

	ListFilesBetweenCommits(ctx context.Context, forkCommitID, branchLastCommitID string) ([]string, error)
}

// CacheStore is an interface used to decouple this package from the concrete store implementation
type CacheStore interface {
	Read(bucket, key string, buff any) (bool, error)
	Write(bucket, key string, data any) error
	Delete(bucket, key string) error
}

type Invalidator interface {
	InvalidateFile(path string) error
}

type GitHist struct {
	gitRepo     GitRepo
	cacheStore  CacheStore
	invalidator Invalidator
}

func New(gitRepo GitRepo, cacheStore CacheStore) *GitHist {
	return &GitHist{
		gitRepo:    gitRepo,
		cacheStore: cacheStore,
	}
}

func (gh *GitHist) RegisterInvalidator(invalidator Invalidator) {
	gh.invalidator = invalidator
}

// FindForkCommit returns the fork commit
// (on the main branch) for the given commitID.
// If the commitID itself exists on the main branch, nil is returned.
func (gh *GitHist) FindForkCommit(ctx context.Context, commitID string) (*git.CommitInfo, error) {
	mainBranchCommits, err := gh.listMainBranchCommits(ctx)
	if err != nil {
		return nil, err
	}

	return findCommitFunc(ctx, mainBranchCommits, commitID, gh.gitRepo.ListAncestorCommits)
}

// FindMergeCommit returns the merge commit
// with the main branch for the given commitID.
// If the commitID itself exists on the main branch, nil is returned.
func (gh *GitHist) FindMergeCommit(ctx context.Context, commitID string) (*git.CommitInfo, error) {
	mainBranchCommits, err := gh.listMainBranchCommits(ctx)
	if err != nil {
		return nil, err
	}

	return findCommitFunc(ctx, mainBranchCommits, commitID, gh.gitRepo.ListMergePoints)
}

func (gh *GitHist) listMainBranchCommits(ctx context.Context) ([]git.CommitInfo, error) {
	return proxycache.Get(
		ctx,
		gh.cacheStore,
		bucketMainBranchCommits,
		"",
		nil,
		func(ctx context.Context) ([]git.CommitInfo, error) {
			return gh.gitRepo.ListMainBranchCommits(ctx)
		},
	)
}

// findCommitFunc invokes listFunc with the given commitID and returns
// the first item of the list returned by listFunc that exists on the main branch. If the commitID itself exists
// on the main branch, nil is returned.
func findCommitFunc(
	ctx context.Context,
	mainBranchCommits []git.CommitInfo,
	commitID string,
	listFunc func(ctx context.Context, commitID string) ([]git.CommitInfo, error),
) (*git.CommitInfo, error) {
	if containsCommit(mainBranchCommits, commitID) {
		return nil, nil
	}

	commits, err := listFunc(ctx, commitID)
	if err != nil {
		return nil, err
	}

	return findFirstCommit(mainBranchCommits, commits), nil
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

// PullRefresh performs a git fetch to retrieve fresh data, detects any changes, runs git pull
// and invalidates changed files.
func (gh *GitHist) PullRefresh(ctx context.Context) error {
	if err := gh.gitRepo.Fetch(ctx); err != nil {
		return fmt.Errorf("git fetch error: %w", err)
	}

	freshCommits, err := gh.gitRepo.ListFreshCommits(ctx)
	if err != nil {
		return fmt.Errorf("git list fresh commits error: %w", err)
	}

	// it would be better if the 'pull' part is after the 'invalidation' part,
	// but the 'invalidation' step checks whether a commit is on the main branch
	// and that's why 'pull' must be executed first.
	if err := gh.gitRepo.Pull(ctx); err != nil {
		return fmt.Errorf("git pull error: %w", err)
	}

	if len(freshCommits) > 0 {
		if err := gh.invalidateMainBranchCommits(); err != nil {
			return fmt.Errorf("error while invalidating main branch commits: %w", err)
		}
	}

	invalidated := make(map[string]int)
	for i, fc := range freshCommits {
		log.Printf("[%d/%d] process fresh commit: %s", i, len(freshCommits), &fc)

		commitFiles, err := gh.gitRepo.ListFilesInCommit(ctx, fc.CommitID)
		if err != nil {
			return fmt.Errorf("list files of commit %s error: %w", fc.CommitID, err)
		}

		var filesToInvalidate []string
		if len(commitFiles) == 0 {
			// it might be a merge commit
			mergeCommitFiles, err := gh.MergeCommitFiles(ctx, fc.CommitID)
			if err != nil {
				return fmt.Errorf("failed to list files of the merge commit %s: %w", fc.CommitID, err)
			}

			filesToInvalidate = mergeCommitFiles

			log.Printf("[%d/%d] files in the merge commit: %s", i, len(freshCommits), mergeCommitFiles)
		} else {
			filesToInvalidate = commitFiles

			log.Printf("[%d/%d] files in the commit: %s", i, len(freshCommits), commitFiles)
		}

		for _, file := range filesToInvalidate {
			if invalidatedAt, isInvalidated := invalidated[file]; !isInvalidated {
				log.Printf("[%d/%d] invalidate file %s", i, len(freshCommits), file)

				if gh.invalidator != nil {
					if err := gh.invalidator.InvalidateFile(file); err != nil {
						return fmt.Errorf("invalidate file %s error: %w", file, err)
					}
				}

				invalidated[file] = i
			} else {
				log.Printf("[%d/%d] invalidate file %s - (skip) already done at %d",
					i, len(freshCommits), file, invalidatedAt)
			}
		}
	}

	return nil
}

func (gh *GitHist) invalidateMainBranchCommits() error {
	return gh.cacheStore.Delete(bucketMainBranchCommits, "")
}

// IsMainBranchCommit checks whether the given commit ID is part of the main branch.
func (gh *GitHist) IsMainBranchCommit(ctx context.Context, commitID string) (bool, error) {
	mainBranchCommits, err := gh.listMainBranchCommits(ctx)
	if err != nil {
		return false, err
	}

	return containsCommit(mainBranchCommits, commitID), nil
}

// MergeCommitFiles lists all files from the branch that was merged in the merge commit specified by mergeCommitID.
func (gh *GitHist) MergeCommitFiles(ctx context.Context, mergeCommitID string) ([]string, error) {
	parents, err := gh.gitRepo.ListCommitParents(ctx, mergeCommitID)
	if err != nil {
		return nil, fmt.Errorf("failed to list parents of merge commit %s: %w", mergeCommitID, err)
	}

	if len(parents) == 1 {
		// it is not a merge commit
		return []string{}, nil
	}

	var files []string
	for i := 0; i < len(parents); i++ {
		branchParentCommitID := parents[i]

		forkCommit, err := gh.FindForkCommit(ctx, branchParentCommitID)
		if err != nil {
			return nil, fmt.Errorf("failed to find fork commit for %s: %w",
				branchParentCommitID, err)
		}

		if forkCommit == nil {
			// is on the main branch
			continue
		}

		filesBetweenCommits, err := gh.gitRepo.ListFilesBetweenCommits(ctx, forkCommit.CommitID, branchParentCommitID)
		if err != nil {
			return nil, fmt.Errorf("failed to list files between commits %s and %s: %w",
				forkCommit.CommitID, branchParentCommitID, err)
		}

		if len(filesBetweenCommits) > 0 {
			files = append(files, filesBetweenCommits...)
		}
	}

	return files, nil
}
