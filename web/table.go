package web

import (
	"go-kweb-lang/git"
	"go-kweb-lang/seek"
)

type TableModel struct {
	Files []FileModel
}

type FileModel struct {
	LangRelPath    LinkModel
	LangLastCommit git.CommitInfo
	OriginStatus   string
	OriginUpdates  []CommitLinkModel
}

type FileLinkModel struct {
	Text string
	Link LinkModel
}

type CommitLinkModel struct {
	Link   LinkModel
	Commit git.CommitInfo
}

type LinkModel struct {
	Text string
	Url  string
}

func BuildTableModel(fileInfos []seek.FileInfo) *TableModel {
	var table TableModel

	for _, fileInfo := range fileInfos {
		var fileModel FileModel
		if len(fileInfo.OriginFileStatus) == 0 && len(fileInfo.OriginUpdates) == 0 {
			continue
		}

		fileModel.LangRelPath = toLangFileLinkModel(fileInfo.LangRelPath)
		fileModel.LangLastCommit = fileInfo.LangCommit
		fileModel.OriginStatus = fileInfo.OriginFileStatus

		// todo: more than one commit can have the same merge point. show only one
		for _, originUpdate := range fileInfo.OriginUpdates {
			fileModel.OriginUpdates = append(fileModel.OriginUpdates, toCommitLinkModel(originUpdate.MergePoint))
		}

		table.Files = append(table.Files, fileModel)
	}

	return &table
}

func toLinkModel(commitId string) LinkModel {
	return LinkModel{
		Text: commitId,
		Url:  "https://github.com/kubernetes/website/commit/" + commitId,
	}
}

func toLangFileLinkModel(langRelPath string) LinkModel {
	return LinkModel{
		Text: langRelPath,
		Url:  "https://github.com/kubernetes/website/blob/main/content/pl/" + langRelPath,
	}
}

func toCommitLinkModel(commit git.CommitInfo) CommitLinkModel {
	return CommitLinkModel{
		Link:   toLinkModel(commit.CommitId),
		Commit: commit,
	}
}
