package web

import (
	"fmt"
	"log"
	"time"

	"go-kweb-lang/git"
)

type LangCodesProvider interface {
	LangCodes() ([]string, error)
}

func buildLangCodesViewModel(langCodesProvider LangCodesProvider) (*LangCodesViewModel, error) {
	tableModel, err := buildLangCodesTableModel(langCodesProvider)
	if err != nil {
		return nil, fmt.Errorf("error while building index web model: %w", err)
	}

	langCodesViewModel := &LangCodesViewModel{
		LangCodes: tableModel,
	}

	return langCodesViewModel, nil
}

func buildLangCodesTableModel(langCodesProvider LangCodesProvider) ([]LinkModel, error) {
	langCodes, err := langCodesProvider.LangCodes()
	if err != nil {
		return nil, fmt.Errorf("error while getting available languages: %w", err)
	}

	model := make([]LinkModel, 0, len(langCodes))
	for _, langCode := range langCodes {
		model = append(model, LinkModel{
			Text: langCode,
			URL:  "lang/" + langCode,
		})
	}

	return model, nil
}

func buildLangTableFilesModel(langCode string, fileInfos []FileInfo) []FileModel {
	var files []FileModel

	for _, fileInfo := range fileInfos {
		var fileModel FileModel

		if len(fileInfo.ENFileStatus) == 0 &&
			len(fileInfo.ENUpdates) == 0 &&
			len(fileInfo.PRs) == 0 {
			continue
		}

		fileModel.LangRelPath = toLangFileLinkModel(langCode, fileInfo.LangRelPath)
		fileModel.LangFilenameLink = LinkModel{
			Text: "#",
			URL:  fmt.Sprintf("/lang/%s?filename=%s", langCode, fileInfo.LangRelPath),
		}
		fileModel.LangLastCommit = convertCommitToUtc(fileInfo.LangLastCommit)
		fileModel.LangMergeCommit = convertCommitToUtcPtr(fileInfo.LangMergeCommit)
		fileModel.LangForkCommit = convertCommitToUtcPtr(fileInfo.LangForkCommit)
		fileModel.ENStatus = fileInfo.ENFileStatus

		var enUpdates []ENUpdate
		for _, enUpdate := range fileInfo.ENUpdates {
			enUpdateModel := ENUpdate{
				Commit: toCommitLinkModel(enUpdate.Commit),
			}

			if enUpdate.MergePoint != nil {
				mergeCommit := toCommitLinkModel(*enUpdate.MergePoint)
				enUpdateModel.MergeCommit = &mergeCommit
			}

			enUpdates = append(enUpdates, enUpdateModel)
		}

		fileModel.ENUpdates = buildENUpdates(
			enUpdates, fileModel.LangForkCommit, fileModel.LangLastCommit, fileModel.LangMergeCommit,
		)

		var prLinks []LinkModel
		for _, pr := range fileInfo.PRs {
			prLinks = append(prLinks, LinkModel{
				Text: fmt.Sprintf("%v", pr),
				URL:  fmt.Sprintf("https://github.com/kubernetes/website/pull/%v", pr),
			})
		}
		fileModel.PRs = prLinks

		files = append(files, fileModel)
	}

	return files
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

func buildENUpdates(
	enUpdates []ENUpdate,
	langForkCommit *git.CommitInfo,
	langLastCommit git.CommitInfo,
	langMergeCommit *git.CommitInfo,
) ENUpdateGroups {
	var groups ENUpdateGroups

	if langForkCommit == nil || langMergeCommit == nil {
		groups.AfterLastCommit = enUpdates
	} else {
		for _, enUpdate := range enUpdates {
			enUpdateDataTime := convertDateStrToUtc(enUpdate.Commit.CommitInfo.DateTime)

			if enUpdateDataTime < langForkCommit.DateTime {
				groups.BeforeForkCommit = append(groups.BeforeForkCommit, enUpdate)
			} else if enUpdateDataTime < langLastCommit.DateTime {
				groups.AfterForkCommit = append(groups.AfterForkCommit, enUpdate)
			} else if enUpdateDataTime < langMergeCommit.DateTime {
				groups.AfterLastCommit = append(groups.AfterLastCommit, enUpdate)
			} else {
				groups.AfterMergeCommit = append(groups.AfterMergeCommit, enUpdate)
			}
		}
	}

	groups.LastCommit = findLastCommitInGroup(enUpdates)

	return groups
}

func findLastCommitInGroup(enUpdates []ENUpdate) git.CommitInfo {
	var lastCommit git.CommitInfo

	for _, enUpdate := range enUpdates {
		if lastCommit.DateTime < enUpdate.Commit.CommitInfo.DateTime {
			lastCommit = enUpdate.Commit.CommitInfo
		}
	}

	return lastCommit
}

func toLangFileLinkModel(langCode, langRelPath string) LinkModel {
	return LinkModel{
		Text: langRelPath,
		URL:  fmt.Sprintf("https://github.com/kubernetes/website/blob/main/content/%s/%s", langCode, langRelPath),
	}
}

func toCommitLinkModel(commit git.CommitInfo) CommitLinkModel {
	return CommitLinkModel{
		Link: LinkModel{
			Text: commit.CommitID,
			URL:  "https://github.com/kubernetes/website/commit/" + commit.CommitID,
		},
		CommitInfo: commit,
	}
}
