// Package git provides an interface to work with a git repository.
package git

//go:generate mockgen -typed -source=git.go -destination=../mocks/mock_git.go -package=mocks

import (
	"context"

	"go-kweb-lang/git/internal"
)

// CommitInfo represents commit details.
type CommitInfo struct {
	// CommitID is a commit hash
	CommitID string
	// DateTime is a commit timestamp
	DateTime string
	// Comment is a commit comment
	Comment string
}

// Repo is an interface with methods to update a repository and get information about changes.
type Repo interface {
	// Create performs a git clone using the given url.
	Create(ctx context.Context, url string) error

	// Checkout checks out the revision specified by the commitID parameter.
	Checkout(ctx context.Context, commitID string) error

	// ListMainBranchCommits lists all commits in the main branch.
	ListMainBranchCommits(ctx context.Context) ([]CommitInfo, error)

	// FileExists checks whether the file exists in a repository.
	FileExists(path string) (bool, error)

	// ListFiles lists all files under the given path (and all subdirectories).
	ListFiles(path string) ([]string, error)

	// FindFileLastCommit provides information about the last commit for the given file.
	FindFileLastCommit(ctx context.Context, path string) (CommitInfo, error)

	// FindFileCommitsAfter lists all commits that contain the given file
	// and that are newer than the commitIDFrom parameter.
	FindFileCommitsAfter(ctx context.Context, path string, commitIDFrom string) ([]CommitInfo, error)

	// ListMergePoints lists all commits that are merge points starting from the commitID parameter.
	ListMergePoints(ctx context.Context, commitID string) ([]CommitInfo, error)

	// Fetch runs git fetch.
	Fetch(ctx context.Context) error

	// ListFreshCommits lists all commits in the main branch that are at origin/main and are not merged yet.
	ListFreshCommits(ctx context.Context) ([]CommitInfo, error)

	// Pull runs git pull.
	Pull(ctx context.Context) error

	// ListFilesInCommit lists all files that are in the commit with the commitID parameter.
	ListFilesInCommit(ctx context.Context, commitID string) ([]string, error)

	// ListAncestorCommits lists ancestor commits starting from the commitID parameter.
	ListAncestorCommits(ctx context.Context, commitID string) ([]CommitInfo, error)

	// ListCommitParents lists the parent commits of the given merge commit ID.
	ListCommitParents(ctx context.Context, commitID string) ([]string, error)

	// ListFilesBetweenCommits lists all files present in the branch between the fork commit and the last branch commit.
	ListFilesBetweenCommits(ctx context.Context, forkCommitID, branchLastCommitID string) ([]string, error)
}

type NewRepoConfig struct {
	Runner CommandRunner
}

func NewRepo(path string, opts ...func(config *NewRepoConfig)) Repo {
	config := NewRepoConfig{
		Runner: &internal.StdCommandRunner{},
	}
	for _, opt := range opts {
		opt(&config)
	}

	return &localRepo{
		path:   path,
		runner: config.Runner,
	}
}
