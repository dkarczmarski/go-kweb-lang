//nolint:testpackage,dupl,goconst,gocyclo
package web

import (
	"testing"

	"github.com/dkarczmarski/go-kweb-lang/dashboard"
	"github.com/dkarczmarski/go-kweb-lang/git"
	"github.com/dkarczmarski/go-kweb-lang/gitseek"
)

func TestFilterAndSortItems(t *testing.T) {
	t.Parallel()

	items := []dashboard.Item{
		{
			FileInfo: gitseek.FileInfo{
				LangPath:   "content/pl/a.md",
				FileStatus: gitseek.StatusEnFileUpdated,
				LangLastCommit: git.CommitInfo{
					DateTime: "2023-01-01T12:00:00Z",
				},
				EnUpdates: []gitseek.EnUpdate{
					{Commit: git.CommitInfo{DateTime: "2023-01-02T12:00:00Z"}},
				},
			},
		},
		{
			FileInfo: gitseek.FileInfo{
				LangPath:   "content/pl/b.md",
				FileStatus: dashboard.StatusWaitingForReview,
				LangLastCommit: git.CommitInfo{
					DateTime: "2023-01-05T12:00:00Z",
				},
			},
			PRs: []int{123},
		},
		{
			FileInfo: gitseek.FileInfo{
				LangPath:   "content/pl/c.md",
				FileStatus: gitseek.StatusLangFileUpToDate,
				LangLastCommit: git.CommitInfo{
					DateTime: "2023-01-03T12:00:00Z",
				},
			},
		},
		{
			FileInfo: gitseek.FileInfo{
				LangPath:   "content/pl/d.md",
				FileStatus: gitseek.StatusEnFileDoesNotExist,
				LangLastCommit: git.CommitInfo{
					DateTime: "2023-01-04T12:00:00Z",
				},
			},
		},
		{
			FileInfo: gitseek.FileInfo{
				LangPath:   "content/pl/e.md",
				FileStatus: gitseek.StatusEnFileNoLongerExists,
				LangLastCommit: git.CommitInfo{
					DateTime: "2023-01-06T12:00:00Z",
				},
			},
		},
	}

	t.Run("Filter by ItemsTypeWithEnUpdates", func(t *testing.T) {
		t.Parallel()

		params := LangDashboardParams{ItemsTypes: []string{ItemsTypeWithEnUpdates}}
		filtered := FilterAndSortItems(items, params)

		if len(filtered) != 1 {
			t.Fatalf("expected 1 item, got %d", len(filtered))
		}

		if filtered[0].LangPath != "content/pl/a.md" {
			t.Fatalf("expected content/pl/a.md, got %q", filtered[0].LangPath)
		}
	})

	t.Run("Filter by ItemsTypeWithPR", func(t *testing.T) {
		t.Parallel()

		params := LangDashboardParams{ItemsTypes: []string{ItemsTypeWithPR}}
		filtered := FilterAndSortItems(items, params)

		if len(filtered) != 1 {
			t.Fatalf("expected 1 item, got %d", len(filtered))
		}

		if filtered[0].LangPath != "content/pl/b.md" {
			t.Fatalf("expected content/pl/b.md, got %q", filtered[0].LangPath)
		}
	})

	t.Run("Filter by ItemsTypeEnFileDoesNotExist", func(t *testing.T) {
		t.Parallel()

		params := LangDashboardParams{ItemsTypes: []string{ItemsTypeEnFileDoesNotExist}}
		filtered := FilterAndSortItems(items, params)

		if len(filtered) != 1 {
			t.Fatalf("expected 1 item, got %d", len(filtered))
		}

		if filtered[0].LangPath != "content/pl/d.md" {
			t.Fatalf("expected content/pl/d.md, got %q", filtered[0].LangPath)
		}
	})

	t.Run("Filter by ItemsTypeEnFileNoLongerExists", func(t *testing.T) {
		t.Parallel()

		params := LangDashboardParams{ItemsTypes: []string{ItemsTypeEnFileNoLongerExists}}
		filtered := FilterAndSortItems(items, params)

		if len(filtered) != 1 {
			t.Fatalf("expected 1 item, got %d", len(filtered))
		}

		if filtered[0].LangPath != "content/pl/e.md" {
			t.Fatalf("expected content/pl/e.md, got %q", filtered[0].LangPath)
		}
	})

	t.Run("Filter by ItemsTypeWaitingForReview", func(t *testing.T) {
		t.Parallel()

		params := LangDashboardParams{ItemsTypes: []string{ItemsTypeWaitingForReview}}
		filtered := FilterAndSortItems(items, params)

		if len(filtered) != 1 {
			t.Fatalf("expected 1 item, got %d", len(filtered))
		}

		if filtered[0].LangPath != "content/pl/b.md" {
			t.Fatalf("expected content/pl/b.md, got %q", filtered[0].LangPath)
		}
	})

	t.Run("Filter by ItemsTypeLangFileUpToDate", func(t *testing.T) {
		t.Parallel()

		params := LangDashboardParams{ItemsTypes: []string{ItemsTypeLangFileUpToDate}}
		filtered := FilterAndSortItems(items, params)

		if len(filtered) != 1 {
			t.Fatalf("expected 1 item, got %d", len(filtered))
		}

		if filtered[0].LangPath != "content/pl/c.md" {
			t.Fatalf("expected content/pl/c.md, got %q", filtered[0].LangPath)
		}
	})

	t.Run("Filter by multiple items types uses or semantics", func(t *testing.T) {
		t.Parallel()

		params := LangDashboardParams{
			ItemsTypes: []string{
				ItemsTypeWithPR,
				ItemsTypeEnFileNoLongerExists,
			},
		}
		filtered := FilterAndSortItems(items, params)

		if len(filtered) != 2 {
			t.Fatalf("expected 2 items, got %d", len(filtered))
		}

		if filtered[0].LangPath != "content/pl/b.md" {
			t.Fatalf("expected first item content/pl/b.md, got %q", filtered[0].LangPath)
		}

		if filtered[1].LangPath != "content/pl/e.md" {
			t.Fatalf("expected second item content/pl/e.md, got %q", filtered[1].LangPath)
		}
	})

	t.Run("Filter by Filepath", func(t *testing.T) {
		t.Parallel()

		params := LangDashboardParams{
			ItemsTypes: defaultItemsTypes(),
			Filepath:   "a.md",
		}
		filtered := FilterAndSortItems(items, params)

		if len(filtered) != 1 {
			t.Fatalf("expected 1 item, got %d", len(filtered))
		}

		if filtered[0].LangPath != "content/pl/a.md" {
			t.Fatalf("expected content/pl/a.md, got %q", filtered[0].LangPath)
		}
	})

	t.Run("Sort by Filename desc", func(t *testing.T) {
		t.Parallel()

		params := LangDashboardParams{
			ItemsTypes: defaultItemsTypes(),
			SortBy:     SortByFilename,
			SortOrder:  SortOrderDesc,
		}
		sorted := FilterAndSortItems(items, params)

		if sorted[0].LangPath != "content/pl/e.md" {
			t.Fatalf("expected first path content/pl/e.md, got %q", sorted[0].LangPath)
		}

		if sorted[1].LangPath != "content/pl/d.md" {
			t.Fatalf("expected second path content/pl/d.md, got %q", sorted[1].LangPath)
		}

		if sorted[2].LangPath != "content/pl/c.md" {
			t.Fatalf("expected third path content/pl/c.md, got %q", sorted[2].LangPath)
		}

		if sorted[3].LangPath != "content/pl/b.md" {
			t.Fatalf("expected fourth path content/pl/b.md, got %q", sorted[3].LangPath)
		}

		if sorted[4].LangPath != "content/pl/a.md" {
			t.Fatalf("expected fifth path content/pl/a.md, got %q", sorted[4].LangPath)
		}
	})

	t.Run("Sort by Status", func(t *testing.T) {
		t.Parallel()

		params := LangDashboardParams{
			ItemsTypes: defaultItemsTypes(),
			SortBy:     SortByStatus,
			SortOrder:  SortOrderAsc,
		}
		sorted := FilterAndSortItems(items, params)

		if sorted[0].FileStatus != ItemsTypeEnFileDoesNotExist {
			t.Fatalf("expected first status en-file-does-not-exist, got %q", sorted[0].FileStatus)
		}

		if sorted[1].FileStatus != ItemsTypeEnFileNoLongerExists {
			t.Fatalf("expected second status en-file-no-longer-exists, got %q", sorted[1].FileStatus)
		}

		if sorted[2].FileStatus != gitseek.StatusEnFileUpdated {
			t.Fatalf("expected third status en-file-updated, got %q", sorted[2].FileStatus)
		}

		if sorted[3].FileStatus != gitseek.StatusLangFileUpToDate {
			t.Fatalf("expected fourth status up-to-date, got %q", sorted[3].FileStatus)
		}

		if sorted[4].FileStatus != ItemsTypeWaitingForReview {
			t.Fatalf("expected fifth status waiting-for-review, got %q", sorted[4].FileStatus)
		}
	})

	t.Run("Sort by Updates asc uses latest en update date", func(t *testing.T) {
		t.Parallel()

		itemsWithUpdates := []dashboard.Item{
			{
				FileInfo: gitseek.FileInfo{
					LangPath: "content/pl/a.md",
					EnUpdates: []gitseek.EnUpdate{
						{Commit: git.CommitInfo{DateTime: "2023-01-10T12:00:00Z"}},
						{Commit: git.CommitInfo{DateTime: "2023-01-02T12:00:00Z"}},
					},
				},
			},
			{
				FileInfo: gitseek.FileInfo{
					LangPath: "content/pl/b.md",
					EnUpdates: []gitseek.EnUpdate{
						{Commit: git.CommitInfo{DateTime: "2023-01-05T12:00:00Z"}},
					},
				},
			},
			{
				FileInfo: gitseek.FileInfo{
					LangPath:   "content/pl/c.md",
					FileStatus: gitseek.StatusEnFileNoLongerExists,
				},
			},
		}

		params := LangDashboardParams{
			ItemsTypes: []string{
				ItemsTypeWithEnUpdates,
				ItemsTypeEnFileNoLongerExists,
			},
			SortBy:    SortByUpdates,
			SortOrder: SortOrderAsc,
		}
		sorted := FilterAndSortItems(itemsWithUpdates, params)

		if sorted[0].LangPath != "content/pl/c.md" {
			t.Fatalf("expected first path content/pl/c.md, got %q", sorted[0].LangPath)
		}

		if sorted[1].LangPath != "content/pl/b.md" {
			t.Fatalf("expected second path content/pl/b.md, got %q", sorted[1].LangPath)
		}

		if sorted[2].LangPath != "content/pl/a.md" {
			t.Fatalf("expected third path content/pl/a.md, got %q", sorted[2].LangPath)
		}
	})

	t.Run("Sort by Updates desc uses latest en update date", func(t *testing.T) {
		t.Parallel()

		itemsWithUpdates := []dashboard.Item{
			{
				FileInfo: gitseek.FileInfo{
					LangPath: "content/pl/a.md",
					EnUpdates: []gitseek.EnUpdate{
						{Commit: git.CommitInfo{DateTime: "2023-01-10T12:00:00Z"}},
						{Commit: git.CommitInfo{DateTime: "2023-01-02T12:00:00Z"}},
					},
				},
			},
			{
				FileInfo: gitseek.FileInfo{
					LangPath: "content/pl/b.md",
					EnUpdates: []gitseek.EnUpdate{
						{Commit: git.CommitInfo{DateTime: "2023-01-05T12:00:00Z"}},
					},
				},
			},
			{
				FileInfo: gitseek.FileInfo{
					LangPath:   "content/pl/c.md",
					FileStatus: gitseek.StatusEnFileNoLongerExists,
				},
			},
		}

		params := LangDashboardParams{
			ItemsTypes: []string{
				ItemsTypeWithEnUpdates,
				ItemsTypeEnFileNoLongerExists,
			},
			SortBy:    SortByUpdates,
			SortOrder: SortOrderDesc,
		}
		sorted := FilterAndSortItems(itemsWithUpdates, params)

		if sorted[0].LangPath != "content/pl/a.md" {
			t.Fatalf("expected first path content/pl/a.md, got %q", sorted[0].LangPath)
		}

		if sorted[1].LangPath != "content/pl/b.md" {
			t.Fatalf("expected second path content/pl/b.md, got %q", sorted[1].LangPath)
		}

		if sorted[2].LangPath != "content/pl/c.md" {
			t.Fatalf("expected third path content/pl/c.md, got %q", sorted[2].LangPath)
		}
	})
}

func TestLatestEnUpdateDate(t *testing.T) {
	t.Parallel()

	t.Run("returns empty string when there are no updates", func(t *testing.T) {
		t.Parallel()

		item := dashboard.Item{
			FileInfo: gitseek.FileInfo{},
		}

		got := latestEnUpdateDate(item)
		if got != "" {
			t.Fatalf("expected empty string, got %q", got)
		}
	})

	t.Run("returns latest datetime from en updates", func(t *testing.T) {
		t.Parallel()

		item := dashboard.Item{
			FileInfo: gitseek.FileInfo{
				EnUpdates: []gitseek.EnUpdate{
					{Commit: git.CommitInfo{DateTime: "2023-01-02T12:00:00Z"}},
					{Commit: git.CommitInfo{DateTime: "2023-01-10T12:00:00Z"}},
					{Commit: git.CommitInfo{DateTime: "2023-01-05T12:00:00Z"}},
				},
			},
		}

		got := latestEnUpdateDate(item)
		if got != "2023-01-10T12:00:00Z" {
			t.Fatalf("expected latest datetime 2023-01-10T12:00:00Z, got %q", got)
		}
	})
}
