//nolint:gosec
package langcnt_test

import (
	"os"
	"path/filepath"
	"reflect"
	"testing"

	"github.com/dkarczmarski/go-kweb-lang/langcnt"
)

func TestLangCodesProvider_LangCodes(t *testing.T) {
	t.Parallel()

	for _, tc := range []struct {
		name   string
		filter []string
		want   []string
	}{
		{
			name:   "returns all language directories except en",
			filter: nil,
			want:   []string{"fr", "pl", "uk"},
		},
		{
			name:   "returns only filtered language directories",
			filter: []string{"uk", "pl"},
			want:   []string{"pl", "uk"},
		},
		{
			name:   "ignores unknown language codes in filter",
			filter: []string{"uk", "de"},
			want:   []string{"uk"},
		},
		{
			name:   "returns empty list when filter matches nothing",
			filter: []string{"de", "es"},
			want:   []string{},
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			repoDir := t.TempDir()
			contentDir := filepath.Join(repoDir, "content")

			mustMkdir(t, filepath.Join(contentDir, "en"))
			mustMkdir(t, filepath.Join(contentDir, "pl"))
			mustMkdir(t, filepath.Join(contentDir, "fr"))
			mustMkdir(t, filepath.Join(contentDir, "uk"))
			mustWriteFile(t, filepath.Join(contentDir, "README.md"), "not a directory")

			provider := &langcnt.LangCodesProvider{
				RepoDir: repoDir,
			}
			provider.SetLangCodesFilter(tc.filter)

			got, err := provider.LangCodes()
			if err != nil {
				t.Fatalf("LangCodes returned error: %v", err)
			}

			if !reflect.DeepEqual(got, tc.want) {
				t.Fatalf("unexpected lang codes\nactual:   %v\nexpected: %v", got, tc.want)
			}
		})
	}
}

func TestLangCodesProvider_LangCodes_ReturnsErrorWhenContentDirDoesNotExist(t *testing.T) {
	t.Parallel()

	repoDir := t.TempDir()

	provider := &langcnt.LangCodesProvider{
		RepoDir: repoDir,
	}

	_, err := provider.LangCodes()
	if err == nil {
		t.Fatal("expected error, got nil")
	}
}

func TestLangCodesProvider_SetLangCodesFilter_OverridesPreviousFilter(t *testing.T) {
	t.Parallel()

	repoDir := t.TempDir()
	contentDir := filepath.Join(repoDir, "content")

	mustMkdir(t, filepath.Join(contentDir, "en"))
	mustMkdir(t, filepath.Join(contentDir, "pl"))
	mustMkdir(t, filepath.Join(contentDir, "fr"))
	mustMkdir(t, filepath.Join(contentDir, "uk"))

	provider := &langcnt.LangCodesProvider{
		RepoDir: repoDir,
	}

	provider.SetLangCodesFilter([]string{"pl"})

	got, err := provider.LangCodes()
	if err != nil {
		t.Fatalf("LangCodes returned error: %v", err)
	}

	if !reflect.DeepEqual(got, []string{"pl"}) {
		t.Fatalf("unexpected lang codes after first filter\nactual:   %v\nexpected: %v", got, []string{"pl"})
	}

	provider.SetLangCodesFilter([]string{"fr", "uk"})

	got, err = provider.LangCodes()
	if err != nil {
		t.Fatalf("LangCodes returned error: %v", err)
	}

	if !reflect.DeepEqual(got, []string{"fr", "uk"}) {
		t.Fatalf("unexpected lang codes after second filter\nactual:   %v\nexpected: %v", got, []string{"fr", "uk"})
	}
}

func mustMkdir(t *testing.T, path string) {
	t.Helper()

	if err := os.MkdirAll(path, 0o755); err != nil {
		t.Fatalf("failed to create directory %s: %v", path, err)
	}
}

func mustWriteFile(t *testing.T, path, content string) {
	t.Helper()

	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatalf("failed to write file %s: %v", path, err)
	}
}
