package web

import (
	"fmt"

	"go-kweb-lang/git"
	"go-kweb-lang/gitseek"
)

type LangModel struct {
	Files []FileModel
}

type FileModel struct {
	LangRelPath    LinkModel
	LangLastCommit git.CommitInfo
	OriginStatus   string
	OriginUpdates  []OriginUpdate
	PRs            []LinkModel
}

type FileLinkModel struct {
	Text string
	Link LinkModel
}

type CommitLinkModel struct {
	Link       LinkModel
	CommitInfo git.CommitInfo
}

type OriginUpdate struct {
	Commit      CommitLinkModel
	MergeCommit *CommitLinkModel
}

type FileInfo struct {
	gitseek.FileInfo
	PRs []int
}

func BuildLangModel(fileInfos []FileInfo) *LangModel {
	var table LangModel

	for _, fileInfo := range fileInfos {
		var fileModel FileModel

		if len(fileInfo.OriginFileStatus) == 0 &&
			len(fileInfo.OriginUpdates) == 0 &&
			len(fileInfo.PRs) == 0 {
			continue
		}

		fileModel.LangRelPath = toLangFileLinkModel(fileInfo.LangRelPath)
		fileModel.LangLastCommit = fileInfo.LangCommit
		fileModel.OriginStatus = fileInfo.OriginFileStatus

		for _, originUpdate := range fileInfo.OriginUpdates {
			originUpdateModel := OriginUpdate{
				Commit: toCommitLinkModel(originUpdate.Commit),
			}

			if originUpdate.MergePoint != nil {
				mergeCommit := toCommitLinkModel(*originUpdate.MergePoint)
				originUpdateModel.MergeCommit = &mergeCommit
			}

			fileModel.OriginUpdates = append(fileModel.OriginUpdates, originUpdateModel)
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
		Link:       toLinkModel(commit.CommitID),
		CommitInfo: commit,
	}
}
