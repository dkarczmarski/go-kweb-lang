package tasks

import (
	"context"
	"fmt"

	"github.com/dkarczmarski/go-kweb-lang/langcnt"
	"github.com/dkarczmarski/go-kweb-lang/pullreq"
)

type RefreshPRTask struct {
	filePRFinder      *pullreq.FilePRFinder
	langCodesProvider *langcnt.LangCodesProvider
}

func NewRefreshPRTask(
	filePRFinder *pullreq.FilePRFinder,
	langCodesProvider *langcnt.LangCodesProvider,
) *RefreshPRTask {
	return &RefreshPRTask{
		filePRFinder:      filePRFinder,
		langCodesProvider: langCodesProvider,
	}
}

func (t *RefreshPRTask) Run(ctx context.Context, langCode string) error {
	err := t.filePRFinder.Update(ctx, langCode)
	if err != nil {
		return fmt.Errorf("error while updating PRs for lang %v: %w", langCode, err)
	}

	return nil
}
