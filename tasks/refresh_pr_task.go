package tasks

import (
	"context"
	"fmt"

	"github.com/dkarczmarski/go-kweb-lang/langcnt"
	"github.com/dkarczmarski/go-kweb-lang/pullreq"
)

type RefreshPRTask struct {
	filePRIndex       *pullreq.FilePRIndex
	langCodesProvider *langcnt.LangCodesProvider
}

func NewRefreshPRTask(
	filePRIndex *pullreq.FilePRIndex,
	langCodesProvider *langcnt.LangCodesProvider,
) *RefreshPRTask {
	return &RefreshPRTask{
		filePRIndex:       filePRIndex,
		langCodesProvider: langCodesProvider,
	}
}

func (t *RefreshPRTask) Run(ctx context.Context, langCode string) error {
	err := t.filePRIndex.RefreshIndex(ctx, langCode)
	if err != nil {
		return fmt.Errorf("error while updating PRs for lang %v: %w", langCode, err)
	}

	return nil
}
