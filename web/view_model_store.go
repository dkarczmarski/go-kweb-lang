package web

type ViewModelStore interface {
	GetLangCodes() (*LangCodesViewModel, error)
	SetLangCodes(model *LangCodesViewModel) error
	GetLangDashboardFiles(langCode string) ([]FileModel, error)
	SetLangDashboardFiles(langCode string, files []FileModel) error
}
