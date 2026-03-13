package dashboard

import (
	"github.com/dkarczmarski/go-kweb-lang/gitseek"
)

const (
	StatusWaitingForReview = "waiting-for-review"
)

func buildDashboard(
	langCode string,
	seekerFileInfos []gitseek.FileInfo,
	prIndex map[string][]int,
) *Dashboard {
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
					LangPath:   prFilePath,
					FileStatus: StatusWaitingForReview,
				},
				PRs: prs,
			})
		}
	}

	return &Dashboard{
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
