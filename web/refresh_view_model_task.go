package web

import (
	"context"
	"fmt"
	"path/filepath"

	"go-kweb-lang/gitseek"
	"go-kweb-lang/langcnt"
	"go-kweb-lang/pullreq"
)

type RefreshViewModelTask struct {
	langCodesProvider *langcnt.LangCodesProvider
	gitSeeker         *gitseek.GitSeek
	filePRFinder      *pullreq.FilePRFinder
	viewModelStore    ViewModelStore
}

func NewRefreshViewModelTask(
	langCodesProvider *langcnt.LangCodesProvider,
	gitSeeker *gitseek.GitSeek,
	filePRFinder *pullreq.FilePRFinder,
	viewModelStore ViewModelStore,
) *RefreshViewModelTask {
	return &RefreshViewModelTask{
		langCodesProvider: langCodesProvider,
		gitSeeker:         gitSeeker,
		filePRFinder:      filePRFinder,
		viewModelStore:    viewModelStore,
	}
}

func (t *RefreshViewModelTask) Run(ctx context.Context) error {
	langCodes, err := t.langCodesProvider.LangCodes()
	if err != nil {
		return fmt.Errorf("error while getting available languages: %w", err)
	}

	for _, langCode := range langCodes {
		if err := t.refreshLangModel(ctx, langCode); err != nil {
			return err
		}
	}

	langCodesViewModel, err := buildLangCodesViewModel(t.langCodesProvider)
	if err != nil {
		return fmt.Errorf("failed to build view model: %w", err)
	}

	if err := t.viewModelStore.SetLangCodes(langCodesViewModel); err != nil {
		return fmt.Errorf("failed to store language codes view model: %w", err)
	}

	return nil
}

func (t *RefreshViewModelTask) refreshLangModel(ctx context.Context, langCode string) error {
	seekerFileInfos, err := t.gitSeeker.CheckLang(ctx, langCode)
	if err != nil {
		return fmt.Errorf("error while checking the content directory for the language code %s: %w", langCode, err)
	}

	prIndex, err := t.filePRFinder.LangIndex(langCode)
	if err != nil {
		return fmt.Errorf("error while getting pull request index for lang code %v: %w", langCode, err)
	}

	var fileInfos []FileInfo
	for _, seekerFileInfo := range seekerFileInfos {
		file := filepath.Join("content", langCode, seekerFileInfo.LangRelPath)
		prs := prIndex[file]

		fileInfo := FileInfo{
			FileInfo: seekerFileInfo,
			PRs:      prs,
		}

		fileInfos = append(fileInfos, fileInfo)
	}

	langDashboardViewModel := buildLangDashboardViewModel(langCode, fileInfos)

	if err := t.viewModelStore.SetLangDashboard(langCode, langDashboardViewModel); err != nil {
		return fmt.Errorf("failed to store language dashboard view model: %w", err)
	}

	return nil
}
