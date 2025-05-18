package tasks

import (
	"context"
	"fmt"
	"path/filepath"

	"go-kweb-lang/gitseek"
	"go-kweb-lang/langcnt"
	"go-kweb-lang/pullreq"
	"go-kweb-lang/web"
)

type RefreshTemplateDataTask struct {
	langCodesProvider *langcnt.LangCodesProvider
	gitSeeker         *gitseek.GitSeek
	filePRFinder      *pullreq.FilePRFinder
	templateData      *web.TemplateData
}

func NewRefreshTemplateDataTask(
	langCodesProvider *langcnt.LangCodesProvider,
	gitSeeker *gitseek.GitSeek,
	filePRFinder *pullreq.FilePRFinder,
	templateData *web.TemplateData,
) *RefreshTemplateDataTask {
	return &RefreshTemplateDataTask{
		langCodesProvider: langCodesProvider,
		gitSeeker:         gitSeeker,
		filePRFinder:      filePRFinder,
		templateData:      templateData,
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

	indexModel, err := web.BuildIndexModel(t.langCodesProvider)
	if err != nil {
		return fmt.Errorf("error while building index web model: %w", err)
	}

	t.templateData.SetIndex(indexModel)

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

	var fileInfos []web.FileInfo
	for _, seekerFileInfo := range seekerFileInfos {
		file := filepath.Join("content", langCode, seekerFileInfo.LangRelPath)
		prs := prIndex[file]

		fileInfo := web.FileInfo{
			FileInfo: seekerFileInfo,
			PRs:      prs,
		}

		fileInfos = append(fileInfos, fileInfo)
	}

	model := web.BuildLangModel(fileInfos)
	t.templateData.SetLang(langCode, model)

	return nil
}
