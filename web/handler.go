package web

import (
	_ "embed"
	"errors"
	"html/template"
	"log"
	"net/http"

	"go-kweb-lang/dashboard"

	"go-kweb-lang/web/internal/weberror"

	"go-kweb-lang/web/internal/reqhelper"
	"go-kweb-lang/web/internal/view"
)

//go:embed lang_codes.html
var langCodesHTML string

func createListLangCodesHandler(dashboardStore *dashboard.Store) func(w http.ResponseWriter, r *http.Request) {
	tmpl := template.New("lang_codes.html")
	htmlTmpl := template.Must(tmpl.Parse(langCodesHTML))

	return func(w http.ResponseWriter, r *http.Request) {
		dashboardIndex, err := dashboardStore.ReadDashboardIndex()
		if err != nil {
			logAndHTTPError(w, "failed to get language codes", err, http.StatusInternalServerError)

			return
		}

		model := view.BuildLangCodesModel(dashboardIndex)

		if err := htmlTmpl.Execute(w, model); err != nil {
			logAndHTTPError(w, "failed to execute template", err, http.StatusInternalServerError)

			return
		}
	}
}

//go:embed lang_dashboard.html
var langDashboardHTML string

func createLangDashboardHandler(dashboardStore *dashboard.Store) func(w http.ResponseWriter, r *http.Request) {
	funcMap := template.FuncMap{
		"truncate": truncate,
	}

	tmpl := template.New("lang_dashboard.html").Funcs(funcMap)
	htmlTmpl := template.Must(tmpl.Parse(langDashboardHTML))

	return func(w http.ResponseWriter, r *http.Request) {
		model, err := handleLangDashboardRequest(w, r, dashboardStore)
		if err != nil {
			logAndHTTPError(w, "failed to prepare view model", err, http.StatusInternalServerError)

			return
		}

		if err := htmlTmpl.Execute(w, model); err != nil {
			logAndHTTPError(w, "failed to execute template", err, http.StatusInternalServerError)

			return
		}
	}
}

func createLangDashboardTableHandler(dashboardStore *dashboard.Store) func(w http.ResponseWriter, r *http.Request) {
	funcMap := template.FuncMap{
		"truncate": truncate,
	}

	tmpl := template.New("lang_dashboard.html").Funcs(funcMap)
	htmlTmpl := template.Must(tmpl.Parse(langDashboardHTML))

	return func(w http.ResponseWriter, r *http.Request) {
		err := r.ParseForm()
		if err != nil {
			logAndHTTPError(w, "failed to parse form", err, http.StatusBadRequest)

			return
		}

		model, err := handleLangDashboardRequest(w, r, dashboardStore)
		if err != nil {
			logAndHTTPError(w, "failed to prepare view model", err, http.StatusInternalServerError)

			return
		}

		if err := htmlTmpl.ExecuteTemplate(w, "table", model); err != nil {
			logAndHTTPError(w, "failed to execute template", err, http.StatusInternalServerError)

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

func handleLangDashboardRequest(
	w http.ResponseWriter,
	r *http.Request,
	dashboardStore *dashboard.Store,
) (*view.LangDashboardViewModel, error) {
	requestModel, err := reqhelper.ParseListLangDashboardRequest(r)
	if err != nil {
		return nil, err
	}

	langDashboard, err := dashboardStore.ReadDashboard(requestModel.LangCode)
	if err != nil {
		return nil, err
	}

	handleHtmx(w, r, requestModel)

	langDashboardFilesModel := view.BuildLangDashboardFilesModel(langDashboard.LangCode, langDashboard.Items)

	return view.BuildLangDashboardModel(r, requestModel, langDashboardFilesModel)
}

func handleHtmx(w http.ResponseWriter, r *http.Request, requestModel reqhelper.RequestModel) {
	url := view.BuildURL(r.URL.Path, requestModel)
	w.Header().Set("HX-Push", url)
}

func logAndHTTPError(w http.ResponseWriter, msg string, err error, code int) {
	log.Printf("%s: %v", msg, err)

	var webError *weberror.WebError
	if errors.As(err, &webError) {
		http.Error(w, http.StatusText(webError.HTTPCode), webError.HTTPCode)
	} else {
		http.Error(w, http.StatusText(code), code)
	}
}
