// Package githist provides information about commits in a git repository.
package githist

//go:generate mockgen -typed -source=githist.go -destination=./internal/mocks/mocks.go -package=mocks

import (
	"context"
	"fmt"
	"log"
	"strings"

	"github.com/dkarczmarski/go-kweb-lang/git"
)

const bucketMainBranchCommits = "git-main-branch-commits"

// GitRepo is an interface used to decouple this package from the concrete git implementation.
type GitRepo interface {
	ListMainBranchCommits(ctx context.Context) ([]git.CommitInfo, error)
	ListMergePoints(ctx context.Context, commitID string) ([]git.CommitInfo, error)
	Fetch(ctx context.Context) error
	ListFreshCommits(ctx context.Context) ([]git.CommitInfo, error)
	Pull(ctx context.Context) error
	ListFilesInCommit(ctx context.Context, commitID string) ([]string, error)
	ListAncestorCommits(ctx context.Context, commitID string) ([]git.CommitInfo, error)
	ListCommitParents(ctx context.Context, commitID string) ([]string, error)
	ListFilesBetweenCommits(
		ctx context.Context,
		forkCommitID, branchLastCommitID string,
	) ([]string, error)
}

// CacheStorage is an interface used to decouple this package from the concrete store implementation.
type CacheStorage interface {
	Read(bucket, key string, buff any) (bool, error)
	Write(bucket, key string, data any) error
	Delete(bucket, key string) error
}

// Invalidator invalidates one concrete gitseek cache entry.
type Invalidator interface {
	InvalidateFile(langCode, path string) error
}

type GitHist struct {
	gitRepo GitRepo
	cache   CacheStorage
}

func New(gitRepo GitRepo, cache CacheStorage) *GitHist {
	return &GitHist{
		gitRepo: gitRepo,
		cache:   cache,
	}
}

// FindForkCommit returns the fork commit
// (on the main branch) for the given commitID.
// If the commitID itself exists on the main branch, nil is returned.
func (gh *GitHist) FindForkCommit(ctx context.Context, commitID string) (*git.CommitInfo, error) {
	mainBranchCommits, err := gh.listMainBranchCommits(ctx)
	if err != nil {
		return nil, err
	}

	return gh.findForkCommit(ctx, commitID, mainBranchCommits)
}

func (gh *GitHist) findForkCommit(
	ctx context.Context,
	commitID string,
	mainBranchCommits []git.CommitInfo,
) (*git.CommitInfo, error) {
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
	bucket := bucketMainBranchCommits
	key := ""

	var cached []git.CommitInfo

	exists, err := gh.cache.Read(bucket, key, &cached)
	if err != nil {
		return nil, fmt.Errorf("read main branch commits from cache: %w", err)
	}

	if exists {
		return cached, nil
	}

	mainBranchCommits, err := gh.gitRepo.ListMainBranchCommits(ctx)
	if err != nil {
		return nil, fmt.Errorf("list main branch commits: %w", err)
	}

	if err := gh.cache.Write(bucket, key, mainBranchCommits); err != nil {
		return nil, fmt.Errorf("write main branch commits to cache: %w", err)
	}

	return mainBranchCommits, nil
}

// findCommitFunc invokes listFunc with the given commitID and returns
// the first item of the list returned by listFunc that exists on the main branch.
// If the commitID itself exists on the main branch, nil is returned.
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

	for i := range commits {
		commit := commits[i]
		if containsCommit(mainBranchCommits, commit.CommitID) {
			clonedCommitInfo := cloneCommitInfo(commit)

			return &clonedCommitInfo
		}
	}

	// todo: move it
	log.Fatal("unexpected state: this should never happen")

	return nil
}

// PullRefresh performs a git fetch to retrieve fresh data, detects any changes, runs git pull
// and returns the list of changed files.
func (gh *GitHist) PullRefresh(ctx context.Context) ([]string, error) {
	mainBranchCommits, err := gh.listMainBranchCommits(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to list main branch commits: %w", err)
	}

	lastMainCommit := gh.getLastMainBranchCommit(mainBranchCommits)

	log.Printf("[githist] the last main branch commit is: %v", lastMainCommit)

	if err := gh.gitRepo.Fetch(ctx); err != nil {
		return nil, fmt.Errorf("git fetch error: %w", err)
	}

	freshCommits, err := gh.gitRepo.ListFreshCommits(ctx)
	if err != nil {
		return nil, fmt.Errorf("git list fresh commits error: %w", err)
	}

	if len(freshCommits) > 0 {
		if err := gh.InvalidateMainBranchCommits(); err != nil {
			return nil, fmt.Errorf("error while invalidating main branch commits: %w", err)
		}
	}

	freshMainBranchCommits := make([]git.CommitInfo, 0, len(freshCommits)+len(mainBranchCommits))
	freshMainBranchCommits = append(freshMainBranchCommits, freshCommits...)
	freshMainBranchCommits = append(freshMainBranchCommits, mainBranchCommits...)

	changedFiles, err := gh.processFreshCommits(ctx, freshCommits, freshMainBranchCommits)
	if err != nil {
		return nil, err
	}

	if err := gh.gitRepo.Pull(ctx); err != nil {
		return nil, fmt.Errorf("git pull error: %w", err)
	}

	return changedFiles, nil
}

func (gh *GitHist) InvalidateMainBranchCommits() error {
	if err := gh.cache.Delete(bucketMainBranchCommits, ""); err != nil {
		return fmt.Errorf("delete main branch commits cache: %w", err)
	}

	return nil
}

func (gh *GitHist) processFreshCommits(
	ctx context.Context,
	freshCommits []git.CommitInfo,
	freshMainBranchCommits []git.CommitInfo,
) ([]string, error) {
	changedFiles := make([]string, 0)

	for idx := range freshCommits {
		freshCommit := freshCommits[len(freshCommits)-1-idx]

		log.Printf(
			"[githist][%d/%d] process fresh commit: %s",
			idx+1,
			len(freshCommits),
			&freshCommit,
		)

		commitFiles, err := gh.gitRepo.ListFilesInCommit(ctx, freshCommit.CommitID)
		if err != nil {
			return nil, fmt.Errorf("list files of commit %s error: %w", freshCommit.CommitID, err)
		}

		if len(commitFiles) == 0 {
			// it might be a merge commit
			mergeCommitFiles, err := gh.mergeCommitFiles(
				ctx,
				freshCommit.CommitID,
				freshMainBranchCommits,
			)
			if err != nil {
				return nil, fmt.Errorf(
					"failed to list files of the merge commit %s: %w",
					freshCommit.CommitID,
					err,
				)
			}

			changedFiles = append(changedFiles, mergeCommitFiles...)

			log.Printf(
				"[githist][%d/%d] files in the merge commit: %s",
				idx+1,
				len(freshCommits),
				mergeCommitFiles,
			)
		} else {
			changedFiles = append(changedFiles, commitFiles...)

			log.Printf(
				"[githist][%d/%d] files in the commit: %s",
				idx+1,
				len(freshCommits),
				commitFiles,
			)
		}
	}

	return changedFiles, nil
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
	mainBranchCommits, err := gh.listMainBranchCommits(ctx)
	if err != nil {
		return nil, err
	}

	return gh.mergeCommitFiles(ctx, mergeCommitID, mainBranchCommits)
}

func (gh *GitHist) mergeCommitFiles(
	ctx context.Context,
	mergeCommitID string,
	mainBranchCommits []git.CommitInfo,
) ([]string, error) {
	parents, err := gh.gitRepo.ListCommitParents(ctx, mergeCommitID)
	if err != nil {
		return nil, fmt.Errorf("failed to list parents of merge commit %s: %w", mergeCommitID, err)
	}

	if len(parents) == 1 {
		// it is not a merge commit
		return []string{}, nil
	}

	var files []string

	for i := range parents {
		branchParentCommitID := parents[i]

		forkCommit, err := gh.findForkCommit(ctx, branchParentCommitID, mainBranchCommits)
		if err != nil {
			return nil, fmt.Errorf("failed to find fork commit for %s: %w", branchParentCommitID, err)
		}

		if forkCommit == nil {
			// is on the main branch
			continue
		}

		filesBetweenCommits, err := gh.gitRepo.ListFilesBetweenCommits(
			ctx,
			forkCommit.CommitID,
			branchParentCommitID,
		)
		if err != nil {
			return nil, fmt.Errorf(
				"failed to list files between commits %s and %s: %w",
				forkCommit.CommitID,
				branchParentCommitID,
				err,
			)
		}

		if len(filesBetweenCommits) > 0 {
			files = append(files, filesBetweenCommits...)
		}
	}

	return files, nil
}

func (gh *GitHist) GetLastMainBranchCommit(ctx context.Context) (git.CommitInfo, error) {
	commits, err := gh.listMainBranchCommits(ctx)
	if err != nil {
		return zeroCommitInfo(), err
	}

	return gh.getLastMainBranchCommit(commits), nil
}

func (gh *GitHist) getLastMainBranchCommit(mainBranchCommits []git.CommitInfo) git.CommitInfo {
	if len(mainBranchCommits) == 0 {
		return zeroCommitInfo()
	}

	return cloneCommitInfo(mainBranchCommits[0])
}

// cloneCommitInfo clones a single commit info to avoid retaining references to large underlying data
// and prevent memory leaks.
func cloneCommitInfo(commit git.CommitInfo) git.CommitInfo {
	return git.CommitInfo{
		CommitID: strings.Clone(commit.CommitID),
		DateTime: strings.Clone(commit.DateTime),
		Comment:  strings.Clone(commit.Comment),
	}
}

func zeroCommitInfo() git.CommitInfo {
	return git.CommitInfo{
		CommitID: "",
		DateTime: "",
		Comment:  "",
	}
}
