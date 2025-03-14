package tasks

import (
	"fmt"
	"go-kweb-lang/gitcache"
	"go-kweb-lang/seek"
	"go-kweb-lang/web"
	"log"
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

	seeker := seek.NewGitLangSeeker(t.gitRepoCache)
	fileInfos, err := seeker.CheckLang("pl")
	if err != nil {
		log.Fatal(err)
	}

	model := web.BuildTableModel(fileInfos)
	t.templateData.Set(model)

	return nil
}
