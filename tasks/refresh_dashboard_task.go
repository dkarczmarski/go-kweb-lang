package tasks

import (
	"context"
	"fmt"
	"log"

	"github.com/dkarczmarski/go-kweb-lang/dashboard"
	"github.com/dkarczmarski/go-kweb-lang/filepairs"
	"github.com/dkarczmarski/go-kweb-lang/gitseek"
	"github.com/dkarczmarski/go-kweb-lang/langcnt"
	"github.com/dkarczmarski/go-kweb-lang/pullreq"
)

type RefreshDashboardTask struct {
	langCodesProvider *langcnt.LangCodesProvider
	pairProviders     *filepairs.PairProviders
	gitSeeker         *gitseek.GitSeek
	filePRIndex       *pullreq.FilePRIndex
	store             *dashboard.Store
}

func NewRefreshDashboardTask(
	langCodesProvider *langcnt.LangCodesProvider,
	pairProviders *filepairs.PairProviders,
	gitSeeker *gitseek.GitSeek,
	filePRIndex *pullreq.FilePRIndex,
	store *dashboard.Store,
) *RefreshDashboardTask {
	return &RefreshDashboardTask{
		langCodesProvider: langCodesProvider,
		pairProviders:     pairProviders,
		gitSeeker:         gitSeeker,
		filePRIndex:       filePRIndex,
		store:             store,
	}
}

func (t *RefreshDashboardTask) Run(ctx context.Context) error {
	langCodes, err := t.langCodesProvider.LangCodes()
	if err != nil {
		return fmt.Errorf("get available languages: %w", err)
	}

	for _, langCode := range langCodes {
		langDashboard, err := t.buildDashboard(ctx, langCode)
		if err != nil {
			return fmt.Errorf("build dashboard for lang code %s: %w", langCode, err)
		}

		if err := t.store.WriteDashboard(langDashboard); err != nil {
			return fmt.Errorf("write dashboard for lang code %s: %w", langCode, err)
		}
	}

	langIndex, err := t.buildLangIndex()
	if err != nil {
		return fmt.Errorf("build dashboard language index: %w", err)
	}

	if err := t.store.WriteDashboardIndex(langIndex); err != nil {
		return fmt.Errorf("write dashboard language index: %w", err)
	}

	return nil
}

func (t *RefreshDashboardTask) buildDashboard(
	ctx context.Context,
	langCode string,
) (dashboard.Dashboard, error) {
	pairs, err := t.pairProviders.ListPairs(langCode)
	if err != nil {
		return dashboard.Dashboard{}, fmt.Errorf(
			"list file pairs for lang code %s: %w",
			langCode,
			err,
		)
	}

	seekerFileInfos := make([]gitseek.FileInfo, 0, len(pairs))

	for pairIndex, pair := range pairs {
		log.Printf(
			"[gitseek][%s][%d/%d] checking for updates for %v",
			langCode,
			pairIndex+1,
			len(pairs),
			pair.LangPath,
		)

		fileInfo, err := t.gitSeeker.CheckLang(ctx, langCode, gitseek.Pair{
			EnPath:   pair.EnPath,
			LangPath: pair.LangPath,
		})
		if err != nil {
			return dashboard.Dashboard{}, fmt.Errorf(
				"check file pair %s for lang code %s: %w",
				pair.LangPath,
				langCode,
				err,
			)
		}

		seekerFileInfos = append(seekerFileInfos, fileInfo)
	}

	prIndex, err := t.filePRIndex.LangIndex(langCode)
	if err != nil {
		return dashboard.Dashboard{}, fmt.Errorf(
			"get pull request index for lang code %s: %w",
			langCode,
			err,
		)
	}

	return dashboard.BuildDashboard(langCode, seekerFileInfos, prIndex), nil
}

func (t *RefreshDashboardTask) buildLangIndex() (dashboard.LangIndex, error) {
	langIndex, err := dashboard.BuildLangIndex(t.langCodesProvider)
	if err != nil {
		return dashboard.LangIndex{}, fmt.Errorf("build lang index: %w", err)
	}

	return langIndex, nil
}
