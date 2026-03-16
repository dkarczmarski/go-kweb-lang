//nolint:goconst
package filepairs_test

import (
	"errors"
	"testing"

	"github.com/dkarczmarski/go-kweb-lang/filepairs"
)

type fakeI18NFileChecker struct {
	fileExistsFunc func(path string) (bool, error)
}

func (f fakeI18NFileChecker) FileExists(path string) (bool, error) {
	return f.fileExistsFunc(path)
}

func TestNewI18NPairProvider(t *testing.T) {
	t.Parallel()

	provider := filepairs.NewI18NPairProvider(fakeI18NFileChecker{
		fileExistsFunc: func(_ string) (bool, error) {
			return false, nil
		},
	})

	if provider == nil {
		t.Fatal("NewI18NPairProvider() returned nil")
	}
}

func TestI18NPairProvider_Name(t *testing.T) {
	t.Parallel()

	provider := filepairs.NewI18NPairProvider(fakeI18NFileChecker{
		fileExistsFunc: func(_ string) (bool, error) {
			return false, nil
		},
	})

	if got := provider.Name(); got != "i18n" {
		t.Fatalf("Name() = %q, want %q", got, "i18n")
	}
}

func TestI18NPairProvider_ListPairs_LangFileExists(t *testing.T) {
	t.Parallel()

	provider := filepairs.NewI18NPairProvider(fakeI18NFileChecker{
		fileExistsFunc: func(path string) (bool, error) {
			switch path {
			case "i18n/pl/pl.toml":
				return true, nil
			case "i18n/en/en.toml":
				t.Fatal("FileExists() should not check EN path when lang file exists")

				return false, nil
			default:
				t.Fatalf("FileExists() path = %q, want %q", path, "i18n/pl/pl.toml")

				return false, nil
			}
		},
	})

	got, err := provider.ListPairs("pl")
	if err != nil {
		t.Fatalf("ListPairs() error = %v", err)
	}

	want := []filepairs.Pair{
		{
			EnPath:   "i18n/en/en.toml",
			LangPath: "i18n/pl/pl.toml",
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

func TestI18NPairProvider_ListPairs_FallbackToEn(t *testing.T) {
	t.Parallel()

	provider := filepairs.NewI18NPairProvider(fakeI18NFileChecker{
		fileExistsFunc: func(path string) (bool, error) {
			switch path {
			case "i18n/pl/pl.toml":
				return false, nil
			case "i18n/en/en.toml":
				return true, nil
			default:
				t.Fatalf("FileExists() unexpected path = %q", path)

				return false, nil
			}
		},
	})

	got, err := provider.ListPairs("pl")
	if err != nil {
		t.Fatalf("ListPairs() error = %v", err)
	}

	want := []filepairs.Pair{
		{
			EnPath:   "i18n/en/en.toml",
			LangPath: "i18n/pl/pl.toml",
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

func TestI18NPairProvider_ListPairs_NoFiles(t *testing.T) {
	t.Parallel()

	provider := filepairs.NewI18NPairProvider(fakeI18NFileChecker{
		fileExistsFunc: func(path string) (bool, error) {
			switch path {
			case "i18n/pl/pl.toml", "i18n/en/en.toml":
				return false, nil
			default:
				t.Fatalf("FileExists() unexpected path = %q", path)

				return false, nil
			}
		},
	})

	got, err := provider.ListPairs("pl")
	if err != nil {
		t.Fatalf("ListPairs() error = %v", err)
	}

	if got == nil {
		t.Fatal("ListPairs() = nil, want empty slice")
	}

	if len(got) != 0 {
		t.Fatalf("ListPairs() len = %d, want 0; got=%#v", len(got), got)
	}
}

func TestI18NPairProvider_ListPairs_LangFileExistsError(t *testing.T) {
	t.Parallel()

	wantErr := errors.New("boom")

	provider := filepairs.NewI18NPairProvider(fakeI18NFileChecker{
		fileExistsFunc: func(path string) (bool, error) {
			if path != "i18n/pl/pl.toml" {
				t.Fatalf("FileExists() path = %q, want %q", path, "i18n/pl/pl.toml")
			}

			return false, wantErr
		},
	})

	_, err := provider.ListPairs("pl")
	if err == nil {
		t.Fatal("ListPairs() error = nil, want error")
	}
}

func TestI18NPairProvider_ListPairs_EnFileExistsError(t *testing.T) {
	t.Parallel()

	wantErr := errors.New("boom")

	provider := filepairs.NewI18NPairProvider(fakeI18NFileChecker{
		fileExistsFunc: func(path string) (bool, error) {
			switch path {
			case "i18n/pl/pl.toml":
				return false, nil
			case "i18n/en/en.toml":
				return false, wantErr
			default:
				t.Fatalf("FileExists() unexpected path = %q", path)

				return false, nil
			}
		},
	})

	_, err := provider.ListPairs("pl")
	if err == nil {
		t.Fatal("ListPairs() error = nil, want error")
	}
}
