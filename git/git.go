package git

import (
	"bytes"
	"log"
	"os"
	"os/exec"
	"strings"
)

type CommitInfo struct {
	CommitId string
	DateTime string
	Comment  string
}

type Repo interface {
	FileExists(path string) bool
	FindFileLastCommit(path string) CommitInfo
	FindFileCommitsAfter(path string, commitIdFrom string) []CommitInfo
	FindMergePoints(commitId string) []CommitInfo
}

func NewRepo(path string) Repo {
	return &LocalRepo{path: path}
}

type LocalRepo struct {
	path string
}

func (lr *LocalRepo) FileExists(path string) bool {
	_, err := os.Stat(lr.path + "/" + path)
	if os.IsNotExist(err) {
		return false
	}
	if err != nil {
		log.Fatal(err)
	}
	return true
}

func (lr *LocalRepo) FindFileLastCommit(path string) CommitInfo {
	out := runCommand(lr.path,
		"git",
		"log",
		"-1",
		"--format=%H %cd %s",
		"--date=iso-strict",
		"--",
		path,
	)
	if len(strings.TrimSpace(out)) == 0 {
		return CommitInfo{}
	}

	return lineToCommitInfo(out)
}

func (lr *LocalRepo) FindFileCommitsAfter(path string, commitIdFrom string) []CommitInfo {
	out := runCommand(lr.path,
		"git",
		"log",
		"--pretty=format:%H %cd %s",
		"--date=iso-strict",
		commitIdFrom+"..",
		"--",
		path,
	)
	if len(strings.TrimSpace(out)) == 0 {
		return []CommitInfo{}
	}

	var commits []CommitInfo

	lines := strings.Split(out, "\n")
	for _, line := range lines {
		commits = append(commits, lineToCommitInfo(line))
	}

	return commits
}

func (lr *LocalRepo) FindMergePoints(commitId string) []CommitInfo {
	out := runCommand(lr.path,
		"git",
		"--no-pager",
		"log",
		"--ancestry-path",
		"--merges",
		"--pretty=format:%H %cd %s",
		"--date=iso-strict",
		commitId+"..main",
	)
	if len(strings.TrimSpace(out)) == 0 {
		return []CommitInfo{}
	}

	var commits []CommitInfo

	lines := strings.Split(out, "\n")
	for _, line := range lines {
		commits = append(commits, lineToCommitInfo(line))
	}

	return commits
}
func runCommand(cwd string, command string, args ...string) string {
	cmd := exec.Command(command, args...)
	cmd.Dir = cwd

	var out bytes.Buffer
	var stderr bytes.Buffer
	cmd.Stdout = &out
	cmd.Stderr = &stderr

	err := cmd.Run()
	if err != nil {
		log.Fatal(err, stderr.String())
	}

	return out.String()
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
