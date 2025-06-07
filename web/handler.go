package web

import (
	_ "embed"
	"html/template"
	"log"
	"net/http"
)

//go:embed index.html
var indexHTML string

func createListLangCodesHandler(templateData *TemplateData) func(w http.ResponseWriter, r *http.Request) {
	indexTmpl := template.Must(template.New("index.html").Parse(indexHTML))

	return func(w http.ResponseWriter, r *http.Request) {
		model := templateData.GetIndex()
		if err := indexTmpl.Execute(w, model); err != nil {
			log.Println(err)
			http.Error(w, "", http.StatusInternalServerError)
			return
		}
	}
}

//go:embed lang.html
var langHTML string

func createLangDashboardHandler(templateData *TemplateData) func(w http.ResponseWriter, r *http.Request) {
	funcMap := template.FuncMap{
		"truncate": truncate,
	}
	langTmpl := template.Must(template.New("lang.html").Funcs(funcMap).Parse(langHTML))

	return func(w http.ResponseWriter, r *http.Request) {
		code := r.PathValue("code")

		model := templateData.GetLang(code)
		if model == nil {
			http.NotFound(w, r)
			return
		}

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
