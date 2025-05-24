// Package gitseek seeks for differences between the EN and other languages content
package gitseek

import (
	"context"
	"fmt"
	"log"
	"os"
	"path/filepath"

	"go-kweb-lang/proxycache"

	"go-kweb-lang/githist"

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
	gitRepo     git.Repo
	gitRepoHist *githist.GitHist
	cacheDir    string
}

func New(gitRepo git.Repo, gitRepoHist *githist.GitHist, cacheDir string) *GitSeek {
	return &GitSeek{
		gitRepo:     gitRepo,
		gitRepoHist: gitRepoHist,
		cacheDir:    cacheDir,
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

		fileInfo, err := s.checkFileCached(ctx, langRelPath, langCode)
		if err != nil {
			return nil, err
		}

		fileInfoList = append(fileInfoList, fileInfo)
	}

	return fileInfoList, nil
}

func (s *GitSeek) checkFileCached(ctx context.Context, langRelPath string, langCode string) (FileInfo, error) {
	return proxycache.Get(
		ctx,
		s.cacheDir,
		fileInfoCacheBucket(langCode),
		langRelPath,
		nil,
		func(ctx context.Context) (FileInfo, error) {
			return s.checkFile(ctx, langRelPath, langCode)
		},
	)
}

func (s *GitSeek) checkFile(ctx context.Context, langRelPath string, langCode string) (FileInfo, error) {
	var fileInfo FileInfo

	originFilePath := repoOriginFilePath(langRelPath)
	langFilePath := repoLangFilePath(langRelPath, langCode)

	fileInfo.LangRelPath = langRelPath

	langLastCommit, err := s.gitRepo.FindFileLastCommit(ctx, langFilePath)
	if err != nil {
		return fileInfo, fmt.Errorf("error while finding the last commit of the file %s: %w", langFilePath, err)
	}

	fileInfo.LangLastCommit = langLastCommit

	forkCommit, err := s.gitRepoHist.FindForkCommit(ctx, langLastCommit.CommitID)
	if err != nil {
		return fileInfo, err
	}

	fileInfo.LangForkCommit = forkCommit

	var startPoint git.CommitInfo
	if forkCommit != nil {
		startPoint = *forkCommit
	} else {
		startPoint = langLastCommit
	}

	originCommitsAfter, err := s.gitRepo.FindFileCommitsAfter(ctx, originFilePath, startPoint.CommitID)
	if err != nil {
		return fileInfo, fmt.Errorf("error while finding commits after commit %s: %w",
			langLastCommit.CommitID, err)
	}

	exists, err := s.gitRepo.FileExists(originFilePath)
	if err != nil {
		return fileInfo, fmt.Errorf("error while checking if the file %s exists: %w", originFilePath, err)
	}
	if !exists {
		fileInfo.OriginFileStatus = "NOT_EXIST"
	} else if len(originCommitsAfter) > 0 {
		fileInfo.OriginFileStatus = "MODIFIED"
	}

	for _, originCommitAfter := range originCommitsAfter {
		mergePoint, err := s.gitRepoHist.FindMergeCommit(ctx, originCommitAfter.CommitID)
		if err != nil {
			return fileInfo, err
		}

		originUpdate := OriginUpdate{
			Commit:     originCommitAfter,
			MergePoint: mergePoint,
		}

		fileInfo.OriginUpdates = append(fileInfo.OriginUpdates, originUpdate)
	}

	return fileInfo, nil
}

func (s *GitSeek) InvalidateFiles(langRelPaths []string) error {
	parentDir := filepath.Join(s.cacheDir, "lang")

	langDirs, err := listSubdirectories(parentDir)
	if err != nil {
		return fmt.Errorf("failed to list cache lang directories %v: %w", parentDir, err)
	}

	for _, langRelPath := range langRelPaths {
		log.Printf("invalidate path: %v", langRelPath)

		for _, langCode := range langDirs {
			bucket := fileInfoCacheBucket(langCode)

			if err := proxycache.InvalidateKey(s.cacheDir, bucket, langRelPath); err != nil {
				return fmt.Errorf("failed to invalidate path %v: %w", langRelPath, err)
			}
		}
	}

	return nil
}

func listSubdirectories(path string) ([]string, error) {
	entries, err := os.ReadDir(path)
	if err != nil {
		if os.IsNotExist(err) {
			return []string{}, nil
		}

		return nil, fmt.Errorf("failed to read directory %q: %w", path, err)
	}

	var dirs []string
	for _, entry := range entries {
		if entry.IsDir() {
			dirs = append(dirs, entry.Name())
		}
	}

	return dirs, nil
}

func fileInfoCacheBucket(langCode string) string {
	return filepath.Join("lang", langCode, "git-file-info")
}

func repoOriginFilePath(relPath string) string {
	return repoLangFilePath(relPath, "en")
}

func repoLangFilePath(relPath string, langCode string) string {
	return filepath.Join("content", langCode, relPath)
}
