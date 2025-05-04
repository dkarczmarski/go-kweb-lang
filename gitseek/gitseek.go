// Package gitseek seeks for differences between the EN and other languages content
package gitseek

import (
	"context"
	"fmt"
	"log"
	"path/filepath"

	"go-kweb-lang/git"
)

type FileInfo struct {
	LangRelPath      string
	LangCommit       git.CommitInfo
	OriginFileStatus string
	OriginUpdates    []OriginUpdate
}

type OriginUpdate struct {
	Commit     git.CommitInfo
	MergePoint *git.CommitInfo
}

type GitSeek struct {
	gitRepo git.Repo
}

func New(gitRepo git.Repo) *GitSeek {
	return &GitSeek{
		gitRepo: gitRepo,
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

// CheckFiles examine selected files in the content/langCode directory for the given langCode
// for corresponding updates in the content/en directory.
func (s *GitSeek) CheckFiles(ctx context.Context, langRelPaths []string, langCode string) ([]FileInfo, error) {
	fileInfoList := make([]FileInfo, 0, len(langRelPaths))

	mainBranchCommits, err := s.gitRepo.ListMainBranchCommits(ctx)
	if err != nil {
		return nil, err
	}

	langRelPathsLen := len(langRelPaths)

	for i, langRelPath := range langRelPaths {
		log.Printf("[%v][%v/%v] checking for updates for %v", langCode, i, langRelPathsLen, langRelPath)

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
		if exists && len(originCommitsAfter) > 0 {
			fileInfo.OriginFileStatus = "MODIFIED"
		}

		for _, originCommitAfter := range originCommitsAfter {
			var mergePoint *git.CommitInfo
			if !containsCommit(mainBranchCommits, originCommitAfter.CommitID) {
				mergePoints, err := s.gitRepo.ListMergePoints(ctx, originCommitAfter.CommitID)
				if err != nil {
					return nil, fmt.Errorf("error while finding merge points for the commit %s: %w",
						originCommitAfter.CommitID, err)
				}

				mergePoint = findMergePoint(mainBranchCommits, mergePoints)
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

func containsCommit(list []git.CommitInfo, commitID string) bool {
	for i := range list {
		if list[i].CommitID == commitID {
			return true
		}
	}

	return false
}

func findMergePoint(mainBranchCommits []git.CommitInfo, mergePoints []git.CommitInfo) *git.CommitInfo {
	mergePointsLen := len(mergePoints)
	if mergePointsLen == 0 {
		return nil
	}

	for i := mergePointsLen - 1; 0 <= i; i-- {
		mergePoint := mergePoints[i]
		if containsCommit(mainBranchCommits, mergePoint.CommitID) {
			return &mergePoint
		}
	}

	log.Fatal("unexpected state: this should never happen")

	return nil
}

func repoOriginFilePath(relPath string) string {
	return repoLangFilePath(relPath, "en")
}

func repoLangFilePath(relPath string, langCode string) string {
	return filepath.Join("content", langCode, relPath)
}
