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
	if r.FormValue(newValueKey) == "" && r.FormValue(currentValueKey) != "" {
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

	itemsTypeParam := findRequestValue(r, "SelectedItemsType", "ParamItemsType", "items-type")
	filenameParam := findRequestValue(r, "", "ParamFilename", "filename")
	filepathParam := findRequestValue(r, "SelectedFilepath", "ParamFilepath", "filepath")
	sortParam := findRequestValue(r, "SelectedSort", "ParamSort", "sort")
	sortOrderParam := findRequestValue(r, "SelectedSortOrder", "ParamSortOrder", "sort-order")

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

	url := r.URL.Path
	query := strings.Join(queryParams, "&")
	if len(query) > 0 {
		url += "?" + query
	}
	w.Header().Set("HX-Push", url)

	model, err := store.GetLangDashboard(code)
	if err != nil {
		http.Error(w, "", http.StatusInternalServerError)
		return nil, err
	}

	if model == nil {
		http.NotFound(w, r)
		return nil, err
	}

	var itemsFilter int
	switch itemsTypeParam {
	case "all":
		itemsFilter = ItemsAll
	case "with-update":
		itemsFilter = ItemsWithUpdate
	case "with-update-or-pr":
		itemsFilter = ItemsWithUpdateOrPR
	case "with-pr":
		itemsFilter = ItemsWithPR
	default:
		itemsFilter = ItemsWithUpdateOrPR
	}

	var sort int
	switch sortParam {
	case "filename":
		sort = SortByFileName
	case "lastcommit":
		sort = SortByLastLangFileCommit
	default:
		sort = SortByFileName
	}

	var sortOrder int
	switch sortOrderParam {
	case "asc":
		sortOrder = SortOrderAsc
	case "desc":
		sortOrder = SortOrderDesc
	default:
		sortOrder = SortOrderDesc
	}

	model.URL = r.URL.RawPath

	model.ParamLangCode = code
	model.ParamItemsType = itemsTypeParam
	model.ParamFilename = filenameParam
	model.ParamFilepath = filepathParam

	if len(filenameParam) == 0 {
		model.ShowPanel = true
	}

	FilterAndSort(model, itemsFilter, filenameParam, filepathParam, sort, sortOrder)

	return model, nil
}

func truncate(s string, length int) string {
	if len(s) > length {
		return s[:length]
	}
	return s
}
