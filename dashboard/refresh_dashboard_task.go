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
	filePRFinder      *pullreq.FilePRFinder
	store             *Store
}

func NewRefreshTask(
	langCodesProvider *langcnt.LangCodesProvider,
	pairProviders *filepairs.PairProviders,
	gitSeeker *gitseek.GitSeek,
	filePRFinder *pullreq.FilePRFinder,
	store *Store,
) *RefreshTask {
	return &RefreshTask{
		langCodesProvider: langCodesProvider,
		pairProviders:     pairProviders,
		gitSeeker:         gitSeeker,
		filePRFinder:      filePRFinder,
		store:             store,
	}
}

func (t *RefreshTask) Run(ctx context.Context) error {
	langCodes, err := t.langCodesProvider.LangCodes()
	if err != nil {
		return fmt.Errorf("failed to get available languages: %w", err)
	}

	for _, langCode := range langCodes {
		langDashboard, err := t.buildDashboard(ctx, langCode)
		if err != nil {
			return err
		}

		if err := t.store.WriteDashboard(langDashboard); err != nil {
			return fmt.Errorf("failed to store language dashboard: %w", err)
		}
	}

	langIndex, err := t.buildLangIndex()
	if err != nil {
		return err
	}

	if err := t.store.WriteDashboardIndex(langIndex); err != nil {
		return fmt.Errorf("failed to store dashboard index: %w", err)
	}

	return nil
}

func (t *RefreshTask) buildDashboard(ctx context.Context, langCode string) (*Dashboard, error) {
	pairs, err := t.pairProviders.ListPairs(langCode)
	if err != nil {
		return nil, fmt.Errorf("error while listing file pairs for lang code %s: %w", langCode, err)
	}

	seekerFileInfos := make([]gitseek.FileInfo, 0, len(pairs))
	for i, pair := range pairs {
		log.Printf("[gitseek][%s][%d/%d] checking for updates for %v", langCode, i+1, len(pairs), pair.LangPath)
		fileInfo, err := t.gitSeeker.CheckLang(ctx, langCode, gitseek.Pair{
			EnPath:   pair.EnPath,
			LangPath: pair.LangPath,
		})
		if err != nil {
			return nil, fmt.Errorf("error while checking file pair %s for the language code %s: %w", pair.LangPath, langCode, err)
		}
		seekerFileInfos = append(seekerFileInfos, fileInfo)
	}

	prIndex, err := t.filePRFinder.LangIndex(langCode)
	if err != nil {
		return nil, fmt.Errorf("error while getting pull request index for lang code %s: %w", langCode, err)
	}

	return buildDashboard(langCode, seekerFileInfos, prIndex), nil
}

func (t *RefreshTask) buildLangIndex() (*LangIndex, error) {
	return buildLangIndex(t.langCodesProvider)
}
