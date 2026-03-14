package web

import (
	_ "embed"
	"fmt"
	"html/template"
	"log"
	"net/http"

	"github.com/dkarczmarski/go-kweb-lang/dashboard"
)

//go:embed lang_codes.html
var langCodesHTML string

//go:embed lang_dashboard.html
var langDashboardHTML string

type Handler struct {
	dashboardStore *dashboard.Store
	langCodesTmpl  *template.Template
	dashboardTmpl  *template.Template
}

func NewHandler(dashboardStore *dashboard.Store) *Handler {
	langCodesTemplate := template.Must(template.New("lang_codes.html").Parse(langCodesHTML))
	dashboardTemplate := template.Must(template.New("lang_dashboard.html").Parse(langDashboardHTML))

	return &Handler{
		dashboardStore: dashboardStore,
		langCodesTmpl:  langCodesTemplate,
		dashboardTmpl:  dashboardTemplate,
	}
}

func (handler *Handler) Register(mux *http.ServeMux) {
	mux.HandleFunc("GET /", handler.ListLangCodes)
	mux.HandleFunc("GET /lang/{code}", handler.ShowLangDashboard)
	mux.HandleFunc("POST /lang/{code}", handler.ShowLangDashboardTable)
}

func (handler *Handler) ListLangCodes(responseWriter http.ResponseWriter, _ *http.Request) {
	index, err := handler.dashboardStore.ReadDashboardIndex()
	if err != nil {
		log.Printf("list lang codes: %v", err)
		http.Error(
			responseWriter,
			http.StatusText(http.StatusInternalServerError),
			http.StatusInternalServerError,
		)

		return
	}

	pageViewModel := BuildLangCodesPageVM(index)
	if err := handler.langCodesTmpl.Execute(responseWriter, pageViewModel); err != nil {
		log.Printf("render lang codes: %v", err)
		http.Error(
			responseWriter,
			http.StatusText(http.StatusInternalServerError),
			http.StatusInternalServerError,
		)

		return
	}
}

func (handler *Handler) ShowLangDashboard(
	responseWriter http.ResponseWriter,
	request *http.Request,
) {
	pageViewModel, err := handler.prepareLangDashboardVM(request)
	if err != nil {
		log.Printf("prepare dashboard model: %v", err)
		http.Error(
			responseWriter,
			http.StatusText(http.StatusInternalServerError),
			http.StatusInternalServerError,
		)

		return
	}

	responseWriter.Header().Set("HX-Push", pageViewModel.PageURL)

	if err := handler.dashboardTmpl.Execute(responseWriter, pageViewModel); err != nil {
		log.Printf("render dashboard: %v", err)
		http.Error(
			responseWriter,
			http.StatusText(http.StatusInternalServerError),
			http.StatusInternalServerError,
		)

		return
	}
}

func (handler *Handler) ShowLangDashboardTable(
	responseWriter http.ResponseWriter,
	request *http.Request,
) {
	pageViewModel, err := handler.prepareLangDashboardVM(request)
	if err != nil {
		log.Printf("prepare dashboard table model: %v", err)
		http.Error(
			responseWriter,
			http.StatusText(http.StatusInternalServerError),
			http.StatusInternalServerError,
		)

		return
	}

	responseWriter.Header().Set("HX-Push", pageViewModel.PageURL)

	if err := handler.dashboardTmpl.ExecuteTemplate(responseWriter, "table", pageViewModel); err != nil {
		log.Printf("render dashboard table: %v", err)
		http.Error(
			responseWriter,
			http.StatusText(http.StatusInternalServerError),
			http.StatusInternalServerError,
		)

		return
	}
}

func (handler *Handler) prepareLangDashboardVM(request *http.Request) (LangDashboardPageVM, error) {
	if err := request.ParseForm(); err != nil {
		return LangDashboardPageVM{}, fmt.Errorf("parse lang dashboard form: %w", err)
	}

	langCode := request.PathValue("code")
	params := ParseLangDashboardParams(langCode, request.Form)

	dashboardData, err := handler.dashboardStore.ReadDashboard(langCode)
	if err != nil {
		return LangDashboardPageVM{}, fmt.Errorf("read dashboard for lang code %s: %w", langCode, err)
	}

	return BuildLangDashboardPageVM(LangDashboardBuildInput{
		PagePath:  request.URL.Path,
		Dashboard: dashboardData,
		Params:    params,
	}), nil
}
