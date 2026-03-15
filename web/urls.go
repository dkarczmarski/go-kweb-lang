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

	addItemsTypesToQuery(queryValues, params.ItemsTypes)
	addFilenameToQuery(queryValues, params.Filename)
	addFilepathToQuery(queryValues, params.Filepath)
	addSortByToQuery(queryValues, params.SortBy)
	addSortOrderToQuery(queryValues, params.SortOrder)

	encodedQuery := queryValues.Encode()
	if encodedQuery == "" {
		return builder.Path
	}

	return builder.Path + "?" + encodedQuery
}

func addItemsTypesToQuery(queryValues url.Values, itemsTypes []string) {
	if len(itemsTypes) == 0 || isDefaultItemsTypes(itemsTypes) {
		return
	}

	for _, itemsType := range itemsTypes {
		queryValues.Add("itemsType", itemsType)
	}
}

func addFilenameToQuery(queryValues url.Values, filename string) {
	if filename == "" {
		return
	}

	queryValues.Set("filename", filename)
}

func addFilepathToQuery(queryValues url.Values, filepath string) {
	if filepath == "" {
		return
	}

	queryValues.Set("filepath", filepath)
}

func addSortByToQuery(queryValues url.Values, sortBy string) {
	if sortBy == "" || sortBy == SortByFilename {
		return
	}

	queryValues.Set("sort", sortBy)
}

func addSortOrderToQuery(queryValues url.Values, sortOrder string) {
	if sortOrder == "" || sortOrder == SortOrderAsc {
		return
	}

	queryValues.Set("order", sortOrder)
}

func toggleSortOrder(order string) string {
	if order == SortOrderAsc {
		return SortOrderDesc
	}

	return SortOrderAsc
}
