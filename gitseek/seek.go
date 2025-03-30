// Package gitseek seeks for differences between the EN and other languages content
package gitseek

import (
	"context"
	"fmt"
	"go-kweb-lang/git"
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

func NewGitLangSeeker(gitRepo git.Repo) *GitLangSeeker {
	return &GitLangSeeker{
		gitRepo: gitRepo,
	}
}

func (s *GitLangSeeker) CheckLang(ctx context.Context, langCode string) ([]FileInfo, error) {
	langRelPaths, err := s.gitRepo.ListFiles("/content/" + langCode)
	if err != nil {
		return nil, fmt.Errorf("error while listing content files for the language code %s: %w", langCode, err)
	}

	return s.CheckFiles(ctx, langRelPaths, langCode)
}

func (s *GitLangSeeker) CheckFiles(ctx context.Context, langRelPaths []string, langCode string) ([]FileInfo, error) {
	fileInfoList := make([]FileInfo, 0, len(langRelPaths))

	for _, langRelPath := range langRelPaths {
		var fileInfo FileInfo

		originFilePath := repoOriginFilePath(langRelPath)
		langFilePath := repoLangFilePath(langRelPath, langCode)

		fileInfo.LangRelPath = langRelPath
		langLastCommit, err := s.gitRepo.FindFileLastCommit(ctx, langFilePath)
		if err != nil {
			return nil, fmt.Errorf("error while finding the last commit of the file %s: %w", langFilePath, err)
		}
		fileInfo.LangCommit = langLastCommit

		exists, err := s.gitRepo.FileExists(originFilePath)
		if err != nil {
			return nil, fmt.Errorf("error while checking if the file %s exists: %w", originFilePath, err)
		}
		if !exists {
			fileInfo.OriginFileStatus = "NOT_EXIST"
		}

		originCommitsAfter, err := s.gitRepo.FindFileCommitsAfter(ctx, originFilePath, langLastCommit.CommitID)
		if err != nil {
			return nil, fmt.Errorf("error while finding commits after commit %s: %w",
				langLastCommit.CommitID, err)
		}
		if len(originCommitsAfter) > 0 {
			fileInfo.OriginFileStatus = "MODIFIED"
		}

		for _, originCommitAfter := range originCommitsAfter {
			mergePoints, err := s.gitRepo.FindMergePoints(ctx, originCommitAfter.CommitID)
			if err != nil {
				return nil, fmt.Errorf("error while finding merge points for the commit %s: %w",
					originCommitAfter.CommitID, err)
			}

			var originUpdate OriginUpdate
			originUpdate.Commit = originCommitAfter

			if len(mergePoints) > 0 {
				originUpdate.MergePoint = mergePoints[len(mergePoints)-1] // todo: always the last? branch to branch to main possible?
			}

			fileInfo.OriginUpdates = append(fileInfo.OriginUpdates, originUpdate)
		}

		fileInfoList = append(fileInfoList, fileInfo)

		fmt.Printf("%+v\n", &fileInfo)
	}

	return fileInfoList, nil
}

func repoOriginFilePath(relPath string) string {
	return repoLangFilePath(relPath, "en")
}

func repoLangFilePath(relPath string, langCode string) string {
	return filepath.Join("content", langCode, relPath)
}
