package tasks

import (
	"context"

	"go-kweb-lang/dashboard"
)

type RefreshTask struct {
	refreshRepoTask      *RefreshRepoTask
	refreshPRTask        *RefreshPRTask
	refreshDashboardTask *dashboard.RefreshTask
}

func NewRefreshTask(
	refreshRepoTask *RefreshRepoTask,
	refreshPRTask *RefreshPRTask,
	refreshDashboardTask *dashboard.RefreshTask,
) *RefreshTask {
	return &RefreshTask{
		refreshRepoTask:      refreshRepoTask,
		refreshPRTask:        refreshPRTask,
		refreshDashboardTask: refreshDashboardTask,
	}
}

func (t *RefreshTask) OnUpdate(ctx context.Context, repoUpdated bool, changedLangCodesInPR []string) error {
	if repoUpdated {
		if err := t.refreshRepoTask.Run(ctx); err != nil {
			return err
		}
	}

	for _, langCode := range changedLangCodesInPR {
		if err := t.refreshPRTask.Run(ctx, langCode); err != nil {
			return err
		}
	}

	if err := t.refreshDashboardTask.Run(ctx); err != nil {
		return err
	}

	return nil
}
