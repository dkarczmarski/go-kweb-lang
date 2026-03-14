package dashboard_test

import (
	"testing"

	"github.com/dkarczmarski/go-kweb-lang/dashboard"
)

func TestStore_ReadWriteDashboard(t *testing.T) {
	t.Parallel()

	store := dashboard.NewStore(newFakeCacheStorage())

	dashboardValue := dashboard.Dashboard{
		LangCode: "pl",
		Items: []dashboard.Item{
			{PRs: []int{1, 2}},
		},
	}

	err := store.WriteDashboard(dashboardValue)
	if err != nil {
		t.Fatalf("WriteDashboard returned error: %v", err)
	}

	got, err := store.ReadDashboard("pl")
	if err != nil {
		t.Fatalf("ReadDashboard returned error: %v", err)
	}

	if got.LangCode != dashboardValue.LangCode {
		t.Fatalf("expected lang code %q, got %q", dashboardValue.LangCode, got.LangCode)
	}

	if len(got.Items) != 1 {
		t.Fatalf("expected 1 item, got %d", len(got.Items))
	}

	if len(got.Items[0].PRs) != 2 || got.Items[0].PRs[0] != 1 || got.Items[0].PRs[1] != 2 {
		t.Fatalf("unexpected PRs: %#v", got.Items[0].PRs)
	}
}

func TestStore_ReadDashboard_NotFound(t *testing.T) {
	t.Parallel()

	store := dashboard.NewStore(newFakeCacheStorage())

	got, err := store.ReadDashboard("pl")
	if err != nil {
		t.Fatalf("ReadDashboard returned error: %v", err)
	}

	if got.LangCode != "" {
		t.Fatalf("expected empty LangCode, got %q", got.LangCode)
	}

	if len(got.Items) != 0 {
		t.Fatalf("expected 0 items, got %d", len(got.Items))
	}
}

func TestStore_ReadWriteDashboardIndex(t *testing.T) {
	t.Parallel()

	store := dashboard.NewStore(newFakeCacheStorage())

	index := dashboard.LangIndex{
		Items: []dashboard.LangIndexItem{
			{LangCode: "pl"},
			{LangCode: "de"},
		},
	}

	err := store.WriteDashboardIndex(index)
	if err != nil {
		t.Fatalf("WriteDashboardIndex returned error: %v", err)
	}

	got, err := store.ReadDashboardIndex()
	if err != nil {
		t.Fatalf("ReadDashboardIndex returned error: %v", err)
	}

	if len(got.Items) != 2 {
		t.Fatalf("expected 2 items, got %d", len(got.Items))
	}

	if got.Items[0].LangCode != "pl" {
		t.Fatalf("expected first lang code pl, got %q", got.Items[0].LangCode)
	}

	if got.Items[1].LangCode != "de" {
		t.Fatalf("expected second lang code de, got %q", got.Items[1].LangCode)
	}
}

type fakeCacheStorage struct {
	data map[string]any
}

func newFakeCacheStorage() *fakeCacheStorage {
	return &fakeCacheStorage{
		data: map[string]any{},
	}
}

func cacheDataKey(bucket, key string) string {
	return bucket + "::" + key
}

//nolint:forcetypeassert
func (s *fakeCacheStorage) Read(bucket, key string, buff any) (bool, error) {
	value, ok := s.data[cacheDataKey(bucket, key)]
	if !ok {
		return false, nil
	}

	switch out := buff.(type) {
	case *dashboard.Dashboard:
		*out = *(value.(*dashboard.Dashboard))
	case *dashboard.LangIndex:
		*out = *(value.(*dashboard.LangIndex))
	default:
		panic("unsupported buffer type")
	}

	return true, nil
}

func (s *fakeCacheStorage) Write(bucket, key string, data any) error {
	s.data[cacheDataKey(bucket, key)] = data

	return nil
}
