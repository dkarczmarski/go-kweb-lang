package tasks

import (
	"fmt"

	"go-kweb-lang/gitcache"
	"go-kweb-lang/seek"
	"go-kweb-lang/web"
)

type RefreshTemplateDataTask struct {
	gitRepoCache *gitcache.GitRepoCache
	templateData *web.TemplateData
}

func NewRefreshTemplateDataTask(
	gitRepoCache *gitcache.GitRepoCache,
	templateData *web.TemplateData,
) *RefreshTemplateDataTask {
	return &RefreshTemplateDataTask{
		gitRepoCache: gitRepoCache,
		templateData: templateData,
	}
}

func (t *RefreshTemplateDataTask) Run() error {
	if err := t.gitRepoCache.PullRefresh(); err != nil {
		return fmt.Errorf("git cache pull refresh error: %w", err)
	}

	if err := t.refreshModel("pl"); err != nil {
		return err
	}

	return nil
}

func (t *RefreshTemplateDataTask) refreshModel(langCode string) error {
	seeker := seek.NewGitLangSeeker(t.gitRepoCache)
	fileInfos, err := seeker.CheckLang(langCode)
	if err != nil {
		return fmt.Errorf("error while checking the content directory for the language code %s: %w", langCode, err)
	}

	model := web.BuildTableModel(fileInfos)
	t.templateData.Set(langCode, model)

	return nil
}
