package reqhelper

import (
	"errors"
	"fmt"
	"net/http"

	"go-kweb-lang/web/internal/weberror"
)

type RequestModel struct {
	LangCode       string
	ItemsTypeParam string
	FilenameParam  string
	FilepathParam  string
	SortParam      string
	SortOrderParam string
}

func ParseListLangDashboardRequest(r *http.Request) (RequestModel, error) {
	if err := r.ParseForm(); err != nil {
		return RequestModel{}, fmt.Errorf("failed to parse form: %w",
			weberror.NewWebError(http.StatusBadRequest, err),
		)
	}

	langCode := r.PathValue("code")
	if len(langCode) == 0 {
		return RequestModel{}, weberror.NewWebError(http.StatusBadRequest, errors.New("missing lang code"))
	}

	itemsTypeParam := findRequestValue(r, "NewItemsType", "CurrentItemsType", "items-type")
	filenameParam := findRequestValue(r, "", "CurrentFilename", "filename")
	filepathParam := findRequestValue(r, "NewFilepath", "CurrentFilepath", "filepath")
	sortParam := findRequestValue(r, "", "CurrentSort", "sort")
	sortOrderParam := findRequestValue(r, "", "CurrentSortOrder", "sort-order")

	return RequestModel{
		LangCode:       langCode,
		ItemsTypeParam: valueOrDefault(itemsTypeParam, "all"),
		FilenameParam:  filenameParam,
		FilepathParam:  filepathParam,
		SortParam:      valueOrDefault(sortParam, "filename"),
		SortOrderParam: valueOrDefault(sortOrderParam, "asc"),
	}, nil
}

func valueOrDefault(value, defaultValue string) string {
	if len(value) > 0 {
		return value
	}

	return defaultValue
}

func findRequestValue(r *http.Request, newValueKey, currentValueKey, urlKey string) string {
	if newValueKey != "" && r.FormValue(newValueKey) == "" && r.FormValue(currentValueKey) != "" {
		// reset value if the new input is empty but the previous one was not
		return ""
	}
	if v := r.FormValue(newValueKey); v != "" {
		// this value comes from hidden inputs that store previous state.
		// it is empty when the page is opened for the first time.
		return v
	}
	if v := r.FormValue(currentValueKey); v != "" {
		// this value represents a newly set value from user interaction on the page.
		return v
	}
	// if the page is opened for the first time, try to get the value from the URL.
	return r.URL.Query().Get(urlKey)
}
