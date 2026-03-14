//nolint:testpackage,dupl,goconst
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
				FileStatus: "en-file-updated",
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
				FileStatus: "waiting-for-review",
				LangLastCommit: git.CommitInfo{
					DateTime: "2023-01-05T12:00:00Z",
				},
			},
			PRs: []int{123},
		},
		{
			FileInfo: gitseek.FileInfo{
				LangPath:   "content/pl/c.md",
				FileStatus: "up-to-date",
				LangLastCommit: git.CommitInfo{
					DateTime: "2023-01-03T12:00:00Z",
				},
			},
		},
	}

	t.Run("Filter by ItemsTypeWithUpdate", func(t *testing.T) {
		t.Parallel()

		params := LangDashboardParams{ItemsType: ItemsTypeWithUpdate}
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

		params := LangDashboardParams{ItemsType: ItemsTypeWithPR}
		filtered := FilterAndSortItems(items, params)

		if len(filtered) != 1 {
			t.Fatalf("expected 1 item, got %d", len(filtered))
		}

		if filtered[0].LangPath != "content/pl/b.md" {
			t.Fatalf("expected content/pl/b.md, got %q", filtered[0].LangPath)
		}
	})

	t.Run("Filter by Filepath", func(t *testing.T) {
		t.Parallel()

		params := LangDashboardParams{Filepath: "a.md"}
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

		params := LangDashboardParams{SortBy: SortByFilename, SortOrder: SortOrderDesc}
		sorted := FilterAndSortItems(items, params)

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

	t.Run("Sort by Status", func(t *testing.T) {
		t.Parallel()

		params := LangDashboardParams{SortBy: SortByStatus, SortOrder: SortOrderAsc}
		sorted := FilterAndSortItems(items, params)

		if sorted[0].FileStatus != "en-file-updated" {
			t.Fatalf("expected first status en-file-updated, got %q", sorted[0].FileStatus)
		}

		if sorted[1].FileStatus != "up-to-date" {
			t.Fatalf("expected second status up-to-date, got %q", sorted[1].FileStatus)
		}

		if sorted[2].FileStatus != "waiting-for-review" {
			t.Fatalf("expected third status waiting-for-review, got %q", sorted[2].FileStatus)
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
					LangPath: "content/pl/c.md",
				},
			},
		}

		params := LangDashboardParams{SortBy: SortByUpdates, SortOrder: SortOrderAsc}
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
					LangPath: "content/pl/c.md",
				},
			},
		}

		params := LangDashboardParams{SortBy: SortByUpdates, SortOrder: SortOrderDesc}
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
