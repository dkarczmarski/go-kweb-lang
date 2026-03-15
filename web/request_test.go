//nolint:testpackage,dupl
package web

import (
	"net/url"
	"reflect"
	"testing"
)

func TestParseLangDashboardParams(t *testing.T) {
	t.Parallel()

	t.Run("parses valid values", func(t *testing.T) {
		t.Parallel()

		values := url.Values{}
		values.Add("itemsType", ItemsTypeWithEnUpdates)
		values.Add("itemsType", ItemsTypeWithPR)
		values.Add("itemsType", ItemsTypeEnFileNoLongerExists)
		values.Set("filename", "content/pl/test.md")
		values.Set("filepath", "content/pl")
		values.Set("sort", SortByStatus)
		values.Set("order", SortOrderDesc)

		got := ParseLangDashboardParams("pl", values)

		if got.LangCode != "pl" {
			t.Fatalf("expected LangCode pl, got %q", got.LangCode)
		}

		wantItemsTypes := []string{
			ItemsTypeWithEnUpdates,
			ItemsTypeWithPR,
			ItemsTypeEnFileNoLongerExists,
		}
		if !reflect.DeepEqual(got.ItemsTypes, wantItemsTypes) {
			t.Fatalf("expected ItemsTypes %#v, got %#v", wantItemsTypes, got.ItemsTypes)
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
		values.Add("itemsType", "nope")
		values.Set("sort", "bad")
		values.Set("order", "sideways")

		got := ParseLangDashboardParams("pl", values)

		if got.LangCode != "pl" {
			t.Fatalf("expected LangCode pl, got %q", got.LangCode)
		}

		wantItemsTypes := defaultItemsTypes()
		if !reflect.DeepEqual(got.ItemsTypes, wantItemsTypes) {
			t.Fatalf("expected ItemsTypes %#v, got %#v", wantItemsTypes, got.ItemsTypes)
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

	t.Run("trims spaces and deduplicates values", func(t *testing.T) {
		t.Parallel()

		values := url.Values{}
		values.Add("itemsType", " "+ItemsTypeWithPR+" ")
		values.Add("itemsType", " "+ItemsTypeWithEnUpdates+" ")
		values.Add("itemsType", " "+ItemsTypeEnFileNoLongerExists+" ")
		values.Add("itemsType", " "+ItemsTypeWithPR+" ")
		values.Set("filename", " content/pl/test.md ")
		values.Set("filepath", " content/pl ")
		values.Set("sort", " "+SortByUpdates+" ")
		values.Set("order", " "+SortOrderDesc+" ")

		got := ParseLangDashboardParams(" pl ", values)

		if got.LangCode != "pl" {
			t.Fatalf("expected LangCode pl, got %q", got.LangCode)
		}

		wantItemsTypes := []string{
			ItemsTypeWithPR,
			ItemsTypeWithEnUpdates,
			ItemsTypeEnFileNoLongerExists,
		}
		if !reflect.DeepEqual(got.ItemsTypes, wantItemsTypes) {
			t.Fatalf("expected ItemsTypes %#v, got %#v", wantItemsTypes, got.ItemsTypes)
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

	t.Run("uses defaults when no items types are provided", func(t *testing.T) {
		t.Parallel()

		values := url.Values{}

		got := ParseLangDashboardParams("pl", values)

		wantItemsTypes := defaultItemsTypes()
		if !reflect.DeepEqual(got.ItemsTypes, wantItemsTypes) {
			t.Fatalf("expected ItemsTypes %#v, got %#v", wantItemsTypes, got.ItemsTypes)
		}
	})
}
