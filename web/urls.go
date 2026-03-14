package web

import "net/url"

type DashboardURLBuilder struct {
	Path   string
	Params LangDashboardParams
}

func NewDashboardURLBuilder(path string, params LangDashboardParams) DashboardURLBuilder {
	return DashboardURLBuilder{
		Path:   path,
		Params: params,
	}
}

func (builder DashboardURLBuilder) Current() string {
	return builder.build(builder.Params)
}

func (builder DashboardURLBuilder) WithItemsType(itemsType string) string {
	params := builder.Params
	params.ItemsType = normalizeItemsType(itemsType)

	return builder.build(params)
}

func (builder DashboardURLBuilder) WithFilename(filename string) string {
	params := builder.Params
	params.Filename = filename

	return builder.build(params)
}

func (builder DashboardURLBuilder) WithoutFilename() string {
	params := builder.Params
	params.Filename = ""

	return builder.build(params)
}

func (builder DashboardURLBuilder) WithFilepath(filepath string) string {
	params := builder.Params
	params.Filepath = filepath

	return builder.build(params)
}

func (builder DashboardURLBuilder) Sort(sortBy string) string {
	params := builder.Params
	if params.SortBy == sortBy {
		params.SortOrder = toggleSortOrder(params.SortOrder)
	} else {
		params.SortBy = normalizeSortBy(sortBy)
		params.SortOrder = SortOrderAsc
	}

	return builder.build(params)
}

func (builder DashboardURLBuilder) build(params LangDashboardParams) string {
	queryValues := url.Values{}

	if params.ItemsType != "" && params.ItemsType != ItemsTypeAll {
		queryValues.Set("itemsType", params.ItemsType)
	}

	if params.Filename != "" {
		queryValues.Set("filename", params.Filename)
	}

	if params.Filepath != "" {
		queryValues.Set("filepath", params.Filepath)
	}

	if params.SortBy != "" && params.SortBy != SortByFilename {
		queryValues.Set("sort", params.SortBy)
	}

	if params.SortOrder != "" && params.SortOrder != SortOrderAsc {
		queryValues.Set("order", params.SortOrder)
	}

	encodedQuery := queryValues.Encode()
	if encodedQuery == "" {
		return builder.Path
	}

	return builder.Path + "?" + encodedQuery
}

func toggleSortOrder(order string) string {
	if order == SortOrderAsc {
		return SortOrderDesc
	}

	return SortOrderAsc
}
