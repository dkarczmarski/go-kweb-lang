package tasks

import (
	"context"
	"errors"
	"fmt"
	"log"

	"github.com/dkarczmarski/go-kweb-lang/filepairs"
	"github.com/dkarczmarski/go-kweb-lang/githist"
	"github.com/dkarczmarski/go-kweb-lang/langcnt"
)

type RefreshRepoTask struct {
	gitRepoHist       *githist.GitHist
	filePaths         *filepairs.FilePaths
	langCodesProvider *langcnt.LangCodesProvider
	invalidator       githist.Invalidator
}

func NewRefreshRepoTask(
	gitRepoHist *githist.GitHist,
	filePaths *filepairs.FilePaths,
	langCodesProvider *langcnt.LangCodesProvider,
	invalidator githist.Invalidator,
) *RefreshRepoTask {
	return &RefreshRepoTask{
		gitRepoHist:       gitRepoHist,
		filePaths:         filePaths,
		langCodesProvider: langCodesProvider,
		invalidator:       invalidator,
	}
}

func (t *RefreshRepoTask) Run(ctx context.Context) error {
	filesToInvalidate, err := t.gitRepoHist.PullRefresh(ctx)
	if err != nil {
		return fmt.Errorf("git cache pull refresh error: %w", err)
	}

	invalidated := make(map[string]int)
	for i, file := range filesToInvalidate {
		if invalidatedAt, isInvalidated := invalidated[file]; !isInvalidated {
			log.Printf("[githist][%d/%d] invalidate file %s", i+1, len(filesToInvalidate), file)

			if err := t.invalidateChangedPath(file); err != nil {
				return fmt.Errorf("invalidate changed path %s error: %w", file, err)
			}

			invalidated[file] = i
		} else {
			log.Printf(
				"[githist][%d/%d] invalidate file %s - (skip) already done at %d",
				i+1,
				len(filesToInvalidate),
				file,
				invalidatedAt,
			)
		}
	}

	return nil
}

// invalidateChangedPath invalidates the cache for a given file path.
func (t *RefreshRepoTask) invalidateChangedPath(changedPath string) error {
	if t.invalidator == nil {
		return nil
	}

	pathInfo, err := t.filePaths.CheckPath(changedPath)
	if err != nil {
		if errors.Is(err, filepairs.ErrPairMatcherNotFound) {
			return nil
		}

		return fmt.Errorf("check changed path %s: %w", changedPath, err)
	}

	if pathInfo.IsEnPath() {
		langCodes, err := t.langCodesProvider.LangCodes()
		if err != nil {
			return fmt.Errorf("get language codes: %w", err)
		}

		for _, langCode := range langCodes {
			langPath, err := pathInfo.LangPath(langCode)
			if err != nil {
				return fmt.Errorf("build language path for %s: %w", langCode, err)
			}

			if err := t.invalidator.InvalidateFile(langCode, langPath); err != nil {
				return fmt.Errorf("invalidate gitseek cache for (%s)%s: %w", langCode, langPath, err)
			}
		}

		return nil
	}

	if err := t.invalidator.InvalidateFile(pathInfo.LangCode, changedPath); err != nil {
		return fmt.Errorf("invalidate gitseek cache for (%s)%s: %w", pathInfo.LangCode, changedPath, err)
	}

	return nil
}
