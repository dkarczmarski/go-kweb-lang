// Package gitseek seeks for differences between the EN and other languages content
package gitseek

//go:generate mockgen -typed -source=gitseek.go -destination=./internal/mocks/mocks.go -package=mocks

import (
	"context"
	"fmt"
	"log"
	"path/filepath"
	"strings"

	"go-kweb-lang/git"
	"go-kweb-lang/proxycache"
)

type FileInfo struct {
	LangRelPath     string
	LangLastCommit  git.CommitInfo
	LangMergeCommit *git.CommitInfo
	LangForkCommit  *git.CommitInfo
	ENFileStatus    string
	ENUpdates       []ENUpdate
}

type ENUpdate struct {
	Commit     git.CommitInfo
	MergePoint *git.CommitInfo
}

type GitRepo interface {
	FindFileLastCommit(ctx context.Context, path string) (git.CommitInfo, error)
	FindFileCommitsAfter(ctx context.Context, path string, commitIDFrom string) ([]git.CommitInfo, error)
	FileExists(path string) (bool, error)
	ListFiles(path string) ([]string, error)
}

type GitRepoHist interface {
	FindForkCommit(ctx context.Context, commitID string) (*git.CommitInfo, error)
	FindMergeCommit(ctx context.Context, commitID string) (*git.CommitInfo, error)
}

type CacheStore interface {
	Read(bucket, key string, buff any) (bool, error)
	Write(bucket, key string, data any) error
	Delete(bucket, key string) error
	ListBuckets(bucketPth string) ([]string, error)
}

type GitSeek struct {
	gitRepo     GitRepo
	gitRepoHist GitRepoHist
	cacheStore  CacheStore
}

func New(gitRepo GitRepo, gitRepoHist GitRepoHist, cacheStore CacheStore) *GitSeek {
	return &GitSeek{
		gitRepo:     gitRepo,
		gitRepoHist: gitRepoHist,
		cacheStore:  cacheStore,
	}
}

// CheckLang checks all files in the content/langCode directory for the given langCode.
func (gs *GitSeek) CheckLang(ctx context.Context, langCode string) ([]FileInfo, error) {
	langRelPaths, err := gs.gitRepo.ListFiles("/content/" + langCode)
	if err != nil {
		return nil, fmt.Errorf("error while listing content files for the language code %s: %w", langCode, err)
	}

	// skip selected files that do not make sense to compare
	langRelPaths = removeStrings(langRelPaths, []string{"OWNERS"})

	return gs.CheckFiles(ctx, langRelPaths, langCode)
}

func removeStrings(input []string, toRemove []string) []string {
	removeMap := make(map[string]bool)
	for _, val := range toRemove {
		removeMap[val] = true
	}

	var result []string
	for _, val := range input {
		if !removeMap[val] {
			result = append(result, val)
		}
	}

	return result
}

// CheckFiles examines selected files in the content/langCode directory for the given langCode
// for corresponding updates in the content/en directory.
func (gs *GitSeek) CheckFiles(ctx context.Context, langRelPaths []string, langCode string) ([]FileInfo, error) {
	fileInfoList := make([]FileInfo, 0, len(langRelPaths))

	langRelPathsLen := len(langRelPaths)

	for i, langRelPath := range langRelPaths {
		log.Printf("[%v][%v/%v] checking for updates for %v", langCode, i, langRelPathsLen, langRelPath)

		fileInfo, err := gs.checkFileCached(ctx, langRelPath, langCode)
		if err != nil {
			return nil, err
		}

		fileInfoList = append(fileInfoList, fileInfo)
	}

	return fileInfoList, nil
}

func (gs *GitSeek) checkFileCached(ctx context.Context, langRelPath string, langCode string) (FileInfo, error) {
	return proxycache.Get(
		ctx,
		gs.cacheStore,
		fileInfoCacheBucket(langCode),
		langRelPath,
		nil,
		func(ctx context.Context) (FileInfo, error) {
			return gs.checkFile(ctx, langRelPath, langCode)
		},
	)
}

func (gs *GitSeek) checkFile(ctx context.Context, langRelPath string, langCode string) (FileInfo, error) {
	var fileInfo FileInfo

	enFilePath := repoENFilePath(langRelPath)
	langFilePath := repoLangFilePath(langRelPath, langCode)

	fileInfo.LangRelPath = langRelPath

	langLastCommit, err := gs.gitRepo.FindFileLastCommit(ctx, langFilePath)
	if err != nil {
		return fileInfo, fmt.Errorf("error while finding the last commit of the file %s: %w", langFilePath, err)
	}

	fileInfo.LangLastCommit = langLastCommit

	mergeCommit, err := gs.gitRepoHist.FindMergeCommit(ctx, langLastCommit.CommitID)
	if err != nil {
		return fileInfo, err
	}

	fileInfo.LangMergeCommit = mergeCommit

	forkCommit, err := gs.gitRepoHist.FindForkCommit(ctx, langLastCommit.CommitID)
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

	enCommitsAfter, err := gs.gitRepo.FindFileCommitsAfter(ctx, enFilePath, startPoint.CommitID)
	if err != nil {
		return fileInfo, fmt.Errorf("error while finding commits after commit %s: %w",
			langLastCommit.CommitID, err)
	}

	exists, err := gs.gitRepo.FileExists(enFilePath)
	if err != nil {
		return fileInfo, fmt.Errorf("error while checking if the file %s exists: %w", enFilePath, err)
	}
	if !exists {
		if len(enCommitsAfter) > 0 {
			fileInfo.ENFileStatus = "DELETED"
		} else {
			fileInfo.ENFileStatus = "NOT_EXIST"
		}
	} else if len(enCommitsAfter) > 0 {
		fileInfo.ENFileStatus = "MODIFIED"
	}

	for _, enCommitAfter := range enCommitsAfter {
		mergePoint, err := gs.gitRepoHist.FindMergeCommit(ctx, enCommitAfter.CommitID)
		if err != nil {
			return fileInfo, err
		}

		enUpdate := ENUpdate{
			Commit:     enCommitAfter,
			MergePoint: mergePoint,
		}

		fileInfo.ENUpdates = append(fileInfo.ENUpdates, enUpdate)
	}

	return fileInfo, nil
}

func (gs *GitSeek) InvalidateFile(file string) error {
	if ok, langCode, relPath := checkFileToInvalidate(file); ok {
		if err := gs.invalidateRelPath(langCode, relPath); err != nil {
			return fmt.Errorf("invalidate error: %w", err)
		}
	}

	return nil
}

func checkFileToInvalidate(fullPath string) (bool, string, string) {
	const contentStr = "content/"

	if !strings.HasPrefix(fullPath, contentStr) {
		return false, "", ""
	}

	langCodeLen := strings.Index(fullPath[len(contentStr):], "/")
	if langCodeLen < 0 {
		return false, "", ""
	}

	if len(contentStr)+langCodeLen+1 == len(fullPath) {
		return false, "", ""
	}

	langCode := fullPath[len(contentStr) : len(contentStr)+langCodeLen]

	return true, langCode, fullPath[len(contentStr)+langCodeLen+1:]
}

func (gs *GitSeek) invalidateRelPath(langCode, relPath string) error {
	log.Printf("init gitseek cache invalidation for (%s)%s", langCode, relPath)

	var dirsToInvalidate []string
	if langCode == "en" {
		langDirs, err := gs.cacheStore.ListBuckets("lang")
		if err != nil {
			return fmt.Errorf("failed to list cache lang directories: %w", err)
		}

		dirsToInvalidate = append(dirsToInvalidate, "en")
		dirsToInvalidate = append(dirsToInvalidate, langDirs...)
	} else {
		dirsToInvalidate = append(dirsToInvalidate, langCode)
	}

	for _, langCodeDir := range dirsToInvalidate {
		log.Printf("invalidate gitseek cache for: (%s)%s", langCodeDir, relPath)

		bucket := fileInfoCacheBucket(langCodeDir)
		if err := gs.cacheStore.Delete(bucket, relPath); err != nil {
			return fmt.Errorf("failed to invalidate path %v: %w", relPath, err)
		}
	}

	return nil
}

func fileInfoCacheBucket(langCode string) string {
	return filepath.Join("lang", langCode, "git-file-info")
}

func repoENFilePath(relPath string) string {
	return repoLangFilePath(relPath, "en")
}

func repoLangFilePath(relPath string, langCode string) string {
	return filepath.Join("content", langCode, relPath)
}
