package filepairs_test

import (
	"errors"
	"testing"

	"github.com/dkarczmarski/go-kweb-lang/filepairs"
)

type fakePairProvider struct {
	name          string
	listPairsFunc func(langCode string) ([]filepairs.Pair, error)
}

func (f fakePairProvider) Name() string {
	return f.name
}

func (f fakePairProvider) ListPairs(langCode string) ([]filepairs.Pair, error) {
	return f.listPairsFunc(langCode)
}

func TestPairProviders_ListPairs(t *testing.T) {
	t.Parallel()

	providers := filepairs.NewPairProviders(
		fakePairProvider{
			name: "first",
			listPairsFunc: func(langCode string) ([]filepairs.Pair, error) {
				if langCode != "pl" {
					t.Fatalf("langCode = %q, want %q", langCode, "pl")
				}

				return []filepairs.Pair{
					{EnPath: "content/en/a.md", LangPath: "content/pl/a.md"},
				}, nil
			},
		},
		fakePairProvider{
			name: "second",
			listPairsFunc: func(_ string) ([]filepairs.Pair, error) {
				return []filepairs.Pair{
					{EnPath: "content/en/b.md", LangPath: "content/pl/b.md"},
				}, nil
			},
		},
	)

	got, err := providers.ListPairs("pl")
	if err != nil {
		t.Fatalf("ListPairs() error = %v", err)
	}

	want := []filepairs.Pair{
		{EnPath: "content/en/a.md", LangPath: "content/pl/a.md"},
		{EnPath: "content/en/b.md", LangPath: "content/pl/b.md"},
	}

	if len(got) != len(want) {
		t.Fatalf("ListPairs() len = %d, want %d; got=%#v", len(got), len(want), got)
	}

	for i := range want {
		if got[i] != want[i] {
			t.Fatalf("ListPairs()[%d] = %#v, want %#v", i, got[i], want[i])
		}
	}
}

func TestPairProviders_ListPairs_ProviderError(t *testing.T) {
	t.Parallel()

	wantErr := errors.New("boom")

	providers := filepairs.NewPairProviders(
		fakePairProvider{
			name: "broken",
			listPairsFunc: func(_ string) ([]filepairs.Pair, error) {
				return nil, wantErr
			},
		},
	)

	_, err := providers.ListPairs("pl")
	if err == nil {
		t.Fatal("ListPairs() error = nil, want error")
	}
}
