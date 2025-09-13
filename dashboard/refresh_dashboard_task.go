package dashboard

import (
	"context"
	"fmt"

	"go-kweb-lang/gitseek"
	"go-kweb-lang/langcnt"
	"go-kweb-lang/pullreq"
)

type RefreshTask struct {
	langCodesProvider *langcnt.LangCodesProvider
	gitSeeker         *gitseek.GitSeek
	filePRFinder      *pullreq.FilePRFinder
	store             *Store
}

func NewRefreshTask(
	langCodesProvider *langcnt.LangCodesProvider,
	gitSeeker *gitseek.GitSeek,
	filePRFinder *pullreq.FilePRFinder,
	store *Store,
) *RefreshTask {
	return &RefreshTask{
		langCodesProvider: langCodesProvider,
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
	seekerFileInfos, err := t.gitSeeker.CheckLang(ctx, langCode)
	if err != nil {
		return nil, fmt.Errorf("error while checking the content directory for the language code %s: %w", langCode, err)
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
