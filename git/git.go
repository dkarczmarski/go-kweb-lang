package git

//go:generate mockgen -typed -source=git.go -destination=../mocks/mock_git.go -package=mocks

import (
	"fmt"
	"go-kweb-lang/git/internal"
	"log"
	"os"
	"strings"
)

// CommitInfo represents commit details
type CommitInfo struct {
	// CommitId is a commit hash
	CommitId string
	// DateTime is a commit timestamp
	DateTime string
	// Comment is a commit comment
	Comment string
}

type Repo interface {
	FileExists(path string) (bool, error)
	FindFileLastCommit(path string) (CommitInfo, error)
	FindFileCommitsAfter(path string, commitIdFrom string) ([]CommitInfo, error)
	FindMergePoints(commitId string) ([]CommitInfo, error)
	Fetch() error
	FreshCommits() ([]CommitInfo, error)
	Pull() error
	CommitFiles(commitId string) ([]string, error)
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
	Exec(workingDir string, cmd string, args ...string) (string, error)
}

type localRepo struct {
	path   string
	runner CommandRunner
}

func (lr *localRepo) FileExists(path string) (bool, error) {
	_, err := os.Stat(lr.path + "/" + path)
	if os.IsNotExist(err) {
		return false, nil
	}
	if err != nil {
		return false, err
	}
	return true, nil
}

func (lr *localRepo) FindFileLastCommit(path string) (CommitInfo, error) {
	out, err := lr.runner.Exec(lr.path,
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
	if len(strings.TrimSpace(out)) == 0 {
		return CommitInfo{}, nil
	}

	return lineToCommitInfo(out), nil
}

func (lr *localRepo) FindFileCommitsAfter(path string, commitIdFrom string) ([]CommitInfo, error) {
	out, err := lr.runner.Exec(lr.path,
		"git",
		"log",
		"--pretty=format:%H %cd %s",
		"--date=iso-strict",
		commitIdFrom+"..",
		"--",
		path,
	)
	if err != nil {
		return nil, fmt.Errorf("git command failed: %w", err)
	}
	if len(strings.TrimSpace(out)) == 0 {
		return nil, nil
	}

	var commits []CommitInfo

	lines := outputToLines(out)
	for _, line := range lines {
		commits = append(commits, lineToCommitInfo(line))
	}

	return commits, nil
}

func (lr *localRepo) FindMergePoints(commitId string) ([]CommitInfo, error) {
	// todo: probably we can do it better, to list only necessary merging point to the main branch
	// todo: return result in reverse order?
	out, err := lr.runner.Exec(lr.path,
		"git",
		"--no-pager",
		"log",
		"--ancestry-path",
		"--merges",
		"--pretty=format:%H %cd %s",
		"--date=iso-strict",
		commitId+"..main",
	)
	if err != nil {
		return nil, fmt.Errorf("git command failed: %w", err)
	}
	if len(strings.TrimSpace(out)) == 0 {
		return nil, nil
	}

	var commits []CommitInfo

	lines := outputToLines(out)
	for _, line := range lines {
		commits = append(commits, lineToCommitInfo(line))
	}

	return commits, nil
}

func (lr *localRepo) Fetch() error {
	_, err := lr.runner.Exec(lr.path,
		"git",
		"fetch",
	)
	if err != nil {
		return fmt.Errorf("git command failed: %w", err)
	}

	return nil
}

func (lr *localRepo) FreshCommits() ([]CommitInfo, error) {
	out, err := lr.runner.Exec(lr.path,
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

	var commits []CommitInfo

	lines := outputToLines(out)
	for _, line := range lines {
		commits = append(commits, lineToCommitInfo(line))
	}

	return commits, nil
}

func (lr *localRepo) Pull() error {
	_, err := lr.runner.Exec(lr.path,
		"git",
		"pull",
	)
	if err != nil {
		return fmt.Errorf("git command failed: %w", err)
	}

	return nil
}

func (lr *localRepo) CommitFiles(commitId string) ([]string, error) {
	out, err := lr.runner.Exec(lr.path,
		"git",
		"diff-tree",
		"--no-commit-id",
		"--name-only",
		"-r",
		commitId,
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
		CommitId: segs[0],
		DateTime: segs[1],
		Comment:  strings.TrimRight(segs[2], "\n"),
	}
}
