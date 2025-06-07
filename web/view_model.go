package web

import (
	"go-kweb-lang/git"
	"go-kweb-lang/gitseek"
)

type LangCodesViewModel struct {
	LangCodes []LinkModel
}

type LangDashboardViewModel struct {
	TableModel LangModel
}

type LangModel struct {
	Files []FileModel
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
