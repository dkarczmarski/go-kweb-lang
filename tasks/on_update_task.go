package tasks

import (
	"context"
	"fmt"
)

type OnGitHubUpdateTask struct {
	refreshRepoTask      *RefreshRepoTask
	refreshPRTask        *RefreshPRTask
	refreshDashboardTask *RefreshDashboardTask
}

func NewOnGitHubUpdateTask(
	refreshRepoTask *RefreshRepoTask,
	refreshPRTask *RefreshPRTask,
	refreshDashboardTask *RefreshDashboardTask,
) *OnGitHubUpdateTask {
	return &OnGitHubUpdateTask{
		refreshRepoTask:      refreshRepoTask,
		refreshPRTask:        refreshPRTask,
		refreshDashboardTask: refreshDashboardTask,
	}
}

func (t *OnGitHubUpdateTask) OnUpdate(
	ctx context.Context,
	repoUpdated bool,
	changedLangCodesInPR []string,
) error {
	uniqueLangCodes := uniqueStrings(changedLangCodesInPR)

	if repoUpdated {
		if err := t.refreshRepoTask.Run(ctx); err != nil {
			return fmt.Errorf("run refresh repo task: %w", err)
		}
	}

	for _, langCode := range uniqueLangCodes {
		if err := t.refreshPRTask.Run(ctx, langCode); err != nil {
			return fmt.Errorf("run refresh PR task for lang code %s: %w", langCode, err)
		}
	}

	if repoUpdated || len(uniqueLangCodes) > 0 {
		if err := t.refreshDashboardTask.Run(ctx); err != nil {
			return fmt.Errorf("run refresh dashboard task: %w", err)
		}
	}

	return nil
}

func uniqueStrings(values []string) []string {
	seen := make(map[string]struct{}, len(values))
	result := make([]string, 0, len(values))

	for _, value := range values {
		if _, exists := seen[value]; exists {
			continue
		}

		seen[value] = struct{}{}

		result = append(result, value)
	}

	return result
}
