// Package git provides an interface to work with a git repository.
package git

//go:generate mockgen -typed -source=git.go -destination=../mocks/mock_git.go -package=mocks

import (
	"context"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"

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

type CommandRunner interface {
	Exec(ctx context.Context, workingDir string, cmd string, args ...string) (string, error)
}

type localRepo struct {
	path   string
	runner CommandRunner
}

func (lr *localRepo) Create(ctx context.Context, url string) error {
	_, err := lr.runner.Exec(ctx, lr.path,
		"git",
		"clone",
		url,
		".",
	)
	if err != nil {
		return fmt.Errorf("git command failed: %w", err)
	}

	return nil
}

func (lr *localRepo) Checkout(ctx context.Context, commitID string) error {
	_, err := lr.runner.Exec(ctx, lr.path,
		"git",
		"checkout",
		commitID,
	)
	if err != nil {
		return fmt.Errorf("git command ( %v ) failed: %w",
			fmt.Sprintf("git checkout %v", commitID), err)
	}

	return nil
}

func (lr *localRepo) MainBranchCommits(ctx context.Context) ([]CommitInfo, error) {
	out, err := lr.runner.Exec(ctx, lr.path,
		"git",
		"--no-pager",
		"log",
		"main",
		"--pretty=format:%H %cd %s",
		"--date=iso-strict",
		"--first-parent",
	)
	if err != nil {
		return nil, fmt.Errorf("git command ( %v ) failed: %w",
			"git --no-pager log main --pretty=format:\"%H %cd %s\" --date=iso-strict --first-parent", err)
	}

	if len(strings.TrimSpace(out)) == 0 {
		return nil, nil
	}

	lines := outputToLines(out)
	commits := make([]CommitInfo, 0, len(lines))

	for _, line := range lines {
		commits = append(commits, lineToCommitInfo(line))
	}

	return commits, nil
}

func (lr *localRepo) FileExists(path string) (bool, error) {
	_, err := os.Stat(lr.path + "/" + path)
	if os.IsNotExist(err) {
		return false, nil
	}
	if err != nil {
		return false, fmt.Errorf("failed to stat file: %w", err)
	}

	return true, nil
}

func (lr *localRepo) ListFiles(path string) ([]string, error) {
	var files []string

	fullPath := filepath.Join(lr.path, path)
	err := filepath.WalkDir(fullPath, func(path string, d os.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if !d.IsDir() {
			relPath, err := filepath.Rel(fullPath, path)
			if err != nil {
				return fmt.Errorf("failed to compute relative path: %w", err)
			}

			if strings.HasPrefix(relPath, ".git/") {
				return nil
			}

			files = append(files, relPath)
		}

		return nil
	})
	if err != nil {
		return nil, fmt.Errorf("failed to walk directory: %w", err)
	}

	return files, nil
}

func (lr *localRepo) FindFileLastCommit(ctx context.Context, path string) (CommitInfo, error) {
	out, err := lr.runner.Exec(ctx, lr.path,
		"git",
		"log",
		"-1",
		"--format=%H %cd %s",
		"--date=iso-strict",
		"--",
		path,
	)
	if err != nil {
		return CommitInfo{}, fmt.Errorf("git command failed: %w", err)
	}

	// todo: probably it should be some "not-found" error
	if len(strings.TrimSpace(out)) == 0 {
		return CommitInfo{}, nil
	}

	return lineToCommitInfo(out), nil
}

func (lr *localRepo) FindFileCommitsAfter(ctx context.Context, path string, commitIDFrom string) ([]CommitInfo, error) {
	out, err := lr.runner.Exec(ctx, lr.path,
		"git",
		"log",
		"--pretty=format:%H %cd %s",
		"--date=iso-strict",
		commitIDFrom+"..",
		"--",
		path,
	)
	if err != nil {
		return nil, fmt.Errorf("git command failed: %w", err)
	}

	if len(strings.TrimSpace(out)) == 0 {
		return nil, nil
	}

	lines := outputToLines(out)
	commits := make([]CommitInfo, 0, len(lines))

	for _, line := range lines {
		commits = append(commits, lineToCommitInfo(line))
	}

	return commits, nil
}

func (lr *localRepo) FindMergePoints(ctx context.Context, commitID string) ([]CommitInfo, error) {
	// todo: probably we can do it better, to list only necessary merging point to the main branch
	// todo: return result in reverse order?
	out, err := lr.runner.Exec(ctx, lr.path,
		"git",
		"--no-pager",
		"log",
		"--ancestry-path",
		"--merges",
		"--pretty=format:%H %cd %s",
		"--date=iso-strict",
		commitID+"..main",
	)
	if err != nil {
		return nil, fmt.Errorf("git command failed: %w", err)
	}
	if len(strings.TrimSpace(out)) == 0 {
		return nil, nil
	}

	lines := outputToLines(out)
	commits := make([]CommitInfo, 0, len(lines))

	for _, line := range lines {
		commits = append(commits, lineToCommitInfo(line))
	}

	return commits, nil
}

func (lr *localRepo) Fetch(ctx context.Context) error {
	_, err := lr.runner.Exec(ctx, lr.path,
		"git",
		"fetch",
	)
	if err != nil {
		return fmt.Errorf("git command failed: %w", err)
	}

	return nil
}

func (lr *localRepo) FreshCommits(ctx context.Context) ([]CommitInfo, error) {
	out, err := lr.runner.Exec(ctx, lr.path,
		"git",
		"--no-pager",
		"log",
		"--pretty=format:%H %cd %s",
		"--date=iso-strict",
		"main..origin/main",
	)
	if err != nil {
		return nil, fmt.Errorf("git command failed: %w", err)
	}

	if len(strings.TrimSpace(out)) == 0 {
		return nil, nil
	}

	lines := outputToLines(out)
	commits := make([]CommitInfo, 0, len(lines))

	for _, line := range lines {
		commits = append(commits, lineToCommitInfo(line))
	}

	return commits, nil
}

func (lr *localRepo) Pull(ctx context.Context) error {
	_, err := lr.runner.Exec(ctx, lr.path,
		"git",
		"pull",
	)
	if err != nil {
		return fmt.Errorf("git command failed: %w", err)
	}

	return nil
}

func (lr *localRepo) FilesInCommit(ctx context.Context, commitID string) ([]string, error) {
	out, err := lr.runner.Exec(ctx, lr.path,
		"git",
		"diff-tree",
		"--no-commit-id",
		"--name-only",
		"-r",
		commitID,
	)
	if err != nil {
		return nil, fmt.Errorf("git command failed: %w", err)
	}

	if len(strings.TrimSpace(out)) == 0 {
		return nil, nil
	}

	return outputToLines(out), nil
}

func outputToLines(out string) []string {
	return strings.Split(strings.TrimSuffix(out, "\n"), "\n")
}

func lineToCommitInfo(line string) CommitInfo {
	segs := strings.SplitN(line, " ", 3)
	if len(segs) != 3 {
		log.Fatalf("line syntax error: %s", line)
	}

	return CommitInfo{
		CommitID: segs[0],
		DateTime: segs[1],
		Comment:  strings.TrimRight(segs[2], "\n"),
	}
}
