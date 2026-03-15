//nolint:testpackage,goconst
package web

import (
	"testing"

	"github.com/dkarczmarski/go-kweb-lang/dashboard"
	"github.com/dkarczmarski/go-kweb-lang/git"
	"github.com/dkarczmarski/go-kweb-lang/gitseek"
)

func TestBuildLangCodesPageVM(t *testing.T) {
	t.Parallel()

	index := dashboard.LangIndex{
		Items: []dashboard.LangIndexItem{
			{LangCode: "pl"},
			{LangCode: "de"},
		},
	}

	viewModel := BuildLangCodesPageVM(index)

	if len(viewModel.LangCodes) != 2 {
		t.Fatalf("expected 2 lang codes, got %d", len(viewModel.LangCodes))
	}

	if viewModel.LangCodes[0].Text != "pl" || viewModel.LangCodes[0].URL != "/lang/pl" {
		t.Fatalf("unexpected first lang code vm: %#v", viewModel.LangCodes[0])
	}

	if viewModel.LangCodes[1].Text != "de" || viewModel.LangCodes[1].URL != "/lang/de" {
		t.Fatalf("unexpected second lang code vm: %#v", viewModel.LangCodes[1])
	}
}

func TestBuildLangDashboardPageVM(t *testing.T) {
	t.Parallel()

	dash := dashboard.Dashboard{
		LangCode: "pl",
		Items: []dashboard.Item{
			{
				FileInfo: gitseek.FileInfo{
					LangPath:   "content/pl/test.md",
					FileStatus: "en-file-updated",
					LangLastCommit: git.CommitInfo{
						DateTime: "2023-01-01T10:00:00Z",
					},
					EnUpdates: []gitseek.EnUpdate{
						{
							Commit: git.CommitInfo{
								Comment:  "update en",
								CommitID: "abc123",
								DateTime: "2023-01-02T10:00:00Z",
							},
						},
					},
				},
				PRs: []int{456},
			},
		},
	}

	params := LangDashboardParams{
		LangCode:   "pl",
		ItemsTypes: defaultItemsTypes(),
		SortBy:     SortByFilename,
		SortOrder:  SortOrderAsc,
	}

	input := LangDashboardBuildInput{
		PagePath:  "/lang/pl",
		Dashboard: dash,
		Params:    params,
	}

	viewModel := BuildLangDashboardPageVM(input)

	if viewModel.LangCode != "pl" {
		t.Fatalf("expected lang code pl, got %q", viewModel.LangCode)
	}

	if !viewModel.ShowPanel {
		t.Fatal("expected ShowPanel to be true")
	}

	if viewModel.PageURL != "/lang/pl" {
		t.Fatalf("expected PageURL /lang/pl, got %q", viewModel.PageURL)
	}

	if !viewModel.Filters.ItemsWithEnUpdates.Active {
		t.Fatal("expected ItemsWithEnUpdates to be active")
	}

	if !viewModel.Filters.ItemsWithPR.Active {
		t.Fatal("expected ItemsWithPR to be active")
	}

	if !viewModel.Filters.ItemsEnFileNoLongerExists.Active {
		t.Fatal("expected ItemsEnFileNoLongerExists to be active")
	}

	if viewModel.Filters.ItemsLangFileMissing.Active {
		t.Fatal("expected ItemsLangFileMissing to be inactive")
	}

	if viewModel.Filters.ItemsLangFileUpToDate.Active {
		t.Fatal("expected ItemsLangFileUpToDate to be inactive")
	}

	if len(viewModel.Table.Rows) != 1 {
		t.Fatalf("expected 1 row, got %d", len(viewModel.Table.Rows))
	}

	row := viewModel.Table.Rows[0]
	if row.Filename.DisplayPath != "content/pl/test.md" {
		t.Fatalf("expected DisplayPath content/pl/test.md, got %q", row.Filename.DisplayPath)
	}

	if row.Filename.GithubURL != "https://github.com/kubernetes/website/blob/main/content/pl/test.md" {
		t.Fatalf("unexpected GithubURL: %q", row.Filename.GithubURL)
	}

	if row.Filename.DetailsURL != "/lang/pl?filename=content%2Fpl%2Ftest.md" {
		t.Fatalf("unexpected DetailsURL: %q", row.Filename.DetailsURL)
	}

	if row.Status.Text != "en-file-updated" {
		t.Fatalf("expected status en-file-updated, got %q", row.Status.Text)
	}

	if !row.Updates.HasUpdates {
		t.Fatal("expected HasUpdates to be true")
	}

	if row.Updates.LastUpdateText != "2023-01-02" {
		t.Fatalf("expected LastUpdateText 2023-01-02, got %q", row.Updates.LastUpdateText)
	}

	if len(row.Updates.Items) != 1 {
		t.Fatalf("expected 1 update item, got %d", len(row.Updates.Items))
	}

	if row.Updates.Items[0].CommitText != "update en" {
		t.Fatalf("expected commit text update en, got %q", row.Updates.Items[0].CommitText)
	}

	if row.Updates.Items[0].CommitURL != "https://github.com/kubernetes/website/commit/abc123" {
		t.Fatalf("unexpected CommitURL: %q", row.Updates.Items[0].CommitURL)
	}

	if len(row.PRs.Links) != 1 {
		t.Fatalf("expected 1 PR link, got %d", len(row.PRs.Links))
	}

	if row.PRs.Links[0].Text != "#456" {
		t.Fatalf("expected PR text #456, got %q", row.PRs.Links[0].Text)
	}

	if row.PRs.Links[0].URL != "https://github.com/kubernetes/website/pull/456" {
		t.Fatalf("unexpected PR URL: %q", row.PRs.Links[0].URL)
	}
}

func TestShouldShowPanel(t *testing.T) {
	t.Parallel()

	if !shouldShowPanel(LangDashboardParams{Filename: ""}) {
		t.Fatal("expected shouldShowPanel to return true for empty filename")
	}

	if shouldShowPanel(LangDashboardParams{Filename: "some-file.md"}) {
		t.Fatal("expected shouldShowPanel to return false for non-empty filename")
	}
}
