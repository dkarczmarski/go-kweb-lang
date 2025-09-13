package dashboard

import (
	"path/filepath"

	"go-kweb-lang/gitseek"
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

	return &Dashboard{
		LangCode: langCode,
		Items:    items,
	}
}
