package web

type ViewModelStore interface {
	GetLangCodes() (*LangCodesViewModel, error)
	SetLangCodes(model *LangCodesViewModel) error
	GetLangDashboard(langCode string) (*LangDashboardViewModel, error)
	SetLangDashboard(langCode string, model *LangDashboardViewModel) error
}
