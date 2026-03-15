package web

type LinkVM struct {
	Text string
	URL  string
}

type LangCodesPageVM struct {
	LangCodes []LinkVM
}

type LangDashboardPageVM struct {
	PageURL  string
	LangCode string

	ShowPanel bool
	Filters   DashboardFiltersVM
	Table     DashboardTableVM
}

type DashboardFiltersVM struct {
	CurrentFilepath string

	ItemsWithEnUpdates        FilterLinkVM
	ItemsWithPR               FilterLinkVM
	ItemsEnFileDoesNotExist   FilterLinkVM
	ItemsEnFileNoLongerExists FilterLinkVM
	ItemsLangFileMissing      FilterLinkVM
	ItemsWaitingForReview     FilterLinkVM
	ItemsLangFileUpToDate     FilterLinkVM
}

type FilterLinkVM struct {
	Label  string
	Value  string
	Active bool
}

type DashboardTableVM struct {
	FilenameHeader SortHeaderVM
	StatusHeader   SortHeaderVM
	UpdatesHeader  SortHeaderVM
	Rows           []DashboardRowVM
	Empty          bool
}

type SortHeaderVM struct {
	URL    string
	Arrow  string
	Active bool
}

type DashboardRowVM struct {
	Filename FilenameCellVM
	Status   StatusCellVM
	Updates  UpdatesCellVM
	PRs      PRsCellVM
}

type FilenameCellVM struct {
	DisplayPath string
	GithubURL   string
	DetailsURL  string

	LastCommitText  string
	MergeCommitText string
	ForkCommitText  string
}

type StatusCellVM struct {
	Text string
}

type UpdatesCellVM struct {
	HasUpdates     bool
	LastUpdateText string
	Items          []UpdateItemVM
}

type UpdateItemVM struct {
	CommitText string
	CommitURL  string
	CommitDate string

	MergeCommitText string
	MergeCommitURL  string
	MergeCommitDate string
	HasMergeCommit  bool
}

type PRsCellVM struct {
	Links []LinkVM
	Empty bool
}
