//nolint:testpackage
package web

import "testing"

func TestDashboardURLBuilder(t *testing.T) {
	t.Parallel()

	baseParams := LangDashboardParams{
		LangCode:  "pl",
		ItemsType: ItemsTypeAll,
		SortBy:    SortByFilename,
		SortOrder: SortOrderAsc,
	}
	builder := NewDashboardURLBuilder("/lang/pl", baseParams)

	t.Run("Current", func(t *testing.T) {
		t.Parallel()

		got := builder.Current()
		want := "/lang/pl"

		if got != want {
			t.Fatalf("expected %q, got %q", want, got)
		}
	})

	t.Run("WithItemsType", func(t *testing.T) {
		t.Parallel()

		got := builder.WithItemsType(ItemsTypeWithUpdate)
		want := "/lang/pl?itemsType=with-update"

		if got != want {
			t.Fatalf("expected %q, got %q", want, got)
		}
	})

	t.Run("WithFilename", func(t *testing.T) {
		t.Parallel()

		got := builder.WithFilename("content/pl/_index.md")
		want := "/lang/pl?filename=content%2Fpl%2F_index.md"

		if got != want {
			t.Fatalf("expected %q, got %q", want, got)
		}
	})

	t.Run("WithoutFilename", func(t *testing.T) {
		t.Parallel()

		params := baseParams
		params.Filename = "some-file"
		builderWithFilename := NewDashboardURLBuilder("/lang/pl", params)

		got := builderWithFilename.WithoutFilename()
		want := "/lang/pl"

		if got != want {
			t.Fatalf("expected %q, got %q", want, got)
		}
	})

	t.Run("WithFilepath", func(t *testing.T) {
		t.Parallel()

		got := builder.WithFilepath("content/pl")
		want := "/lang/pl?filepath=content%2Fpl"

		if got != want {
			t.Fatalf("expected %q, got %q", want, got)
		}
	})

	t.Run("Sort same column toggles order", func(t *testing.T) {
		t.Parallel()

		got := builder.Sort(SortByFilename)
		want := "/lang/pl?order=desc"

		if got != want {
			t.Fatalf("expected %q, got %q", want, got)
		}

		params := baseParams
		params.SortOrder = SortOrderDesc
		builderDesc := NewDashboardURLBuilder("/lang/pl", params)

		got = builderDesc.Sort(SortByFilename)
		want = "/lang/pl"

		if got != want {
			t.Fatalf("expected %q, got %q", want, got)
		}
	})

	t.Run("Sort different column resets to asc", func(t *testing.T) {
		t.Parallel()

		params := baseParams
		params.SortOrder = SortOrderDesc
		builderDesc := NewDashboardURLBuilder("/lang/pl", params)

		got := builderDesc.Sort(SortByStatus)
		want := "/lang/pl?sort=status"

		if got != want {
			t.Fatalf("expected %q, got %q", want, got)
		}
	})

	t.Run("Omit defaults", func(t *testing.T) {
		t.Parallel()

		params := LangDashboardParams{
			LangCode:  "pl",
			ItemsType: ItemsTypeAll,
			SortBy:    SortByFilename,
			SortOrder: SortOrderAsc,
		}
		defaultBuilder := NewDashboardURLBuilder("/lang/pl", params)

		got := defaultBuilder.Current()
		want := "/lang/pl"

		if got != want {
			t.Fatalf("expected %q, got %q", want, got)
		}
	})
}
