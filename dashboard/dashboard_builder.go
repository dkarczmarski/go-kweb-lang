package dashboard

import (
	"github.com/dkarczmarski/go-kweb-lang/git"
	"github.com/dkarczmarski/go-kweb-lang/gitseek"
	"github.com/dkarczmarski/go-kweb-lang/pullreq"
)

const (
	StatusWaitingForReview = "waiting-for-review"
)

func BuildDashboard(
	langCode string,
	seekerFileInfos []gitseek.FileInfo,
	prIndex pullreq.FilePRIndexData,
) Dashboard {
	items := make([]Item, 0, len(seekerFileInfos))

	for _, seekerFileInfo := range seekerFileInfos {
		prs := prIndex[seekerFileInfo.LangPath]

		item := Item{
			FileInfo: seekerFileInfo,
			PRs:      prs,
		}

		items = append(items, item)
	}

	for prFilePath, prs := range prIndex {
		if !containsItem(items, prFilePath) {
			items = append(items, Item{
				FileInfo: gitseek.FileInfo{
					LangPath:        prFilePath,
					FileStatus:      StatusWaitingForReview,
					LangLastCommit:  git.CommitInfo{}, //nolint:exhaustruct
					LangMergeCommit: nil,
					LangForkCommit:  nil,
					EnUpdates:       nil,
				},
				PRs: prs,
			})
		}
	}

	return Dashboard{
		LangCode: langCode,
		Items:    items,
	}
}

func containsItem(items []Item, fileRelPath string) bool {
	for _, item := range items {
		if item.LangPath == fileRelPath {
			return true
		}
	}

	return false
}
