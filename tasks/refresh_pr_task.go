package tasks

import (
	"context"
	"fmt"

	"go-kweb-lang/langcnt"
	"go-kweb-lang/pullreq"
)

type RefreshPRTask struct {
	filePRFinder *pullreq.FilePRFinder
	content      *langcnt.Content
}

func NewRefreshPRTask(
	filePRFinder *pullreq.FilePRFinder,
	content *langcnt.Content,
) *RefreshPRTask {
	return &RefreshPRTask{
		filePRFinder: filePRFinder,
		content:      content,
	}
}

func (t *RefreshPRTask) Run(ctx context.Context, langCode string) error {
	err := t.filePRFinder.Update(ctx, langCode)
	if err != nil {
		return fmt.Errorf("error while updating PRs for lang %v: %w", langCode, err)
	}

	return nil
}

func (t *RefreshPRTask) RunAll(ctx context.Context) error {
	langs, err := t.content.Langs()
	if err != nil {
		return fmt.Errorf("error while getting available languages: %w", err)
	}

	for _, langCode := range langs {
		err := t.Run(ctx, langCode)
		if err != nil {
			return fmt.Errorf("error while updating PRs for lang %v: %w", langCode, err)
		}

		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
		}
	}

	return nil
}
