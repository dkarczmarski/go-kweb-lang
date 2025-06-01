package web

import (
	"fmt"
	"log"
	"time"

	"go-kweb-lang/git"
	"go-kweb-lang/gitseek"
)

type LangModel struct {
	Files []FileModel
}

type FileModel struct {
	LangRelPath     LinkModel
	LangLastCommit  git.CommitInfo
	LangMergeCommit *git.CommitInfo
	LangForkCommit  *git.CommitInfo
	ENStatus        string
	ENUpdates       []ENUpdate
	PRs             []LinkModel
}

type FileLinkModel struct {
	Text string
	Link LinkModel
}

type CommitLinkModel struct {
	Link       LinkModel
	CommitInfo git.CommitInfo
}

type ENUpdate struct {
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

		if len(fileInfo.ENFileStatus) == 0 &&
			len(fileInfo.ENUpdates) == 0 &&
			len(fileInfo.PRs) == 0 {
			continue
		}

		fileModel.LangRelPath = toLangFileLinkModel(fileInfo.LangRelPath)
		fileModel.LangLastCommit = convertCommitToUtc(fileInfo.LangLastCommit)
		fileModel.LangMergeCommit = convertCommitToUtcPtr(fileInfo.LangMergeCommit)
		fileModel.LangForkCommit = convertCommitToUtcPtr(fileInfo.LangForkCommit)
		fileModel.ENStatus = fileInfo.ENFileStatus

		for _, enUpdate := range fileInfo.ENUpdates {
			enUpdateModel := ENUpdate{
				Commit: toCommitLinkModel(enUpdate.Commit),
			}

			if enUpdate.MergePoint != nil {
				mergeCommit := toCommitLinkModel(*enUpdate.MergePoint)
				enUpdateModel.MergeCommit = &mergeCommit
			}

			fileModel.ENUpdates = append(fileModel.ENUpdates, enUpdateModel)
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

func convertCommitToUtc(commit git.CommitInfo) git.CommitInfo {
	modifyCommitDateTimeToUtc(&commit)

	return commit
}

func convertCommitToUtcPtr(commit *git.CommitInfo) *git.CommitInfo {
	if commit == nil {
		return nil
	}

	modifyCommitDateTimeToUtc(commit)

	return commit
}

func modifyCommitDateTimeToUtc(commit *git.CommitInfo) {
	commit.DateTime = convertDateStrToUtc(commit.DateTime)
}

func convertDateStrToUtc(dateStr string) string {
	if len(dateStr) == 0 {
		return dateStr
	}

	t, err := time.Parse(time.RFC3339, dateStr)
	if err != nil {
		log.Printf("time string %s parse error: %v", dateStr, err)
		return dateStr
	}

	return t.UTC().String()
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
