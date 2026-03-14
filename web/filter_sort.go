package web

import (
	"cmp"
	"slices"
	"strings"

	"github.com/dkarczmarski/go-kweb-lang/dashboard"
)

func FilterAndSortItems(items []dashboard.Item, params LangDashboardParams) []dashboard.Item {
	filteredItems := filterItems(items, params)
	sortItems(filteredItems, params)

	return filteredItems
}

func filterItems(items []dashboard.Item, params LangDashboardParams) []dashboard.Item {
	if params.Filename != "" {
		return filterItemsByFilename(items, params.Filename)
	}

	result := make([]dashboard.Item, 0, len(items))

	for _, item := range items {
		if !matchesFilepath(item, params.Filepath) {
			continue
		}

		if !matchesItemsType(item, params.ItemsType) {
			continue
		}

		result = append(result, item)
	}

	return result
}

func filterItemsByFilename(items []dashboard.Item, filename string) []dashboard.Item {
	result := make([]dashboard.Item, 0, 1)

	for _, item := range items {
		if item.LangPath != filename {
			continue
		}

		result = append(result, item)

		break
	}

	return result
}

func matchesFilepath(item dashboard.Item, filepath string) bool {
	if filepath == "" {
		return true
	}

	return strings.Contains(item.LangPath, filepath)
}

func matchesItemsType(item dashboard.Item, itemsType string) bool {
	switch itemsType {
	case ItemsTypeWithUpdate:
		return len(item.EnUpdates) > 0
	case ItemsTypeWithUpdateOrPR:
		return len(item.EnUpdates) > 0 || len(item.PRs) > 0
	case ItemsTypeWithPR:
		return len(item.PRs) > 0
	case ItemsTypeAll:
		return true
	default:
		return true
	}
}

func sortItems(items []dashboard.Item, params LangDashboardParams) {
	slices.SortFunc(items, func(leftItem dashboard.Item, rightItem dashboard.Item) int {
		var comparisonResult int

		switch params.SortBy {
		case SortByStatus:
			comparisonResult = cmp.Compare(leftItem.FileStatus, rightItem.FileStatus)
		case SortByUpdates:
			leftDate := latestEnUpdateDate(leftItem)
			rightDate := latestEnUpdateDate(rightItem)
			comparisonResult = cmp.Compare(leftDate, rightDate)
		default:
			comparisonResult = cmp.Compare(leftItem.LangPath, rightItem.LangPath)
		}

		if params.SortOrder == SortOrderDesc {
			comparisonResult *= -1
		}

		return comparisonResult
	})
}

func latestEnUpdateDate(item dashboard.Item) string {
	if len(item.EnUpdates) == 0 {
		return ""
	}

	latestDate := item.EnUpdates[0].Commit.DateTime
	for _, update := range item.EnUpdates {
		if update.Commit.DateTime > latestDate {
			latestDate = update.Commit.DateTime
		}
	}

	return latestDate
}
