package web

import (
	"strconv"

	"github.com/dkarczmarski/go-kweb-lang/dashboard"
	"github.com/dkarczmarski/go-kweb-lang/git"
)

const shortDateLength = 10

type LangDashboardBuildInput struct {
	PagePath  string
	Dashboard dashboard.Dashboard
	Params    LangDashboardParams
}

func BuildLangCodesPageVM(index dashboard.LangIndex) LangCodesPageVM {
	items := make([]LinkVM, 0, len(index.Items))
	for _, item := range index.Items {
		items = append(items, LinkVM{
			Text: item.LangCode,
			URL:  "/lang/" + item.LangCode,
		})
	}

	return LangCodesPageVM{
		LangCodes: items,
	}
}

func BuildLangDashboardPageVM(input LangDashboardBuildInput) LangDashboardPageVM {
	urlBuilder := NewDashboardURLBuilder(input.PagePath, input.Params)
	visibleItems := FilterAndSortItems(input.Dashboard.Items, input.Params)
	rows := buildRows(urlBuilder, visibleItems)

	return LangDashboardPageVM{
		PageURL:   urlBuilder.Current(),
		LangCode:  input.Dashboard.LangCode,
		ShowPanel: shouldShowPanel(input.Params),
		Filters:   buildFiltersVM(urlBuilder, input.Params),
		Table:     buildTableVM(urlBuilder, input.Params, rows),
	}
}

func shouldShowPanel(params LangDashboardParams) bool {
	return params.Filename == ""
}

func buildFiltersVM(urlBuilder DashboardURLBuilder, params LangDashboardParams) DashboardFiltersVM {
	return DashboardFiltersVM{
		CurrentFilepath: params.Filepath,
		AllItems: FilterLinkVM{
			Label:  "all",
			URL:    urlBuilder.WithItemsType(ItemsTypeAll),
			Active: params.ItemsType == ItemsTypeAll,
		},
		ItemsWithUpdate: FilterLinkVM{
			Label:  "with update",
			URL:    urlBuilder.WithItemsType(ItemsTypeWithUpdate),
			Active: params.ItemsType == ItemsTypeWithUpdate,
		},
		ItemsWithUpdateOrPR: FilterLinkVM{
			Label:  "with update or pr",
			URL:    urlBuilder.WithItemsType(ItemsTypeWithUpdateOrPR),
			Active: params.ItemsType == ItemsTypeWithUpdateOrPR,
		},
		ItemsWithPR: FilterLinkVM{
			Label:  "with pr",
			URL:    urlBuilder.WithItemsType(ItemsTypeWithPR),
			Active: params.ItemsType == ItemsTypeWithPR,
		},
	}
}

func buildTableVM(
	urlBuilder DashboardURLBuilder,
	params LangDashboardParams,
	rows []DashboardRowVM,
) DashboardTableVM {
	return DashboardTableVM{
		FilenameHeader: buildSortHeaderVM(urlBuilder, params, SortByFilename),
		StatusHeader:   buildSortHeaderVM(urlBuilder, params, SortByStatus),
		UpdatesHeader:  buildSortHeaderVM(urlBuilder, params, SortByUpdates),
		Rows:           rows,
		Empty:          len(rows) == 0,
	}
}

func buildSortHeaderVM(
	urlBuilder DashboardURLBuilder,
	params LangDashboardParams,
	sortBy string,
) SortHeaderVM {
	active := params.SortBy == sortBy
	arrow := ""

	if active {
		if params.SortOrder == SortOrderDesc {
			arrow = "↓"
		} else {
			arrow = "↑"
		}
	}

	return SortHeaderVM{
		URL:    urlBuilder.Sort(sortBy),
		Arrow:  arrow,
		Active: active,
	}
}

func buildRows(urlBuilder DashboardURLBuilder, items []dashboard.Item) []DashboardRowVM {
	rows := make([]DashboardRowVM, 0, len(items))
	for _, item := range items {
		rows = append(rows, DashboardRowVM{
			Filename: buildFilenameCellVM(urlBuilder, item),
			Status:   buildStatusCellVM(item),
			Updates:  buildUpdatesCellVM(item),
			PRs:      buildPRsCellVM(item),
		})
	}

	return rows
}

func buildFilenameCellVM(urlBuilder DashboardURLBuilder, item dashboard.Item) FilenameCellVM {
	links := GitHubLinks{}
	displayPath := item.LangPath

	return FilenameCellVM{
		DisplayPath:     displayPath,
		GithubURL:       links.File(displayPath),
		DetailsURL:      urlBuilder.WithFilename(displayPath),
		LastCommitText:  buildCommitLabel("Last Commit", item.LangLastCommit.DateTime),
		MergeCommitText: buildOptionalCommitLabel("Merge Commit", item.LangMergeCommit),
		ForkCommitText:  buildOptionalCommitLabel("Fork Commit", item.LangForkCommit),
	}
}

func buildStatusCellVM(item dashboard.Item) StatusCellVM {
	return StatusCellVM{
		Text: item.FileStatus,
	}
}

func buildUpdatesCellVM(item dashboard.Item) UpdatesCellVM {
	updates := make([]UpdateItemVM, 0, len(item.EnUpdates))
	for _, update := range item.EnUpdates {
		updates = append(
			updates,
			buildUpdateItemVM(
				update.Commit.Comment,
				update.Commit.CommitID,
				update.Commit.DateTime,
				update.MergePoint,
			),
		)
	}

	lastUpdateText := ""
	if len(item.EnUpdates) > 0 {
		lastUpdateText = trimDate(latestEnUpdateDate(item))
	}

	return UpdatesCellVM{
		HasUpdates:     len(updates) > 0,
		LastUpdateText: lastUpdateText,
		Items:          updates,
	}
}

func buildUpdateItemVM(
	commitText string,
	commitID string,
	commitDate string,
	mergeCommit *git.CommitInfo,
) UpdateItemVM {
	links := GitHubLinks{}
	//nolint:exhaustruct
	viewModel := UpdateItemVM{
		CommitText: commitText,
		CommitURL:  links.Commit(commitID),
		CommitDate: trimDate(commitDate),
	}

	if mergeCommit != nil {
		viewModel.HasMergeCommit = true
		viewModel.MergeCommitText = mergeCommit.Comment
		viewModel.MergeCommitURL = links.Commit(mergeCommit.CommitID)
		viewModel.MergeCommitDate = trimDate(mergeCommit.DateTime)
	}

	return viewModel
}

func buildPRsCellVM(item dashboard.Item) PRsCellVM {
	linksBuilder := GitHubLinks{}
	links := make([]LinkVM, 0, len(item.PRs))

	for _, pullRequestNumber := range item.PRs {
		links = append(links, LinkVM{
			Text: "#" + strconv.Itoa(pullRequestNumber),
			URL:  linksBuilder.PR(pullRequestNumber),
		})
	}

	return PRsCellVM{
		Links: links,
		Empty: len(links) == 0,
	}
}

func buildCommitLabel(prefix string, date string) string {
	if date == "" {
		return ""
	}

	return prefix + ": " + trimDate(date)
}

func buildOptionalCommitLabel(prefix string, commit *git.CommitInfo) string {
	if commit == nil {
		return ""
	}

	date := commit.DateTime
	if date == "" {
		return ""
	}

	return prefix + ": " + trimDate(date)
}

func trimDate(value string) string {
	if len(value) >= shortDateLength {
		return value[:shortDateLength]
	}

	return value
}
