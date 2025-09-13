package dashboard

import (
	"os"
	"path/filepath"
	"strings"

	"go-kweb-lang/gitseek"
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
		file := filepath.Join("content", langCode, seekerFileInfo.LangRelPath)
		prs := prIndex[file]

		item := Item{
			FileInfo: seekerFileInfo,
			PRs:      prs,
		}

		items = append(items, item)
	}

	contentPath := filepath.Join("content", langCode) + string(os.PathSeparator)

	for prFilePath, prs := range prIndex {
		fileRelPath := strings.TrimPrefix(prFilePath, contentPath)
		if !containsItem(items, fileRelPath) {
			items = append(items, Item{
				FileInfo: gitseek.FileInfo{
					LangRelPath: fileRelPath,
					FileStatus:  StatusWaitingForReview,
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
		if item.LangRelPath == fileRelPath {
			return true
		}
	}

	return false
}
