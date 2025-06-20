package view

import (
	"go-kweb-lang/git"
	"go-kweb-lang/gitseek"
)

type LangCodesViewModel struct {
	LangCodes []LinkModel
}

type LangDashboardViewModel struct {
	URL string

	CurrentLangCode  string
	CurrentItemsType string
	CurrentFilename  string
	CurrentFilepath  string
	CurrentSort      string
	CurrentSortOrder string

	ShowPanel bool

	TableModel TableModel
}

type TableModel struct {
	FilenameColumnLink string
	StatusColumnLink   string
	UpdatesColumnLink  string
	Files              []FileModel
}

type FileModel struct {
	LangRelPath      LinkModel
	LangFilenameLink LinkModel
	LangLastCommit   git.CommitInfo
	LangMergeCommit  *git.CommitInfo
	LangForkCommit   *git.CommitInfo
	ENStatus         string
	ENUpdates        ENUpdateGroups
	PRs              []LinkModel
}

type FileLinkModel struct {
	Text string
	Link LinkModel
}

type CommitLinkModel struct {
	Link       LinkModel
	CommitInfo git.CommitInfo
}

type ENUpdateGroups struct {
	LastCommit       git.CommitInfo
	AfterMergeCommit []ENUpdate
	AfterLastCommit  []ENUpdate
	AfterForkCommit  []ENUpdate
	BeforeForkCommit []ENUpdate
}

type ENUpdate struct {
	Commit      CommitLinkModel
	MergeCommit *CommitLinkModel
}

type FileInfo struct {
	gitseek.FileInfo
	PRs []int
}

type LinkModel struct {
	Text string
	URL  string
}
