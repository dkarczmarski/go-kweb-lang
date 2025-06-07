package web

import (
	"context"
	"fmt"
	"path/filepath"

	"go-kweb-lang/gitseek"
	"go-kweb-lang/langcnt"
	"go-kweb-lang/pullreq"
)

type RefreshTemplateDataTask struct {
	langCodesProvider *langcnt.LangCodesProvider
	gitSeeker         *gitseek.GitSeek
	filePRFinder      *pullreq.FilePRFinder
	viewModelStore    ViewModelStore
}

func NewRefreshTemplateDataTask(
	langCodesProvider *langcnt.LangCodesProvider,
	gitSeeker *gitseek.GitSeek,
	filePRFinder *pullreq.FilePRFinder,
	viewModelStore ViewModelStore,
) *RefreshTemplateDataTask {
	return &RefreshTemplateDataTask{
		langCodesProvider: langCodesProvider,
		gitSeeker:         gitSeeker,
		filePRFinder:      filePRFinder,
		viewModelStore:    viewModelStore,
	}
}

func (t *RefreshTemplateDataTask) Run(ctx context.Context) error {
	langCodes, err := t.langCodesProvider.LangCodes()
	if err != nil {
		return fmt.Errorf("error while getting available languages: %w", err)
	}

	for _, langCode := range langCodes {
		if err := t.refreshLangModel(ctx, langCode); err != nil {
			return err
		}
	}

	indexModel, err := BuildIndexModel(t.langCodesProvider)
	if err != nil {
		return fmt.Errorf("error while building index web model: %w", err)
	}

	langCodesViewModel := &LangCodesViewModel{
		LangCodes: indexModel,
	}
	if err := t.viewModelStore.SetLangCodes(langCodesViewModel); err != nil {
		return fmt.Errorf("failed to store language codes view model: %w", err)
	}

	return nil
}

func (t *RefreshTemplateDataTask) refreshLangModel(ctx context.Context, langCode string) error {
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

	model := BuildLangModel(fileInfos)
	langDashboardViewModel := &LangDashboardViewModel{
		TableModel: *model, // todo:
	}

	if err := t.viewModelStore.SetLangDashboard(langCode, langDashboardViewModel); err != nil {
		return fmt.Errorf("failed to store language dashboard view model: %w", err)
	}

	return nil
}
