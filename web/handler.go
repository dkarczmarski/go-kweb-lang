package web

import (
	_ "embed"
	"errors"
	"fmt"
	"html/template"
	"log"
	"net/http"
	"strings"
)

//go:embed lang_codes.html
var langCodesHTML string

func createListLangCodesHandler(store ViewModelStore) func(w http.ResponseWriter, r *http.Request) {
	tmpl := template.New("lang_codes.html")
	htmlTmpl := template.Must(tmpl.Parse(langCodesHTML))

	return func(w http.ResponseWriter, r *http.Request) {
		model, err := store.GetLangCodes()
		if err != nil {
			log.Printf("failed to get language codes: %v", err)
			http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)

			return
		}

		if err := htmlTmpl.Execute(w, model); err != nil {
			log.Printf("failed to execute template: %v", err)
			http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)

			return
		}
	}
}

//go:embed lang_dashboard.html
var langDashboardHTML string

func createLangDashboardHandler(store ViewModelStore) func(w http.ResponseWriter, r *http.Request) {
	funcMap := template.FuncMap{
		"truncate": truncate,
	}

	tmpl := template.New("lang_dashboard.html").Funcs(funcMap)
	htmlTmpl := template.Must(tmpl.Parse(langDashboardHTML))

	return func(w http.ResponseWriter, r *http.Request) {
		model, err := handleLangDashboardRequest(w, r, store)
		if err != nil {
			log.Printf("failed to prepare view model: %v", err)

			return
		}

		if err := htmlTmpl.Execute(w, model); err != nil {
			log.Printf("failed to execute template: %v", err)
			http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)

			return
		}
	}
}

func createLangDashboardTableHandler(store ViewModelStore) func(w http.ResponseWriter, r *http.Request) {
	funcMap := template.FuncMap{
		"truncate": truncate,
	}

	tmpl := template.New("lang_dashboard.html").Funcs(funcMap)
	htmlTmpl := template.Must(tmpl.Parse(langDashboardHTML))

	return func(w http.ResponseWriter, r *http.Request) {
		err := r.ParseForm()
		if err != nil {
			log.Printf("failed to parse form: %v", err)
			http.Error(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)

			return
		}

		model, err := handleLangDashboardRequest(w, r, store)
		if err != nil {
			log.Printf("failed to prepare view model: %v", err)
			http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)

			return
		}

		if err := htmlTmpl.ExecuteTemplate(w, "table", model); err != nil {
			log.Printf("failed to execute template: %v", err)
			http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)

			return
		}
	}
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

func handleLangDashboardRequest(
	w http.ResponseWriter,
	r *http.Request,
	store ViewModelStore,
) (*LangDashboardViewModel, error) {
	code := r.PathValue("code")
	if len(code) == 0 {
		http.Error(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)

		return nil, errors.New("missing lang code")
	}

	itemsTypeParam := findRequestValue(r, "NewItemsType", "CurrentItemsType", "items-type")
	filenameParam := findRequestValue(r, "", "CurrentFilename", "filename")
	filepathParam := findRequestValue(r, "NewFilepath", "CurrentFilepath", "filepath")
	sortParam := findRequestValue(r, "", "CurrentSort", "sort")
	sortOrderParam := findRequestValue(r, "", "CurrentSortOrder", "sort-order")

	url := buildURL(r.URL.Path, itemsTypeParam, filenameParam, filepathParam, sortParam, sortOrderParam)
	w.Header().Set("HX-Push", url)

	files, err := store.GetLangDashboardFiles(code)
	if err != nil {
		http.Error(w, "", http.StatusInternalServerError)
		return nil, err
	}

	itemsType := parseItemsTypeParam(itemsTypeParam)
	sort := parseSortParam(sortParam)
	sortOrder := parseSortOrderParam(sortOrderParam)

	var model LangDashboardViewModel

	model.URL = r.URL.RawPath
	model.CurrentLangCode = code
	model.CurrentItemsType = itemsTypeParam
	model.CurrentFilename = filenameParam
	model.CurrentFilepath = filepathParam
	model.CurrentSort = sortParam
	model.CurrentSortOrder = sortOrderParam

	if len(filenameParam) == 0 {
		model.ShowPanel = true
	}

	model.TableModel.FilenameColumnLink = buildColumnLinkURL(
		"filename", r.URL.Path, itemsTypeParam, filenameParam, filepathParam, sortParam, sortOrderParam,
	)
	model.TableModel.StatusColumnLink = buildColumnLinkURL(
		"status", r.URL.Path, itemsTypeParam, filenameParam, filepathParam, sortParam, sortOrderParam,
	)
	model.TableModel.UpdatesColumnLink = buildColumnLinkURL(
		"updates", r.URL.Path, itemsTypeParam, filenameParam, filepathParam, sortParam, sortOrderParam,
	)

	model.TableModel.Files = files

	FilterAndSort(&model, itemsType, filenameParam, filepathParam, sort, sortOrder)

	return &model, nil
}

func truncate(s string, length int) string {
	if len(s) > length {
		return s[:length]
	}
	return s
}

func buildURL(baseURL, itemsTypeParam, filenameParam, filepathParam, sortParam, sortOrderParam string) string {
	var queryParams []string
	if len(itemsTypeParam) > 0 {
		queryParams = append(queryParams, fmt.Sprintf("%s=%s", "items-type", itemsTypeParam))
	}
	if len(filenameParam) > 0 {
		queryParams = append(queryParams, fmt.Sprintf("%s=%s", "filename", filenameParam))
	}
	if len(filepathParam) > 0 {
		queryParams = append(queryParams, fmt.Sprintf("%s=%s", "filepath", filepathParam))
	}
	if len(sortParam) > 0 {
		queryParams = append(queryParams, fmt.Sprintf("%s=%s", "sort", sortParam))
	}
	if len(sortOrderParam) > 0 {
		queryParams = append(queryParams, fmt.Sprintf("%s=%s", "sort-order", sortOrderParam))
	}

	url := baseURL
	query := strings.Join(queryParams, "&")
	if len(query) > 0 {
		url += "?" + query
	}

	return url
}

func buildColumnLinkURL(sortFieldName, baseURL, itemsTypeParam, filenameParam, filepathParam, sortParam, sortOrderParam string) string {
	if sortFieldName == sortParam {
		if sortOrderParam == "asc" {
			sortOrderParam = "desc"
		} else {
			sortOrderParam = "asc"
		}

		return buildURL(baseURL, itemsTypeParam, filenameParam, filepathParam, sortFieldName, sortOrderParam)
	} else {
		return buildURL(baseURL, itemsTypeParam, filenameParam, filepathParam, sortFieldName, "")
	}
}

func parseItemsTypeParam(itemsTypeParam string) int {
	var itemsType int

	switch itemsTypeParam {
	case "all":
		itemsType = ItemsAll
	case "with-update":
		itemsType = ItemsWithUpdate
	case "with-update-or-pr":
		itemsType = ItemsWithUpdateOrPR
	case "with-pr":
		itemsType = ItemsWithPR
	default:
		itemsType = ItemsWithUpdateOrPR
	}

	return itemsType
}

func parseSortParam(sortParam string) int {
	var sort int

	switch sortParam {
	case "filename":
		sort = SortByFileName
	case "status":
		sort = SortByStatus
	case "updates":
		sort = SortByEnUpdate
	default:
		sort = SortByFileName
	}

	return sort
}

func parseSortOrderParam(sortOrderParam string) int {
	var sortOrder int

	switch sortOrderParam {
	case "asc":
		sortOrder = SortOrderAsc
	case "desc":
		sortOrder = SortOrderDesc
	default:
		sortOrder = SortOrderDesc
	}

	return sortOrder
}
