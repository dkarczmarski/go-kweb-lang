package git

import (
	"fmt"
	"go-kweb-lang/git/internal"
	"log"
	"os"
	"strings"
)

type CommitInfo struct {
	CommitId string
	DateTime string
	Comment  string
}

type Repo interface {
	FileExists(path string) (bool, error)
	FindFileLastCommit(path string) (CommitInfo, error)
	FindFileCommitsAfter(path string, commitIdFrom string) ([]CommitInfo, error)
	FindMergePoints(commitId string) ([]CommitInfo, error)
}

func NewRepo(path string) Repo {
	return &LocalRepo{
		path:   path,
		runner: &internal.StdCommandRunner{},
	}
}

type CommandRunner interface {
	Exec(workingDir string, cmd string, args ...string) (string, error)
}

type LocalRepo struct {
	path   string
	runner CommandRunner
}

func (lr *LocalRepo) FileExists(path string) (bool, error) {
	_, err := os.Stat(lr.path + "/" + path)
	if os.IsNotExist(err) {
		return false, nil
	}
	if err != nil {
		return false, err
	}
	return true, nil
}

func (lr *LocalRepo) FindFileLastCommit(path string) (CommitInfo, error) {
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

func (lr *LocalRepo) FindFileCommitsAfter(path string, commitIdFrom string) ([]CommitInfo, error) {
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
		return []CommitInfo{}, nil
	}

	var commits []CommitInfo

	lines := strings.Split(out, "\n")
	for _, line := range lines {
		commits = append(commits, lineToCommitInfo(line))
	}

	return commits, nil
}

func (lr *LocalRepo) FindMergePoints(commitId string) ([]CommitInfo, error) {
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
		return []CommitInfo{}, nil
	}

	var commits []CommitInfo

	lines := strings.Split(out, "\n")
	for _, line := range lines {
		commits = append(commits, lineToCommitInfo(line))
	}

	return commits, nil
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
