package web

import (
	"fmt"
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
	PRs            []LinkModel
}

type FileLinkModel struct {
	Text string
	Link LinkModel
}

type CommitLinkModel struct {
	Link   LinkModel
	Commit git.CommitInfo
}

type FileInfo struct {
	seek.FileInfo
	PRs []int
}

func BuildLangModel(fileInfos []FileInfo) *LangModel {
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
			if _, ok := mergePoints[originUpdate.MergePoint.CommitID]; ok {
				continue
			}

			fileModel.OriginUpdates = append(fileModel.OriginUpdates, toCommitLinkModel(originUpdate.MergePoint))
			mergePoints[originUpdate.MergePoint.CommitID] = struct{}{}
		}

		var prLinks []LinkModel
		for _, pr := range fileInfo.PRs {
			prLinks = append(prLinks, LinkModel{
				Text: fmt.Sprintf("%v", pr),
				URL:  fmt.Sprintf("https://github.com/kubernetes/website/pull/%v", pr),
			})
		}
		fileModel.PRs = prLinks

		table.Files = append(table.Files, fileModel)
	}

	return &table
}

func toLangFileLinkModel(langRelPath string) LinkModel {
	return LinkModel{
		Text: langRelPath,
		URL:  "https://github.com/kubernetes/website/blob/main/content/pl/" + langRelPath,
	}
}

func toCommitLinkModel(commit git.CommitInfo) CommitLinkModel {
	return CommitLinkModel{
		Link:   toLinkModel(commit.CommitID),
		Commit: commit,
	}
}
