package web

import (
	"go-kweb-lang/git"
	"go-kweb-lang/seek"
)

type LangModel struct {
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

func BuildLangModel(fileInfos []seek.FileInfo) *LangModel {
	var table LangModel

	for _, fileInfo := range fileInfos {
		var fileModel FileModel
		if len(fileInfo.OriginFileStatus) == 0 && len(fileInfo.OriginUpdates) == 0 {
			continue
		}

		fileModel.LangRelPath = toLangFileLinkModel(fileInfo.LangRelPath)
		fileModel.LangLastCommit = fileInfo.LangCommit
		fileModel.OriginStatus = fileInfo.OriginFileStatus

		mergePoints := make(map[string]interface{})
		for _, originUpdate := range fileInfo.OriginUpdates {
			if _, ok := mergePoints[originUpdate.MergePoint.CommitId]; ok {
				continue
			}

			fileModel.OriginUpdates = append(fileModel.OriginUpdates, toCommitLinkModel(originUpdate.MergePoint))
			mergePoints[originUpdate.MergePoint.CommitId] = struct{}{}
		}

		table.Files = append(table.Files, fileModel)
	}

	return &table
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
