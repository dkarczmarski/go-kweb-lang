//nolint:goconst
package filepairs_test

import (
	"errors"
	"testing"

	"github.com/dkarczmarski/go-kweb-lang/filepairs"
)

type fakeContentFilesLister struct {
	listFilesFunc func(path string) ([]string, error)
}

func (f fakeContentFilesLister) ListFiles(path string) ([]string, error) {
	return f.listFilesFunc(path)
}

func TestNewContentPairProvider(t *testing.T) {
	t.Parallel()

	provider := filepairs.NewContentPairProvider(fakeContentFilesLister{
		listFilesFunc: func(_ string) ([]string, error) {
			return nil, nil
		},
	})

	if provider == nil {
		t.Fatal("NewContentPairProvider() returned nil")
	}
}

func TestContentPairProvider_Name(t *testing.T) {
	t.Parallel()

	provider := filepairs.NewContentPairProvider(fakeContentFilesLister{
		listFilesFunc: func(_ string) ([]string, error) {
			return nil, nil
		},
	})

	if got := provider.Name(); got != "content" {
		t.Fatalf("Name() = %q, want %q", got, "content")
	}
}

func TestContentPairProvider_ListPairs(t *testing.T) {
	t.Parallel()

	provider := filepairs.NewContentPairProvider(fakeContentFilesLister{
		listFilesFunc: func(path string) ([]string, error) {
			switch path {
			case "content/en":
				return []string{
					"docs/page1.md",
					"OWNERS",
					"docs/page2.md",
				}, nil
			case "content/pl":
				return []string{
					"docs/page2.md",
					"OWNERS",
					"docs/page3.md",
				}, nil
			default:
				t.Fatalf("ListFiles() path = %q, want one of %q or %q", path, "content/en", "content/pl")

				return nil, nil
			}
		},
	})

	got, err := provider.ListPairs("pl")
	if err != nil {
		t.Fatalf("ListPairs() error = %v", err)
	}

	want := []filepairs.Pair{
		{
			EnPath:   "content/en/docs/page2.md",
			LangPath: "content/pl/docs/page2.md",
		},
		{
			EnPath:   "content/en/docs/page3.md",
			LangPath: "content/pl/docs/page3.md",
		},
		{
			EnPath:   "content/en/docs/page1.md",
			LangPath: "content/pl/docs/page1.md",
		},
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

func TestContentPairProvider_ListPairs_ListFilesEnError(t *testing.T) {
	t.Parallel()

	wantErr := errors.New("boom")

	provider := filepairs.NewContentPairProvider(fakeContentFilesLister{
		listFilesFunc: func(path string) ([]string, error) {
			switch path {
			case "content/pl":
				return []string{"docs/page1.md"}, nil
			case "content/en":
				return nil, wantErr
			default:
				t.Fatalf("ListFiles() path = %q, want one of %q or %q", path, "content/pl", "content/en")

				return nil, nil
			}
		},
	})

	_, err := provider.ListPairs("pl")
	if err == nil {
		t.Fatal("ListPairs() error = nil, want error")
	}
}

func TestContentPairProvider_ListPairs_ListFilesLangError(t *testing.T) {
	t.Parallel()

	wantErr := errors.New("boom")

	provider := filepairs.NewContentPairProvider(fakeContentFilesLister{
		listFilesFunc: func(path string) ([]string, error) {
			switch path {
			case "content/pl":
				return nil, wantErr
			default:
				t.Fatalf("ListFiles() path = %q, want %q", path, "content/pl")

				return nil, nil
			}
		},
	})

	_, err := provider.ListPairs("pl")
	if err == nil {
		t.Fatal("ListPairs() error = nil, want error")
	}
}

func TestContentPairProvider_ListPairs_InvalidListedEnPath(t *testing.T) {
	t.Parallel()

	provider := filepairs.NewContentPairProvider(fakeContentFilesLister{
		listFilesFunc: func(path string) ([]string, error) {
			switch path {
			case "content/pl":
				return nil, nil
			case "content/en":
				return []string{""}, nil
			default:
				t.Fatalf("ListFiles() path = %q, want one of %q or %q", path, "content/pl", "content/en")

				return nil, nil
			}
		},
	})

	_, err := provider.ListPairs("pl")
	if err == nil {
		t.Fatal("ListPairs() error = nil, want error")
	}
}

func TestContentPairProvider_ListPairs_InvalidListedLangPath(t *testing.T) {
	t.Parallel()

	provider := filepairs.NewContentPairProvider(fakeContentFilesLister{
		listFilesFunc: func(path string) ([]string, error) {
			switch path {
			case "content/pl":
				return []string{""}, nil
			case "content/en":
				return []string{"docs/page1.md"}, nil
			default:
				t.Fatalf("ListFiles() path = %q, want one of %q or %q", path, "content/pl", "content/en")

				return nil, nil
			}
		},
	})

	_, err := provider.ListPairs("pl")
	if err == nil {
		t.Fatal("ListPairs() error = nil, want error")
	}
}
