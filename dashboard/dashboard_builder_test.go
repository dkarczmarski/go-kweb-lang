//nolint:testpackage
package dashboard

import (
	"testing"

	"github.com/dkarczmarski/go-kweb-lang/gitseek"
	"github.com/dkarczmarski/go-kweb-lang/pullreq"
)

func TestBuildDashboard(t *testing.T) {
	t.Parallel()

	t.Run("adds prs to items matched by lang path", func(t *testing.T) {
		t.Parallel()

		seekerFileInfos := []gitseek.FileInfo{
			{
				LangPath:   "content/pl/a.md",
				FileStatus: "up-to-date",
			},
			{
				LangPath:   "content/pl/b.md",
				FileStatus: "en-file-updated",
			},
		}

		prIndex := pullreq.FilePRIndexData{
			"content/pl/b.md": {101, 102},
		}

		got := BuildDashboard("pl", seekerFileInfos, prIndex)

		if got.LangCode != "pl" {
			t.Fatalf("expected lang code pl, got %q", got.LangCode)
		}

		if len(got.Items) != 2 {
			t.Fatalf("expected 2 items, got %d", len(got.Items))
		}

		if got.Items[0].LangPath != "content/pl/a.md" {
			t.Fatalf("expected first item path content/pl/a.md, got %q", got.Items[0].LangPath)
		}

		if len(got.Items[0].PRs) != 0 {
			t.Fatalf("expected first item PRs to be empty, got %#v", got.Items[0].PRs)
		}

		if got.Items[1].LangPath != "content/pl/b.md" {
			t.Fatalf("expected second item path content/pl/b.md, got %q", got.Items[1].LangPath)
		}

		if len(got.Items[1].PRs) != 2 {
			t.Fatalf("expected second item to have 2 PRs, got %d", len(got.Items[1].PRs))
		}

		if got.Items[1].PRs[0] != 101 || got.Items[1].PRs[1] != 102 {
			t.Fatalf("unexpected PRs: %#v", got.Items[1].PRs)
		}
	})

	t.Run("adds waiting-for-review item for file existing only in pr index", func(t *testing.T) {
		t.Parallel()

		seekerFileInfos := []gitseek.FileInfo{
			{
				LangPath:   "content/pl/a.md",
				FileStatus: "up-to-date",
			},
		}

		prIndex := pullreq.FilePRIndexData{
			"content/pl/missing.md": {555},
		}

		got := BuildDashboard("pl", seekerFileInfos, prIndex)

		if len(got.Items) != 2 {
			t.Fatalf("expected 2 items, got %d", len(got.Items))
		}

		added := got.Items[1]

		if added.LangPath != "content/pl/missing.md" {
			t.Fatalf("expected added item path content/pl/missing.md, got %q", added.LangPath)
		}

		if added.FileStatus != StatusWaitingForReview {
			t.Fatalf("expected added item status %q, got %q", StatusWaitingForReview, added.FileStatus)
		}

		if len(added.PRs) != 1 || added.PRs[0] != 555 {
			t.Fatalf("unexpected added item PRs: %#v", added.PRs)
		}
	})

	t.Run("does not duplicate item when pr index path already exists in seeker file infos", func(t *testing.T) {
		t.Parallel()

		seekerFileInfos := []gitseek.FileInfo{
			{
				LangPath:   "content/pl/a.md",
				FileStatus: "up-to-date",
			},
		}

		prIndex := pullreq.FilePRIndexData{
			"content/pl/a.md": {123},
		}

		got := BuildDashboard("pl", seekerFileInfos, prIndex)

		if len(got.Items) != 1 {
			t.Fatalf("expected 1 item, got %d", len(got.Items))
		}

		if got.Items[0].LangPath != "content/pl/a.md" {
			t.Fatalf("expected path content/pl/a.md, got %q", got.Items[0].LangPath)
		}

		if len(got.Items[0].PRs) != 1 || got.Items[0].PRs[0] != 123 {
			t.Fatalf("unexpected PRs: %#v", got.Items[0].PRs)
		}
	})
}

func TestContainsItem(t *testing.T) {
	t.Parallel()

	items := []Item{
		{FileInfo: gitseek.FileInfo{LangPath: "content/pl/a.md"}},
		{FileInfo: gitseek.FileInfo{LangPath: "content/pl/b.md"}},
	}

	if !containsItem(items, "content/pl/a.md") {
		t.Fatal("expected containsItem to return true for content/pl/a.md")
	}

	if !containsItem(items, "content/pl/b.md") {
		t.Fatal("expected containsItem to return true for content/pl/b.md")
	}

	if containsItem(items, "content/pl/c.md") {
		t.Fatal("expected containsItem to return false for content/pl/c.md")
	}
}
