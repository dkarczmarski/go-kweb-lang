package web

import "go-kweb-lang/web/internal/view"

type ViewModelStore interface {
	GetLangCodes() (*view.LangCodesViewModel, error)
	SetLangCodes(model *view.LangCodesViewModel) error
	GetLangDashboardFiles(langCode string) ([]view.FileModel, error)
	SetLangDashboardFiles(langCode string, files []view.FileModel) error
}
