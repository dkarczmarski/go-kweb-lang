package web

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
	model *LangDashboardViewModel,
	filter int,
	fileName string,
	filePath string,
	sort int,
	sortOrder int,
) {
	var newFiles []FileModel

	if len(fileName) > 0 {
		for _, f := range model.TableModel.Files {
			if f.LangRelPath.Text == fileName {
				newFiles = append(newFiles, f)
				break
			}
		}
	} else {
		for _, f := range model.TableModel.Files {
			if len(filePath) > 0 && !strings.Contains(f.LangRelPath.Text, filePath) {
				continue
			}

			switch filter {
			case ItemsAll:
				break
			case ItemsWithUpdate:
				if !(len(f.ENUpdates.AfterMergeCommit) > 0 ||
					len(f.ENUpdates.AfterLastCommit) > 0 ||
					len(f.ENUpdates.AfterForkCommit) > 0 ||
					len(f.ENUpdates.BeforeForkCommit) > 0) {
					continue
				}
			case ItemsWithUpdateOrPR:
				if !(len(f.ENUpdates.AfterMergeCommit) > 0 ||
					len(f.ENUpdates.AfterLastCommit) > 0 ||
					len(f.ENUpdates.AfterForkCommit) > 0 ||
					len(f.ENUpdates.BeforeForkCommit) > 0 ||
					len(f.PRs) > 0) {
					continue
				}
			case ItemsWithPR:
				if !(len(f.PRs) > 0) {
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

	model.TableModel.Files = newFiles
}
