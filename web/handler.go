package web

import (
	_ "embed"
	"html/template"
	"log"
	"net/http"
)

//go:embed index.html
var indexHTML string

func createListLangCodesHandler(store ViewModelStore) func(w http.ResponseWriter, r *http.Request) {
	indexTmpl := template.Must(template.New("index.html").Parse(indexHTML))

	return func(w http.ResponseWriter, r *http.Request) {
		model, err := store.GetLangCodes()
		if err != nil {
			log.Println(err)
			http.Error(w, "", http.StatusInternalServerError)
			return
		}

		if err := indexTmpl.Execute(w, model); err != nil {
			log.Println(err)
			http.Error(w, "", http.StatusInternalServerError)
			return
		}
	}
}

//go:embed lang.html
var langHTML string

func createLangDashboardHandler(store ViewModelStore) func(w http.ResponseWriter, r *http.Request) {
	funcMap := template.FuncMap{
		"truncate": truncate,
	}
	langTmpl := template.Must(template.New("lang.html").Funcs(funcMap).Parse(langHTML))

	return func(w http.ResponseWriter, r *http.Request) {
		code := r.PathValue("code")

		itemsParam := r.URL.Query().Get("items")
		filenameParam := r.URL.Query().Get("filename")
		filepathParam := r.URL.Query().Get("filepath")
		sortParam := r.URL.Query().Get("sort")
		sortOrderParam := r.URL.Query().Get("order")

		model, err := store.GetLangDashboard(code)
		if err != nil {
			log.Println(err)
			http.Error(w, "", http.StatusInternalServerError)
			return
		}

		if model == nil {
			http.NotFound(w, r)
			return
		}

		var itemsFilter int
		switch itemsParam {
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

		model = FilterAndSort(model, itemsFilter, filenameParam, filepathParam, sort, sortOrder)
		if err := langTmpl.Execute(w, model); err != nil {
			log.Println(err)
			http.Error(w, "", http.StatusInternalServerError)
			return
		}
	}
}

func truncate(s string, length int) string {
	if len(s) > length {
		return s[:length]
	}
	return s
}
