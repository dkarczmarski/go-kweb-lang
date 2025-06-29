package web

import "go-kweb-lang/web/internal/view"

type ViewModelStore interface {
	GetLangCodes() (*view.LangCodesViewModel, error)
	SetLangCodes(model *view.LangCodesViewModel) error
	GetLangDashboardFiles(langCode string) (view.LangDashboardFilesModel, error)
	SetLangDashboardFiles(langCode string, langDashboardFiles view.LangDashboardFilesModel) error
}
