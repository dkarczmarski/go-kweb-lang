// Package githist provides information about commits in a git repository.
package githist

//go:generate mockgen -typed -source=githist.go -destination=../mocks/mock_githist.go -package=mocks

import (
	"context"
	"errors"
	"fmt"
	"log"

	"go-kweb-lang/git"
	"go-kweb-lang/proxycache"
)

const (
	categoryMainBranchCommits = "git-main-branch-commits"
)

type Invalidator interface {
	InvalidateFiles(paths []string) error
}

type GitHist struct {
	gitRepo     git.Repo
	cacheDir    string
	invalidator Invalidator
}

func New(gitRepo git.Repo, cacheDir string) *GitHist {
	return &GitHist{
		gitRepo:  gitRepo,
		cacheDir: cacheDir,
	}
}

func (gh *GitHist) RegisterInvalidator(invalidator Invalidator) {
	gh.invalidator = invalidator
}

// FindForkCommit returns the fork commit
// (on the main branch) for the given commitID. The commitID can point to a commit on the main branch,
// or to a commit from another branch.
// The result is never invalidated.
func (gh *GitHist) FindForkCommit(ctx context.Context, commitID string) (*git.CommitInfo, error) {
	mainBranchCommits, err := gh.listMainBranchCommits(ctx)
	if err != nil {
		return nil, err
	}

	return findCommitFunc(ctx, mainBranchCommits, commitID, gh.gitRepo.ListAncestorCommits)
}

// FindMergeCommit returns the merge commit
// with the main branch for the given commitID. The commitID can point to a commit on the main branch,
// or to a commit from another branch.
// When the result is not nil, the cache never needs to be invalidated.
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
		gh.cacheDir,
		categoryMainBranchCommits,
		"",
		nil,
		func(ctx context.Context) ([]git.CommitInfo, error) {
			return gh.gitRepo.ListMainBranchCommits(ctx)
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
func (gh *GitHist) PullRefresh(ctx context.Context) error {
	if err := gh.gitRepo.Fetch(ctx); err != nil {
		return fmt.Errorf("git fetch error: %w", err)
	}

	freshCommits, err := gh.gitRepo.ListFreshCommits(ctx)
	if err != nil {
		return fmt.Errorf("git list fresh commits error: %w", err)
	}

	if err := gh.gitRepo.Pull(ctx); err != nil {
		return fmt.Errorf("git pull error: %w", err)
	}

	var filesToInvalidate []string

	var mergeCommits []git.CommitInfo
	for _, fc := range freshCommits {
		commitFiles, err := gh.gitRepo.ListFilesInCommit(ctx, fc.CommitID)
		if err != nil {
			return fmt.Errorf("git list files of commit %s error: %w", fc.CommitID, err)
		}

		if len(commitFiles) == 0 {
			// it is a merge commit
			mergeCommits = append(mergeCommits, fc)
		}

		filesToInvalidate = append(filesToInvalidate, commitFiles...)
	}

	for _, mc := range mergeCommits {
		files, err := gh.MergeCommitFiles(ctx, mc.CommitID)
		if err != nil {
			return fmt.Errorf("failed to list files of the merge commit %v: %w", mc.CommitID, err)
		}

		filesToInvalidate = append(filesToInvalidate, files...)
	}

	if len(freshCommits) > 0 {
		if err := gh.invalidateMainBranchCommits(); err != nil {
			return fmt.Errorf("error while invalidating main branch commits: %w", err)
		}
	}

	filesToInvalidate = removeDuplicates(filesToInvalidate)
	if gh.invalidator != nil {
		if err := gh.invalidator.InvalidateFiles(filesToInvalidate); err != nil {
			return fmt.Errorf("invalidate paths error: %w", err)
		}
	}

	return nil
}

func removeDuplicates(list []string) []string {
	seen := make(map[string]bool)
	var result []string

	for _, item := range list {
		if !seen[item] {
			seen[item] = true
			result = append(result, item)
		}
	}

	return result
}

func (gh *GitHist) invalidateMainBranchCommits() error {
	return proxycache.InvalidateKey(gh.cacheDir, categoryMainBranchCommits, "")
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
		return nil, fmt.Errorf("failed to list parents of merge commit %v: %w", mergeCommitID, err)
	}

	if len(parents) == 1 {
		// it is not a merge commit
		return []string{}, nil
	} else if len(parents) != 2 {
		// todo: check it. are there cases with a merge commit of 3 parents
		return nil, errors.New("merge commit should have two parents")
	}

	var branchParentCommitID string
	for i := 0; i < 2; i++ {
		parent := parents[i]
		isMain, err := gh.IsMainBranchCommit(ctx, parent)
		if err != nil {
			return nil, fmt.Errorf("error while checking if the commit %v is on the main branch: %w",
				parent, err)
		}
		if !isMain {
			if len(branchParentCommitID) > 0 {
				return nil, errors.New("more than one parent is on the main branch. it should be impossible")
			}

			branchParentCommitID = parent
		}
	}
	if len(branchParentCommitID) == 0 {
		return nil, errors.New("no parent is on the main branch. it should be impossible")
	}

	forkCommit, err := gh.FindForkCommit(ctx, branchParentCommitID)
	if err != nil {
		return nil, fmt.Errorf("failed to find fork commit for %v: %w",
			branchParentCommitID, err)
	}

	files, err := gh.gitRepo.ListFilesBetweenCommits(ctx, forkCommit.CommitID, branchParentCommitID)
	if err != nil {
		return nil, fmt.Errorf("failed to list files between commits %v and %v: %w",
			forkCommit.CommitID, branchParentCommitID, err)
	}

	return files, nil
}
