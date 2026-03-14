//nolint:testpackage,dupl
package web

import (
	"net/url"
	"testing"
)

func TestParseLangDashboardParams(t *testing.T) {
	t.Parallel()

	t.Run("parses valid values", func(t *testing.T) {
		t.Parallel()

		values := url.Values{}
		values.Set("itemsType", "with-update")
		values.Set("filename", "content/pl/test.md")
		values.Set("filepath", "content/pl")
		values.Set("sort", "status")
		values.Set("order", "desc")

		got := ParseLangDashboardParams("pl", values)

		if got.LangCode != "pl" {
			t.Fatalf("expected LangCode pl, got %q", got.LangCode)
		}

		if got.ItemsType != ItemsTypeWithUpdate {
			t.Fatalf("expected ItemsType %q, got %q", ItemsTypeWithUpdate, got.ItemsType)
		}

		if got.Filename != "content/pl/test.md" {
			t.Fatalf("expected Filename content/pl/test.md, got %q", got.Filename)
		}

		if got.Filepath != "content/pl" {
			t.Fatalf("expected Filepath content/pl, got %q", got.Filepath)
		}

		if got.SortBy != SortByStatus {
			t.Fatalf("expected SortBy %q, got %q", SortByStatus, got.SortBy)
		}

		if got.SortOrder != SortOrderDesc {
			t.Fatalf("expected SortOrder %q, got %q", SortOrderDesc, got.SortOrder)
		}
	})

	t.Run("uses defaults for invalid values", func(t *testing.T) {
		t.Parallel()

		values := url.Values{}
		values.Set("itemsType", "nope")
		values.Set("sort", "bad")
		values.Set("order", "sideways")

		got := ParseLangDashboardParams("pl", values)

		if got.LangCode != "pl" {
			t.Fatalf("expected LangCode pl, got %q", got.LangCode)
		}

		if got.ItemsType != ItemsTypeAll {
			t.Fatalf("expected ItemsType %q, got %q", ItemsTypeAll, got.ItemsType)
		}

		if got.Filename != "" {
			t.Fatalf("expected empty Filename, got %q", got.Filename)
		}

		if got.Filepath != "" {
			t.Fatalf("expected empty Filepath, got %q", got.Filepath)
		}

		if got.SortBy != SortByFilename {
			t.Fatalf("expected SortBy %q, got %q", SortByFilename, got.SortBy)
		}

		if got.SortOrder != SortOrderAsc {
			t.Fatalf("expected SortOrder %q, got %q", SortOrderAsc, got.SortOrder)
		}
	})

	t.Run("trims spaces", func(t *testing.T) {
		t.Parallel()

		values := url.Values{}
		values.Set("itemsType", " with-pr ")
		values.Set("filename", " content/pl/test.md ")
		values.Set("filepath", " content/pl ")
		values.Set("sort", " updates ")
		values.Set("order", " desc ")

		got := ParseLangDashboardParams(" pl ", values)

		if got.LangCode != "pl" {
			t.Fatalf("expected LangCode pl, got %q", got.LangCode)
		}

		if got.ItemsType != ItemsTypeWithPR {
			t.Fatalf("expected ItemsType %q, got %q", ItemsTypeWithPR, got.ItemsType)
		}

		if got.Filename != "content/pl/test.md" {
			t.Fatalf("expected Filename content/pl/test.md, got %q", got.Filename)
		}

		if got.Filepath != "content/pl" {
			t.Fatalf("expected Filepath content/pl, got %q", got.Filepath)
		}

		if got.SortBy != SortByUpdates {
			t.Fatalf("expected SortBy %q, got %q", SortByUpdates, got.SortBy)
		}

		if got.SortOrder != SortOrderDesc {
			t.Fatalf("expected SortOrder %q, got %q", SortOrderDesc, got.SortOrder)
		}
	})
}
