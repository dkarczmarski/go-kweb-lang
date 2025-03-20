package tasks

import (
	"context"
	"fmt"
	"go-kweb-lang/gitcache"
	"go-kweb-lang/langcnt"
	"go-kweb-lang/seek"
	"go-kweb-lang/web"
)

type RefreshTemplateDataTask struct {
	content      *langcnt.Content
	gitRepoCache *gitcache.GitRepoCache
	templateData *web.TemplateData
}

func NewRefreshTemplateDataTask(
	content *langcnt.Content,
	gitRepoCache *gitcache.GitRepoCache,
	templateData *web.TemplateData,
) *RefreshTemplateDataTask {
	return &RefreshTemplateDataTask{
		content:      content,
		gitRepoCache: gitRepoCache,
		templateData: templateData,
	}
}

func (t *RefreshTemplateDataTask) Run(ctx context.Context) error {
	langs, err := t.content.Langs()
	if err != nil {
		return fmt.Errorf("error while getting available languages: %w", err)
	}

	indexModel, err := web.BuildIndexModel(t.content)
	if err != nil {
		return fmt.Errorf("error while building index web model: %w", err)
	}

	t.templateData.SetIndex(indexModel)

	for _, lang := range langs {
		if err := t.refreshModel(ctx, lang); err != nil {
			return err
		}
	}

	return nil
}

func (t *RefreshTemplateDataTask) refreshModel(ctx context.Context, langCode string) error {
	seeker := seek.NewGitLangSeeker(t.gitRepoCache)
	fileInfos, err := seeker.CheckLang(ctx, langCode)
	if err != nil {
		return fmt.Errorf("error while checking the content directory for the language code %s: %w", langCode, err)
	}

	model := web.BuildLangModel(fileInfos)
	t.templateData.SetLang(langCode, model)

	return nil
}
