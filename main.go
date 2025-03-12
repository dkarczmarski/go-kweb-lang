package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

type CommitInfo struct {
	CommitId string
	DateTime string
	Comment  string
}

type GitRepo interface {
	FileExists(path string) bool
	FindFileLastCommit(path string) CommitInfo
	FindFileCommitsAfter(path string, commitIdFrom string) []CommitInfo
	FindMergePoints(commitId string) []CommitInfo
}

type LocalGitRepo struct {
	path string
}

func (gr *LocalGitRepo) FileExists(path string) bool {
	_, err := os.Stat(gr.path + "/" + path)
	if os.IsNotExist(err) {
		return false
	}
	if err != nil {
		log.Fatal(err)
	}
	return true
}

func (gr *LocalGitRepo) FindFileLastCommit(path string) CommitInfo {
	out := runCommand(gr.path,
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

func (gr *LocalGitRepo) FindFileCommitsAfter(path string, commitIdFrom string) []CommitInfo {
	out := runCommand(gr.path,
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

func (gr *LocalGitRepo) FindMergePoints(commitId string) []CommitInfo {
	out := runCommand(gr.path,
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

type FileInfo struct {
	LangRelPath      string
	LangCommit       CommitInfo
	OriginFileStatus string
	OriginUpdates    []OriginUpdate
}

type OriginUpdate struct {
	Commit     CommitInfo
	MergePoint CommitInfo
}

type GitLangSeeker struct {
	gitRepo GitRepo
}

func repoOriginFilePath(relPath string) string {
	return "content/en/" + relPath
}

func repoLangFilePath(relPath string) string {
	return "content/pl/" + relPath
}

func (s *GitLangSeeker) checkFiles(langRelPaths []string) []FileInfo {
	fileInfoList := make([]FileInfo, len(langRelPaths))

	for i, langRelPath := range langRelPaths {
		var fileInfo FileInfo

		originFilePath := repoOriginFilePath(langRelPath)
		langFilePath := repoLangFilePath(langRelPath)

		fileInfo.LangRelPath = langRelPath
		langLastCommit := s.gitRepo.FindFileLastCommit(langFilePath)
		fileInfo.LangCommit = langLastCommit

		if !s.gitRepo.FileExists(originFilePath) {
			fileInfo.OriginFileStatus = "NOT_EXIST"
		}

		originCommitsAfter := s.gitRepo.FindFileCommitsAfter(originFilePath, langLastCommit.CommitId)
		if len(originCommitsAfter) > 0 {
			fileInfo.OriginFileStatus = "MODIFIED"
		}

		for _, originCommitAfter := range originCommitsAfter {
			mergePoints := s.gitRepo.FindMergePoints(originCommitAfter.CommitId)

			var originUpdate OriginUpdate
			originUpdate.Commit = originCommitAfter

			if len(mergePoints) > 0 {
				originUpdate.MergePoint = mergePoints[len(mergePoints)-1] // todo: always the last? branch to branch to main possible?
			}

			fileInfo.OriginUpdates = append(fileInfo.OriginUpdates, originUpdate)
		}

		fileInfoList[i] = fileInfo

		fmt.Printf("%+v\n", &fileInfo)
	}

	return fileInfoList
}

func ListFiles(root string) ([]string, error) {
	var files []string

	err := filepath.WalkDir(root, func(path string, d os.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if !d.IsDir() {
			relPath, err := filepath.Rel(root, path)
			if err != nil {
				return err
			}
			files = append(files, relPath)
		}
		return nil
	})

	if err != nil {
		return nil, err
	}

	return files, nil
}

var repoDirPath = "../kubernetes-website"

func Run() {
	langRelPaths, err := ListFiles(repoDirPath + "/content/pl")
	if err != nil {
		log.Fatal(err)
	}

	gitRepo := &LocalGitRepo{
		path: repoDirPath,
	}

	seeker := &GitLangSeeker{
		gitRepo: gitRepo,
	}

	result := seeker.checkFiles(langRelPaths)
	b, err := json.MarshalIndent(&result, "", "\t")
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println(string(b))
}

func main() {
	Run()
}
