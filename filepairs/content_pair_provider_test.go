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
			if path != "content/pl" {
				t.Fatalf("ListFiles() path = %q, want %q", path, "content/pl")
			}

			return []string{
				"docs/page1.md",
				"OWNERS",
				"docs/page2.md",
			}, nil
		},
	})

	got, err := provider.ListPairs("pl")
	if err != nil {
		t.Fatalf("ListPairs() error = %v", err)
	}

	want := []filepairs.Pair{
		{
			EnPath:   "content/en/docs/page1.md",
			LangPath: "content/pl/docs/page1.md",
		},
		{
			EnPath:   "content/en/docs/page2.md",
			LangPath: "content/pl/docs/page2.md",
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

func TestContentPairProvider_ListPairs_ListFilesError(t *testing.T) {
	t.Parallel()

	wantErr := errors.New("boom")

	provider := filepairs.NewContentPairProvider(fakeContentFilesLister{
		listFilesFunc: func(_ string) ([]string, error) {
			return nil, wantErr
		},
	})

	_, err := provider.ListPairs("pl")
	if err == nil {
		t.Fatal("ListPairs() error = nil, want error")
	}
}

func TestContentPairProvider_ListPairs_InvalidListedPath(t *testing.T) {
	t.Parallel()

	provider := filepairs.NewContentPairProvider(fakeContentFilesLister{
		listFilesFunc: func(_ string) ([]string, error) {
			return []string{""}, nil
		},
	})

	_, err := provider.ListPairs("pl")
	if err == nil {
		t.Fatal("ListPairs() error = nil, want error")
	}
}
