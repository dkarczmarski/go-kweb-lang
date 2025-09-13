package view

import (
	"fmt"
	"log"
	"net/http"
	"strings"
	"time"

	"go-kweb-lang/dashboard"

	"go-kweb-lang/web/internal/reqhelper"

	"go-kweb-lang/git"
)

type LangCodesProvider interface {
	LangCodes() ([]string, error)
}

func BuildLangCodesModel(dashboardIndex *dashboard.LangIndex) *LangCodesViewModel {
	items := dashboardIndex.Items

	model := make([]LinkModel, 0, len(items))
	for _, item := range items {
		model = append(model, LinkModel{
			Text: item.LangCode,
			URL:  "lang/" + item.LangCode,
		})
	}

	langCodesViewModel := &LangCodesViewModel{
		LangCodes: model,
	}

	return langCodesViewModel
}

func BuildLangDashboardFilesModel(langCode string, fileInfos []dashboard.Item) LangDashboardFilesModel {
	var files []FileModel
	for _, fileInfo := range fileInfos {
		var fileModel FileModel

		if len(fileInfo.FileStatus) == 0 && len(fileInfo.ENUpdates) == 0 && len(fileInfo.PRs) == 0 {
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
		fileModel.Status = fileInfo.FileStatus

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

	return LangDashboardFilesModel{
		Files: files,
	}
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

			// assumes all dates have already been converted to UTC.
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

func BuildLangDashboardModel(
	r *http.Request,
	requestModel reqhelper.RequestModel,
	langDashboardFilesModel LangDashboardFilesModel,
) (*LangDashboardViewModel, error) {
	var model LangDashboardViewModel

	model.URL = r.URL.RawPath
	model.CurrentLangCode = requestModel.LangCode
	model.CurrentItemsType = requestModel.ItemsTypeParam
	model.CurrentFilename = requestModel.FilenameParam
	model.CurrentFilepath = requestModel.FilepathParam
	model.CurrentSort = requestModel.SortParam
	model.CurrentSortOrder = requestModel.SortOrderParam

	if len(requestModel.FilenameParam) == 0 {
		model.ShowPanel = true
	}

	model.TableModel.FilenameColumnLink = buildColumnLinkURL("filename", r.URL.Path, requestModel)
	model.TableModel.StatusColumnLink = buildColumnLinkURL("status", r.URL.Path, requestModel)
	model.TableModel.UpdatesColumnLink = buildColumnLinkURL("updates", r.URL.Path, requestModel)
	model.TableModel.Files = filterAndSortFromRequest(langDashboardFilesModel.Files, requestModel)

	return &model, nil
}

func BuildURL(baseURL string, requestModel reqhelper.RequestModel) string {
	var queryParams []string
	if len(requestModel.ItemsTypeParam) > 0 {
		queryParams = append(queryParams, fmt.Sprintf("%s=%s", "items-type", requestModel.ItemsTypeParam))
	}
	if len(requestModel.FilenameParam) > 0 {
		queryParams = append(queryParams, fmt.Sprintf("%s=%s", "filename", requestModel.FilenameParam))
	}
	if len(requestModel.FilepathParam) > 0 {
		queryParams = append(queryParams, fmt.Sprintf("%s=%s", "filepath", requestModel.FilepathParam))
	}
	if len(requestModel.SortParam) > 0 {
		queryParams = append(queryParams, fmt.Sprintf("%s=%s", "sort", requestModel.SortParam))
	}
	if len(requestModel.SortOrderParam) > 0 {
		queryParams = append(queryParams, fmt.Sprintf("%s=%s", "sort-order", requestModel.SortOrderParam))
	}

	url := baseURL
	query := strings.Join(queryParams, "&")
	if len(query) > 0 {
		url += "?" + query
	}

	return url
}

func buildColumnLinkURL(sortFieldName, baseURL string, requestModel reqhelper.RequestModel) string {
	if sortFieldName == requestModel.SortParam {
		if requestModel.SortOrderParam == "asc" {
			requestModel.SortOrderParam = "desc"
		} else {
			requestModel.SortOrderParam = "asc"
		}

		return BuildURL(baseURL, requestModel)
	} else {
		requestModel.SortOrderParam = ""

		return BuildURL(baseURL, requestModel)
	}
}

func filterAndSortFromRequest(files []FileModel, requestModel reqhelper.RequestModel) []FileModel {
	itemsType := parseItemsTypeParam(requestModel.ItemsTypeParam)
	sort := parseSortParam(requestModel.SortParam)
	sortOrder := parseSortOrderParam(requestModel.SortOrderParam)

	return FilterAndSort(files, itemsType, requestModel.FilenameParam, requestModel.FilepathParam, sort, sortOrder)
}

func parseItemsTypeParam(itemsTypeParam string) int {
	var itemsType int

	switch itemsTypeParam {
	case "all":
		itemsType = ItemsAll
	case "with-update":
		itemsType = ItemsWithUpdate
	case "with-update-or-pr":
		itemsType = ItemsWithUpdateOrPR
	case "with-pr":
		itemsType = ItemsWithPR
	default:
		itemsType = ItemsWithUpdateOrPR
	}

	return itemsType
}

func parseSortParam(sortParam string) int {
	var sort int

	switch sortParam {
	case "filename":
		sort = SortByFileName
	case "status":
		sort = SortByStatus
	case "updates":
		sort = SortByEnUpdate
	default:
		sort = SortByFileName
	}

	return sort
}

func parseSortOrderParam(sortOrderParam string) int {
	var sortOrder int

	switch sortOrderParam {
	case "asc":
		sortOrder = SortOrderAsc
	case "desc":
		sortOrder = SortOrderDesc
	default:
		sortOrder = SortOrderDesc
	}

	return sortOrder
}
