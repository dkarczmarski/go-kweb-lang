package web

import (
	"net/url"
	"slices"
	"strings"
)

const (
	ItemsTypeWithEnUpdates        = "with-en-updates"
	ItemsTypeWithPR               = "with-pr"
	ItemsTypeEnFileDoesNotExist   = "en-file-does-not-exist"
	ItemsTypeEnFileNoLongerExists = "en-file-no-longer-exists"
	ItemsTypeLangFileMissing      = "lang-file-missing"
	ItemsTypeWaitingForReview     = "waiting-for-review"
	ItemsTypeLangFileUpToDate     = "up-to-date"
)

const (
	SortByFilename = "filename"
	SortByStatus   = "status"
	SortByUpdates  = "updates"
)

const (
	SortOrderAsc  = "asc"
	SortOrderDesc = "desc"
)

type LangDashboardParams struct {
	LangCode   string
	ItemsTypes []string
	Filename   string
	Filepath   string
	SortBy     string
	SortOrder  string
}

func ParseLangDashboardParams(langCode string, values url.Values) LangDashboardParams {
	params := LangDashboardParams{
		LangCode:   strings.TrimSpace(langCode),
		ItemsTypes: normalizeItemsTypes(values["itemsType"]),
		Filename:   strings.TrimSpace(values.Get("filename")),
		Filepath:   strings.TrimSpace(values.Get("filepath")),
		SortBy:     normalizeSortBy(values.Get("sort")),
		SortOrder:  normalizeSortOrder(values.Get("order")),
	}

	return params
}

func normalizeItemsTypes(values []string) []string {
	normalized := make([]string, 0, len(values))

	for _, value := range values {
		switch strings.TrimSpace(value) {
		case ItemsTypeWithEnUpdates:
			normalized = appendIfMissing(normalized, ItemsTypeWithEnUpdates)
		case ItemsTypeWithPR:
			normalized = appendIfMissing(normalized, ItemsTypeWithPR)
		case ItemsTypeEnFileDoesNotExist:
			normalized = appendIfMissing(normalized, ItemsTypeEnFileDoesNotExist)
		case ItemsTypeEnFileNoLongerExists:
			normalized = appendIfMissing(normalized, ItemsTypeEnFileNoLongerExists)
		case ItemsTypeLangFileMissing:
			normalized = appendIfMissing(normalized, ItemsTypeLangFileMissing)
		case ItemsTypeWaitingForReview:
			normalized = appendIfMissing(normalized, ItemsTypeWaitingForReview)
		case ItemsTypeLangFileUpToDate:
			normalized = appendIfMissing(normalized, ItemsTypeLangFileUpToDate)
		}
	}

	if len(normalized) == 0 {
		return defaultItemsTypes()
	}

	return normalized
}

func defaultItemsTypes() []string {
	return []string{
		ItemsTypeWithEnUpdates,
		ItemsTypeWithPR,
		ItemsTypeEnFileDoesNotExist,
		ItemsTypeEnFileNoLongerExists,
		ItemsTypeWaitingForReview,
	}
}

func isDefaultItemsTypes(itemsTypes []string) bool {
	return slices.Equal(itemsTypes, defaultItemsTypes())
}

func hasItemsType(itemsTypes []string, itemsType string) bool {
	return slices.Contains(itemsTypes, itemsType)
}

func appendIfMissing(values []string, value string) []string {
	if slices.Contains(values, value) {
		return values
	}

	return append(values, value)
}

func normalizeSortBy(value string) string {
	switch strings.TrimSpace(value) {
	case SortByStatus:
		return SortByStatus
	case SortByUpdates:
		return SortByUpdates
	default:
		return SortByFilename
	}
}

func normalizeSortOrder(value string) string {
	if strings.TrimSpace(value) == SortOrderDesc {
		return SortOrderDesc
	}

	return SortOrderAsc
}
