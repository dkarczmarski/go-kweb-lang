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
		return fmt.Errorf("pull refresh: %w", err)
	}

	invalidatedAtByPath := make(map[string]int, len(filesToInvalidate))

	for fileIndex, filePath := range filesToInvalidate {
		previousIndex, alreadyInvalidated := invalidatedAtByPath[filePath]
		if alreadyInvalidated {
			log.Printf(
				"[githist][%d/%d] invalidate file %s - (skip) already done at %d",
				fileIndex+1,
				len(filesToInvalidate),
				filePath,
				previousIndex+1,
			)

			continue
		}

		log.Printf(
			"[githist][%d/%d] invalidate file %s",
			fileIndex+1,
			len(filesToInvalidate),
			filePath,
		)

		if err := t.invalidateChangedPath(filePath); err != nil {
			return fmt.Errorf("invalidate changed path %s: %w", filePath, err)
		}

		invalidatedAtByPath[filePath] = fileIndex
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
				return fmt.Errorf("build language path for lang code %s: %w", langCode, err)
			}

			if err := t.invalidator.InvalidateFile(langCode, langPath); err != nil {
				return fmt.Errorf(
					"invalidate gitseek cache for lang code %s and path %s: %w",
					langCode,
					langPath,
					err,
				)
			}
		}

		return nil
	}

	if err := t.invalidator.InvalidateFile(pathInfo.LangCode, changedPath); err != nil {
		return fmt.Errorf(
			"invalidate gitseek cache for lang code %s and path %s: %w",
			pathInfo.LangCode,
			changedPath,
			err,
		)
	}

	return nil
}
