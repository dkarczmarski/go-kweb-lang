package main

import (
	"encoding/json"
	"fmt"
	"go-kweb-lang/git"
	"log"
	"os"
	"path/filepath"
)

type FileInfo struct {
	LangRelPath      string
	LangCommit       git.CommitInfo
	OriginFileStatus string
	OriginUpdates    []OriginUpdate
}

type OriginUpdate struct {
	Commit     git.CommitInfo
	MergePoint git.CommitInfo
}

type GitLangSeeker struct {
	gitRepo git.Repo
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
		langLastCommit, err := s.gitRepo.FindFileLastCommit(langFilePath)
		if err != nil {
			log.Fatal(err)
		}
		fileInfo.LangCommit = langLastCommit

		exists, err := s.gitRepo.FileExists(originFilePath)
		if err != nil {
			log.Fatal(err)
		}
		if !exists {
			fileInfo.OriginFileStatus = "NOT_EXIST"
		}

		originCommitsAfter, err := s.gitRepo.FindFileCommitsAfter(originFilePath, langLastCommit.CommitId)
		if err != nil {
			log.Fatal(err)
		}
		if len(originCommitsAfter) > 0 {
			fileInfo.OriginFileStatus = "MODIFIED"
		}

		for _, originCommitAfter := range originCommitsAfter {
			mergePoints, err := s.gitRepo.FindMergePoints(originCommitAfter.CommitId)
			if err != nil {
				log.Fatal(err)
			}

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

	gitRepo := git.NewRepo(repoDirPath)

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
