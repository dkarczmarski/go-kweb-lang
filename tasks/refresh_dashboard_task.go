package tasks

import (
	"context"
	"fmt"
	"log"

	"github.com/dkarczmarski/go-kweb-lang/dashboard"
	"github.com/dkarczmarski/go-kweb-lang/filepairs"
	"github.com/dkarczmarski/go-kweb-lang/gitseek"
	"github.com/dkarczmarski/go-kweb-lang/pullreq"
)

type PairLister interface {
	ListPairs(langCode string) ([]filepairs.Pair, error)
}

type LangChecker interface {
	CheckLang(ctx context.Context, langCode string, pair gitseek.Pair) (gitseek.FileInfo, error)
}

type FilePRIndexer interface {
	LangIndex(langCode string) (pullreq.FilePRIndexData, error)
}

type DashboardStore interface {
	WriteDashboard(langDashboard dashboard.Dashboard) error
	WriteDashboardIndex(langIndex dashboard.LangIndex) error
}

type RefreshDashboardTask struct {
	langCodesProvider dashboard.LangCodesProvider
	pairProviders     PairLister
	gitSeeker         LangChecker
	filePRIndex       FilePRIndexer
	store             DashboardStore
}

func NewRefreshDashboardTask(
	langCodesProvider dashboard.LangCodesProvider,
	pairProviders PairLister,
	gitSeeker LangChecker,
	filePRIndex FilePRIndexer,
	store DashboardStore,
) *RefreshDashboardTask {
	return &RefreshDashboardTask{
		langCodesProvider: langCodesProvider,
		pairProviders:     pairProviders,
		gitSeeker:         gitSeeker,
		filePRIndex:       filePRIndex,
		store:             store,
	}
}

func (task *RefreshDashboardTask) Run(ctx context.Context) error {
	langCodes, err := task.langCodesProvider.LangCodes()
	if err != nil {
		return fmt.Errorf("get available languages: %w", err)
	}

	for _, langCode := range langCodes {
		langDashboard, err := task.buildDashboard(ctx, langCode)
		if err != nil {
			return fmt.Errorf("build dashboard for lang code %s: %w", langCode, err)
		}

		if err := task.store.WriteDashboard(langDashboard); err != nil {
			return fmt.Errorf("write dashboard for lang code %s: %w", langCode, err)
		}
	}

	langIndex, err := task.buildLangIndex()
	if err != nil {
		return fmt.Errorf("build dashboard language index: %w", err)
	}

	if err := task.store.WriteDashboardIndex(langIndex); err != nil {
		return fmt.Errorf("write dashboard language index: %w", err)
	}

	return nil
}

func (task *RefreshDashboardTask) buildDashboard(
	ctx context.Context,
	langCode string,
) (dashboard.Dashboard, error) {
	pairs, err := task.pairProviders.ListPairs(langCode)
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

		fileInfo, err := task.gitSeeker.CheckLang(ctx, langCode, gitseek.Pair{
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

	prIndex, err := task.filePRIndex.LangIndex(langCode)
	if err != nil {
		return dashboard.Dashboard{}, fmt.Errorf(
			"get pull request index for lang code %s: %w",
			langCode,
			err,
		)
	}

	return dashboard.BuildDashboard(langCode, seekerFileInfos, prIndex), nil
}

func (task *RefreshDashboardTask) buildLangIndex() (dashboard.LangIndex, error) {
	langIndex, err := dashboard.BuildLangIndex(task.langCodesProvider)
	if err != nil {
		return dashboard.LangIndex{}, fmt.Errorf("build lang index: %w", err)
	}

	return langIndex, nil
}
