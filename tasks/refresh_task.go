package tasks

import (
	"context"

	"go-kweb-lang/web"
)

type RefreshTask struct {
	refreshRepoTask         *RefreshRepoTask
	refreshPRTask           *RefreshPRTask
	refreshTemplateDataTask *web.RefreshTemplateDataTask
}

func NewRefreshTask(
	refreshRepoTask *RefreshRepoTask,
	refreshPRTask *RefreshPRTask,
	refreshTemplateDataTask *web.RefreshTemplateDataTask,
) *RefreshTask {
	return &RefreshTask{
		refreshRepoTask:         refreshRepoTask,
		refreshPRTask:           refreshPRTask,
		refreshTemplateDataTask: refreshTemplateDataTask,
	}
}

func (t *RefreshTask) OnUpdate(ctx context.Context, repoUpdated bool, prUpdatedLangCodes []string) error {
	if repoUpdated {
		if err := t.refreshRepoTask.Run(ctx); err != nil {
			return err
		}
	}

	for _, langCode := range prUpdatedLangCodes {
		if err := t.refreshPRTask.Run(ctx, langCode); err != nil {
			return err
		}
	}

	if err := t.refreshTemplateDataTask.Run(ctx); err != nil {
		return err
	}

	return nil
}
