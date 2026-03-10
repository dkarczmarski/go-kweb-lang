package filepairs_test

import (
	"errors"
	"testing"

	"github.com/dkarczmarski/go-kweb-lang/filepairs"
)

//nolint:goconst
func TestContentPairMatcher_Name(t *testing.T) {
	t.Parallel()

	m := filepairs.ContentPairMatcher{}

	if got := m.Name(); got != "content" {
		t.Fatalf("Name() = %q, want %q", got, "content")
	}
}

func TestContentPairMatcher_CheckPath(t *testing.T) {
	t.Parallel()

	m := filepairs.ContentPairMatcher{}

	tests := []struct {
		name         string
		path         string
		wantMatch    bool
		wantLangCode string
		wantErr      error
	}{
		{
			name:         "en path",
			path:         "content/en/docs/page.md",
			wantMatch:    true,
			wantLangCode: "en",
		},
		{
			name:         "lang path",
			path:         "content/pl/docs/page.md",
			wantMatch:    true,
			wantLangCode: "pl",
		},
		{
			name:         "different root",
			path:         "static/logo.svg",
			wantMatch:    false,
			wantLangCode: "",
		},
		{
			name:         "too short path",
			path:         "content/en",
			wantMatch:    false,
			wantLangCode: "",
		},
		{
			name:         "normalized path",
			path:         "./content/en/../en/docs/page.md",
			wantMatch:    true,
			wantLangCode: "en",
		},
		{
			name:         "path with empty language segment after clean becomes invalid match false",
			path:         "content",
			wantMatch:    false,
			wantLangCode: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			gotMatch, gotLangCode, err := m.CheckPath(tt.path)

			if tt.wantErr != nil {
				if !errors.Is(err, tt.wantErr) {
					t.Fatalf("CheckPath() error = %v, want %v", err, tt.wantErr)
				}

				return
			}

			if err != nil {
				t.Fatalf("CheckPath() error = %v", err)
			}

			if gotMatch != tt.wantMatch {
				t.Fatalf("CheckPath() match = %v, want %v", gotMatch, tt.wantMatch)
			}

			if gotLangCode != tt.wantLangCode {
				t.Fatalf("CheckPath() langCode = %q, want %q", gotLangCode, tt.wantLangCode)
			}
		})
	}
}

func TestContentPairMatcher_LangPath(t *testing.T) {
	t.Parallel()

	m := filepairs.ContentPairMatcher{}

	tests := []struct {
		name      string
		path      string
		langCode  string
		wantPath  string
		wantError error
	}{
		{
			name:     "from en path",
			path:     "content/en/docs/page.md",
			langCode: "pl",
			wantPath: "content/pl/docs/page.md",
		},
		{
			name:     "from lang path",
			path:     "content/de/docs/page.md",
			langCode: "fr",
			wantPath: "content/fr/docs/page.md",
		},
		{
			name:      "invalid short path",
			path:      "content/en",
			langCode:  "pl",
			wantError: filepairs.ErrInvalidContentPath,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			got, err := m.LangPath(tt.path, tt.langCode)

			if tt.wantError != nil {
				if err == nil {
					t.Fatal("LangPath() error = nil, want error")
				}

				if !errors.Is(err, tt.wantError) {
					t.Fatalf("LangPath() error = %v, want %v", err, tt.wantError)
				}

				return
			}

			if err != nil {
				t.Fatalf("LangPath() error = %v", err)
			}

			if got != tt.wantPath {
				t.Fatalf("LangPath() = %q, want %q", got, tt.wantPath)
			}
		})
	}
}
