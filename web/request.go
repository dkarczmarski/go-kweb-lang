package web

import (
	"net/url"
	"strings"
)

const (
	ItemsTypeAll            = "all"
	ItemsTypeWithUpdate     = "with-update"
	ItemsTypeWithUpdateOrPR = "with-update-or-pr"
	ItemsTypeWithPR         = "with-pr"
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
	LangCode  string
	ItemsType string
	Filename  string
	Filepath  string
	SortBy    string
	SortOrder string
}

func ParseLangDashboardParams(langCode string, values url.Values) LangDashboardParams {
	params := LangDashboardParams{
		LangCode:  strings.TrimSpace(langCode),
		ItemsType: normalizeItemsType(values.Get("itemsType")),
		Filename:  strings.TrimSpace(values.Get("filename")),
		Filepath:  strings.TrimSpace(values.Get("filepath")),
		SortBy:    normalizeSortBy(values.Get("sort")),
		SortOrder: normalizeSortOrder(values.Get("order")),
	}

	return params
}

func normalizeItemsType(value string) string {
	switch strings.TrimSpace(value) {
	case ItemsTypeWithUpdate:
		return ItemsTypeWithUpdate
	case ItemsTypeWithUpdateOrPR:
		return ItemsTypeWithUpdateOrPR
	case ItemsTypeWithPR:
		return ItemsTypeWithPR
	default:
		return ItemsTypeAll
	}
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
