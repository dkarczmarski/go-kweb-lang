package dashboard

import (
	"context"
	"fmt"
	"log"

	"github.com/dkarczmarski/go-kweb-lang/filepairs"
	"github.com/dkarczmarski/go-kweb-lang/gitseek"
	"github.com/dkarczmarski/go-kweb-lang/langcnt"
	"github.com/dkarczmarski/go-kweb-lang/pullreq"
)

type RefreshTask struct {
	langCodesProvider *langcnt.LangCodesProvider
	pairProviders     *filepairs.PairProviders
	gitSeeker         *gitseek.GitSeek
	filePRIndex       *pullreq.FilePRIndex
	store             *Store
}

func NewRefreshTask(
	langCodesProvider *langcnt.LangCodesProvider,
	pairProviders *filepairs.PairProviders,
	gitSeeker *gitseek.GitSeek,
	filePRIndex *pullreq.FilePRIndex,
	store *Store,
) *RefreshTask {
	return &RefreshTask{
		langCodesProvider: langCodesProvider,
		pairProviders:     pairProviders,
		gitSeeker:         gitSeeker,
		filePRIndex:       filePRIndex,
		store:             store,
	}
}

func (t *RefreshTask) Run(ctx context.Context) error {
	langCodes, err := t.langCodesProvider.LangCodes()
	if err != nil {
		return fmt.Errorf("failed to get available languages: %w", err)
	}

	for _, langCode := range langCodes {
		langDashboard, buildErr := t.buildDashboard(ctx, langCode)
		if buildErr != nil {
			return buildErr
		}

		writeErr := t.store.WriteDashboard(langDashboard)
		if writeErr != nil {
			return fmt.Errorf("failed to write dashboard for lang code %s: %w", langCode, writeErr)
		}
	}

	langIndex, err := t.buildLangIndex()
	if err != nil {
		return err
	}

	writeErr := t.store.WriteDashboardIndex(langIndex)
	if writeErr != nil {
		return fmt.Errorf("failed to write dashboard index: %w", writeErr)
	}

	return nil
}

func (t *RefreshTask) buildDashboard(ctx context.Context, langCode string) (Dashboard, error) {
	pairs, err := t.pairProviders.ListPairs(langCode)
	if err != nil {
		return Dashboard{}, fmt.Errorf(
			"error while listing file pairs for lang code %s: %w",
			langCode,
			err,
		)
	}

	seekerFileInfos := make([]gitseek.FileInfo, 0, len(pairs))

	for i, pair := range pairs {
		log.Printf(
			"[gitseek][%s][%d/%d] checking for updates for %v",
			langCode,
			i+1,
			len(pairs),
			pair.LangPath,
		)

		fileInfo, checkErr := t.gitSeeker.CheckLang(ctx, langCode, gitseek.Pair{
			EnPath:   pair.EnPath,
			LangPath: pair.LangPath,
		})
		if checkErr != nil {
			return Dashboard{}, fmt.Errorf(
				"error while checking file pair %s for the language code %s: %w",
				pair.LangPath,
				langCode,
				checkErr,
			)
		}

		seekerFileInfos = append(seekerFileInfos, fileInfo)
	}

	prIndex, err := t.filePRIndex.LangIndex(langCode)
	if err != nil {
		return Dashboard{}, fmt.Errorf(
			"error while getting pull request index for lang code %s: %w",
			langCode,
			err,
		)
	}

	return buildDashboard(langCode, seekerFileInfos, prIndex), nil
}

func (t *RefreshTask) buildLangIndex() (LangIndex, error) {
	return buildLangIndex(t.langCodesProvider)
}
