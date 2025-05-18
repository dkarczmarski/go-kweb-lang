// Package gitseek seeks for differences between the EN and other languages content
package gitseek

import (
	"context"
	"fmt"
	"log"
	"path/filepath"

	"go-kweb-lang/gitpc"

	"go-kweb-lang/git"
)

type FileInfo struct {
	LangRelPath      string
	LangLastCommit   git.CommitInfo
	LangForkCommit   *git.CommitInfo
	OriginFileStatus string
	OriginUpdates    []OriginUpdate
}

type OriginUpdate struct {
	Commit     git.CommitInfo
	MergePoint *git.CommitInfo
}

type GitSeek struct {
	gitRepo   git.Repo
	gitRepoPC *gitpc.ProxyCache
}

func New(gitRepo git.Repo, gitRepoPC *gitpc.ProxyCache) *GitSeek {
	return &GitSeek{
		gitRepo:   gitRepo,
		gitRepoPC: gitRepoPC,
	}
}

// CheckLang checks all files in the content/langCode directory for the given langCode.
func (s *GitSeek) CheckLang(ctx context.Context, langCode string) ([]FileInfo, error) {
	langRelPaths, err := s.gitRepo.ListFiles("/content/" + langCode)
	if err != nil {
		return nil, fmt.Errorf("error while listing content files for the language code %s: %w", langCode, err)
	}

	return s.CheckFiles(ctx, langRelPaths, langCode)
}

// CheckFiles examines selected files in the content/langCode directory for the given langCode
// for corresponding updates in the content/en directory.
func (s *GitSeek) CheckFiles(ctx context.Context, langRelPaths []string, langCode string) ([]FileInfo, error) {
	fileInfoList := make([]FileInfo, 0, len(langRelPaths))

	langRelPathsLen := len(langRelPaths)

	for i, langRelPath := range langRelPaths {
		log.Printf("[%v][%v/%v] checking for updates for %v", langCode, i, langRelPathsLen, langRelPath)

		var fileInfo FileInfo

		originFilePath := repoOriginFilePath(langRelPath)
		langFilePath := repoLangFilePath(langRelPath, langCode)

		fileInfo.LangRelPath = langRelPath

		langLastCommit, err := s.gitRepoPC.FindFileLastCommit(ctx, langFilePath)
		if err != nil {
			return nil, fmt.Errorf("error while finding the last commit of the file %s: %w", langFilePath, err)
		}

		fileInfo.LangLastCommit = langLastCommit

		forkCommit, err := s.gitRepoPC.FindForkCommit(ctx, langLastCommit.CommitID)
		if err != nil {
			return nil, err
		}

		fileInfo.LangForkCommit = forkCommit

		var startPoint git.CommitInfo
		if forkCommit != nil {
			startPoint = *forkCommit
		} else {
			startPoint = langLastCommit
		}

		originCommitsAfter, err := s.gitRepoPC.FindFileCommitsAfter(ctx, originFilePath, startPoint.CommitID)
		if err != nil {
			return nil, fmt.Errorf("error while finding commits after commit %s: %w",
				langLastCommit.CommitID, err)
		}

		exists, err := s.gitRepo.FileExists(originFilePath)
		if err != nil {
			return nil, fmt.Errorf("error while checking if the file %s exists: %w", originFilePath, err)
		}
		if !exists {
			fileInfo.OriginFileStatus = "NOT_EXIST"
		} else if len(originCommitsAfter) > 0 {
			fileInfo.OriginFileStatus = "MODIFIED"
		}

		for _, originCommitAfter := range originCommitsAfter {
			mergePoint, err := s.gitRepoPC.FindMergeCommit(ctx, originCommitAfter.CommitID)
			if err != nil {
				return nil, err
			}

			originUpdate := OriginUpdate{
				Commit:     originCommitAfter,
				MergePoint: mergePoint,
			}

			fileInfo.OriginUpdates = append(fileInfo.OriginUpdates, originUpdate)
		}

		fileInfoList = append(fileInfoList, fileInfo)
	}

	return fileInfoList, nil
}

func repoOriginFilePath(relPath string) string {
	return repoLangFilePath(relPath, "en")
}

func repoLangFilePath(relPath string, langCode string) string {
	return filepath.Join("content", langCode, relPath)
}
