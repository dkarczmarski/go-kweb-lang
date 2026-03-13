// Package gitseek seeks for differences between the EN and other languages content.
// It analyzes Git history of language-specific files and compares them with the
// corresponding English source files to determine whether translations are outdated
// or the source file has changed.
package gitseek

import (
	"context"
	"errors"
	"fmt"
	"path/filepath"

	"github.com/dkarczmarski/go-kweb-lang/git"
	"github.com/dkarczmarski/go-kweb-lang/githist"
)

const (
	// StatusEnFileDoesNotExist indicates that the EN file does not exist
	// and there are no commits after the fork point.
	StatusEnFileDoesNotExist = "en-file-does-not-exist"

	// StatusEnFileNoLongerExists indicates that the EN file existed in the past
	// but has been removed after the fork point.
	StatusEnFileNoLongerExists = "en-file-no-longer-exists"

	// StatusEnFileUpdated indicates that the EN file has new commits after
	// the fork or last translation commit.
	StatusEnFileUpdated = "en-file-updated"
)

// FileInfo contains information about the state of a language file compared
// to its English counterpart.
type FileInfo struct {
	// LangPath is the path to the language file.
	LangPath string

	// LangLastCommit is the last commit affecting the language file.
	LangLastCommit git.CommitInfo

	// LangMergeCommit is the merge commit associated with the last language commit.
	LangMergeCommit *git.CommitInfo

	// LangForkCommit is the commit where the language file diverged from the EN file.
	LangForkCommit *git.CommitInfo

	// FileStatus describes the detected state of the EN file relative to the language file.
	FileStatus string

	// EnUpdates lists commits that modified the EN file after the fork point.
	EnUpdates []EnUpdate
}

// Pair represents a mapping between an English file and its translated version.
type Pair struct {
	// EnPath is the path to the English source file.
	EnPath string

	// LangPath is the path to the translated language file.
	LangPath string
}

// EnUpdate represents a single update to the English file after the fork point.
type EnUpdate struct {
	// Commit is the commit that modified the EN file.
	Commit git.CommitInfo

	// MergePoint is the merge commit associated with that change.
	MergePoint *git.CommitInfo
}

// GitRepo defines the operations required to inspect the history of files
// in a Git repository.
type GitRepo interface {
	// FindFileLastCommit returns the most recent commit that modified the file.
	FindFileLastCommit(ctx context.Context, path string) (git.CommitInfo, error)

	// FindFileCommitsAfter returns commits affecting the file after the given commit ID.
	FindFileCommitsAfter(ctx context.Context, path string, commitIDFrom string) ([]git.CommitInfo, error)

	// FileExists checks whether the given file currently exists in the repository.
	FileExists(path string) (bool, error)
}

// GitRepoHist defines operations related to merge and fork history in Git.
type GitRepoHist interface {
	// FindForkCommit finds the commit where a branch diverged from another.
	FindForkCommit(ctx context.Context, commitID string) (*git.CommitInfo, error)

	// FindMergeCommit finds the merge commit associated with the given commit.
	FindMergeCommit(ctx context.Context, commitID string) (*git.CommitInfo, error)
}

// CacheStorage defines a storage abstraction used by GitSeek to cache FileInfo
// results. A bucket usually maps to a directory on disk, and the key identifies
// a specific cached entry.
type CacheStorage interface {
	Read(bucket, key string, buff any) (bool, error)
	Write(bucket, key string, data any) error
	Delete(bucket, key string) error
}

// GitSeek provides functionality for detecting differences between EN files
// and their translated counterparts using Git history.
type GitSeek struct {
	gitRepo     GitRepo
	gitRepoHist GitRepoHist
	cache       CacheStorage
}

// New creates a new GitSeek instance using the provided Git repository
// implementations and cache storage.
func New(gitRepo GitRepo, gitRepoHist GitRepoHist, cache CacheStorage) *GitSeek {
	return &GitSeek{
		gitRepo:     gitRepo,
		gitRepoHist: gitRepoHist,
		cache:       cache,
	}
}

// FileInfoCacheBucket returns the cache bucket used for file info of a given language.
func FileInfoCacheBucket(langCode string) string {
	return filepath.Join("lang", langCode, "git-file-info")
}

// CheckLang analyzes the provided file pair for a given language and returns
// FileInfo describing the relationship between the language file and the EN file.
// The result may be returned from cache if available.
func (gs *GitSeek) CheckLang(ctx context.Context, langCode string, pair Pair) (FileInfo, error) {
	return gs.checkFileCached(ctx, pair, langCode)
}

// InvalidateFile removes the cached FileInfo entry for the specified language file.
func (gs *GitSeek) InvalidateFile(langCode, path string) error {
	bucket := FileInfoCacheBucket(langCode)

	if err := gs.cache.Delete(bucket, path); err != nil {
		return fmt.Errorf("invalidate gitseek cache for (%s)%s: %w", langCode, path, err)
	}

	return nil
}

func (gs *GitSeek) checkFileCached(ctx context.Context, pair Pair, langCode string) (FileInfo, error) {
	bucket := FileInfoCacheBucket(langCode)
	key := pair.LangPath

	var cached FileInfo

	exists, err := gs.cache.Read(bucket, key, &cached)
	if err != nil {
		var zero FileInfo

		return zero, fmt.Errorf("read file info from cache for (%s)%s: %w", langCode, key, err)
	}

	if exists {
		return cached, nil
	}

	fileInfo, err := gs.checkFile(ctx, pair)
	if err != nil {
		return fileInfo, err
	}

	if err := gs.cache.Write(bucket, key, fileInfo); err != nil {
		var zero FileInfo

		return zero, fmt.Errorf("write file info to cache for (%s)%s: %w", langCode, key, err)
	}

	return fileInfo, nil
}

func (gs *GitSeek) checkFile(ctx context.Context, pair Pair) (FileInfo, error) {
	var fileInfo FileInfo

	enFilePath := pair.EnPath
	langFilePath := pair.LangPath

	fileInfo.LangPath = langFilePath

	langLastCommit, err := gs.gitRepo.FindFileLastCommit(ctx, langFilePath)
	if err != nil {
		return fileInfo, fmt.Errorf("error while finding the last commit of the file %s: %w", langFilePath, err)
	}

	fileInfo.LangLastCommit = langLastCommit

	mergeCommit, err := gs.gitRepoHist.FindMergeCommit(ctx, langLastCommit.CommitID)
	if err != nil {
		if !errors.Is(err, githist.ErrCommitOnMainBranch) {
			return fileInfo, fmt.Errorf("find merge commit for %s: %w", langLastCommit.CommitID, err)
		}
	}

	fileInfo.LangMergeCommit = mergeCommit

	forkCommit, err := gs.gitRepoHist.FindForkCommit(ctx, langLastCommit.CommitID)
	if err != nil {
		if !errors.Is(err, githist.ErrCommitOnMainBranch) {
			return fileInfo, fmt.Errorf("find fork commit for %s: %w", langLastCommit.CommitID, err)
		}
	}

	fileInfo.LangForkCommit = forkCommit

	startPoint := determineStartPoint(forkCommit, langLastCommit)

	enCommitsAfter, err := gs.gitRepo.FindFileCommitsAfter(ctx, enFilePath, startPoint.CommitID)
	if err != nil {
		return fileInfo, fmt.Errorf("error while finding commits after commit %s: %w", startPoint.CommitID, err)
	}

	exists, err := gs.gitRepo.FileExists(enFilePath)
	if err != nil {
		return fileInfo, fmt.Errorf("error while checking if the file %s exists: %w", enFilePath, err)
	}

	fileInfo.FileStatus = determineFileStatus(exists, enCommitsAfter)

	enUpdates, err := gs.getEnUpdates(ctx, enCommitsAfter)
	if err != nil {
		return fileInfo, err
	}

	fileInfo.EnUpdates = enUpdates

	return fileInfo, nil
}

func (gs *GitSeek) getEnUpdates(ctx context.Context, enCommitsAfter []git.CommitInfo) ([]EnUpdate, error) {
	if len(enCommitsAfter) == 0 {
		return nil, nil
	}

	enUpdates := make([]EnUpdate, 0, len(enCommitsAfter))

	for _, enCommitAfter := range enCommitsAfter {
		mergePoint, err := gs.gitRepoHist.FindMergeCommit(ctx, enCommitAfter.CommitID)
		if err != nil {
			if !errors.Is(err, githist.ErrCommitOnMainBranch) {
				return nil, fmt.Errorf("find merge commit for EN update %s: %w", enCommitAfter.CommitID, err)
			}
		}

		enUpdate := EnUpdate{
			Commit:     enCommitAfter,
			MergePoint: mergePoint,
		}

		enUpdates = append(enUpdates, enUpdate)
	}

	return enUpdates, nil
}

func determineStartPoint(forkCommit *git.CommitInfo, langLastCommit git.CommitInfo) git.CommitInfo {
	if forkCommit != nil {
		return *forkCommit
	}

	return langLastCommit
}

func determineFileStatus(exists bool, enCommitsAfter []git.CommitInfo) string {
	if !exists {
		if len(enCommitsAfter) > 0 {
			return StatusEnFileNoLongerExists
		}

		return StatusEnFileDoesNotExist
	}

	if len(enCommitsAfter) > 0 {
		return StatusEnFileUpdated
	}

	return ""
}
