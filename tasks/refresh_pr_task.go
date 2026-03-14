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
	if err := t.filePRIndex.RefreshIndex(ctx, langCode); err != nil {
		return fmt.Errorf("refresh PR index for lang code %s: %w", langCode, err)
	}

	return nil
}
