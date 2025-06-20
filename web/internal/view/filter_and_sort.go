package view

import (
	"cmp"
	"slices"
	"strings"
)

const (
	ItemsAll = iota
	ItemsWithUpdate
	ItemsWithUpdateOrPR
	ItemsWithPR
)

const (
	SortByFileName = iota
	SortByStatus
	SortByEnUpdate
)

const (
	SortOrderAsc = iota
	SortOrderDesc
)

func FilterAndSort(
	files []FileModel,
	itemsType int,
	fileName string,
	filePath string,
	sort int,
	sortOrder int,
) []FileModel {
	var newFiles []FileModel

	if len(fileName) > 0 {
		for _, f := range files {
			if f.LangRelPath.Text == fileName {
				newFiles = append(newFiles, f)
				break
			}
		}
	} else {
		for _, f := range files {
			if len(filePath) > 0 && !strings.Contains(f.LangRelPath.Text, filePath) {
				continue
			}

			switch itemsType {
			case ItemsAll:
				break
			case ItemsWithUpdate:
				if !hasAnyENUpdate(f) {
					continue
				}
			case ItemsWithUpdateOrPR:
				if !hasAnyENUpdate(f) && !hasAnyPR(f) {
					continue
				}
			case ItemsWithPR:
				if !hasAnyPR(f) {
					continue
				}
			default:
				continue
			}

			newFiles = append(newFiles, f)
		}

		slices.SortFunc(newFiles, func(a, b FileModel) int {
			var cmpValue int

			switch sort {
			case SortByFileName:
				cmpValue = cmp.Compare(a.LangRelPath.Text, b.LangRelPath.Text)
			case SortByStatus:
				cmpValue = cmp.Compare(a.ENStatus, b.ENStatus)
			case SortByEnUpdate:
				cmpValue = cmp.Compare(a.ENUpdates.LastCommit.DateTime, b.ENUpdates.LastCommit.DateTime)
			}

			switch sortOrder {
			case SortOrderAsc:
				break
			case SortOrderDesc:
				cmpValue *= -1
			}

			return cmpValue
		})
	}

	return newFiles
}

func hasAnyENUpdate(f FileModel) bool {
	u := f.ENUpdates
	return len(u.AfterMergeCommit) > 0 ||
		len(u.AfterLastCommit) > 0 ||
		len(u.AfterForkCommit) > 0 ||
		len(u.BeforeForkCommit) > 0
}

func hasAnyPR(f FileModel) bool {
	return len(f.PRs) > 0
}
