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
	// Create method do git clone with given url.
	Create(ctx context.Context, url string) error

	// Checkout checks out the revision specified by the commitID parameter.
	Checkout(ctx context.Context, commitID string) error

	// MainBranchCommits lists all commits in the main branch
	MainBranchCommits(ctx context.Context) ([]CommitInfo, error)

	// FileExists checks whether file exists in a repository.
	FileExists(path string) (bool, error)

	// ListFiles list all files for given path and all subdirectories.
	ListFiles(path string) ([]string, error)

	// FindFileLastCommit provides information about the last commit for given file.
	FindFileLastCommit(ctx context.Context, path string) (CommitInfo, error)

	// FindFileCommitsAfter lists all commits that contain the given file
	// and that are newer than then commitIDFrom parameter.
	FindFileCommitsAfter(ctx context.Context, path string, commitIDFrom string) ([]CommitInfo, error)

	// todo: it may require some refinement
	// FindMergePoints list all commits that are merge points starting from the commitID parameter.
	FindMergePoints(ctx context.Context, commitID string) ([]CommitInfo, error)

	// Fetch method do git fetch.
	Fetch(ctx context.Context) error

	// FreshCommits list all commits for the main branch that are at origin/main and are not merged yet.
	FreshCommits(ctx context.Context) ([]CommitInfo, error)

	// Pull method do git pull.
	Pull(ctx context.Context) error

	// FilesInCommit list all files that are in the commit with the commitID parameter.
	FilesInCommit(ctx context.Context, commitID string) ([]string, error)
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
