//nolint:dupl
package filepairs_test

import (
	"errors"
	"testing"

	"github.com/dkarczmarski/go-kweb-lang/filepairs"
)

func TestI18NPairMatcher_Name(t *testing.T) {
	t.Parallel()

	matcher := filepairs.I18NPairMatcher{}

	if got := matcher.Name(); got != "i18n" {
		t.Fatalf("Name() = %q, want %q", got, "i18n")
	}
}

func TestI18NPairMatcher_CheckPath(t *testing.T) {
	t.Parallel()

	matcher := filepairs.I18NPairMatcher{}

	tests := []struct {
		name         string
		path         string
		wantMatch    bool
		wantLangCode string
		wantErr      error
	}{
		{
			name:         "en path",
			path:         "i18n/en/en.toml",
			wantMatch:    true,
			wantLangCode: "en",
		},
		{
			name:         "lang path",
			path:         "i18n/pl/pl.toml",
			wantMatch:    true,
			wantLangCode: "pl",
		},
		{
			name:         "different root",
			path:         "content/pl/pl.toml",
			wantMatch:    false,
			wantLangCode: "",
		},
		{
			name:         "too short path",
			path:         "i18n/pl",
			wantMatch:    false,
			wantLangCode: "",
		},
		{
			name:         "too long path",
			path:         "i18n/pl/messages/pl.toml",
			wantMatch:    false,
			wantLangCode: "",
		},
		{
			name:         "wrong file name",
			path:         "i18n/pl/en.toml",
			wantMatch:    false,
			wantLangCode: "",
		},
		{
			name:         "normalized path",
			path:         "./i18n/en/../en/en.toml",
			wantMatch:    true,
			wantLangCode: "en",
		},
	}

	for _, testCase := range tests {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()

			gotMatch, gotLangCode, err := matcher.CheckPath(testCase.path)

			if testCase.wantErr != nil {
				if !errors.Is(err, testCase.wantErr) {
					t.Fatalf("CheckPath() error = %v, want %v", err, testCase.wantErr)
				}

				return
			}

			if err != nil {
				t.Fatalf("CheckPath() error = %v", err)
			}

			if gotMatch != testCase.wantMatch {
				t.Fatalf("CheckPath() match = %v, want %v", gotMatch, testCase.wantMatch)
			}

			if gotLangCode != testCase.wantLangCode {
				t.Fatalf("CheckPath() langCode = %q, want %q", gotLangCode, testCase.wantLangCode)
			}
		})
	}
}

func TestI18NPairMatcher_LangPath(t *testing.T) {
	t.Parallel()

	matcher := filepairs.I18NPairMatcher{}

	tests := []struct {
		name      string
		path      string
		langCode  string
		wantPath  string
		wantError error
	}{
		{
			name:     "from en path",
			path:     "i18n/en/en.toml",
			langCode: "pl",
			wantPath: "i18n/pl/pl.toml",
		},
		{
			name:     "from lang path",
			path:     "i18n/de/de.toml",
			langCode: "fr",
			wantPath: "i18n/fr/fr.toml",
		},
		{
			name:      "invalid path",
			path:      "i18n/de/messages.toml",
			langCode:  "fr",
			wantError: filepairs.ErrInvalidI18NPath,
		},
	}

	for _, testCase := range tests {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()

			got, err := matcher.LangPath(testCase.path, testCase.langCode)

			if testCase.wantError != nil {
				if err == nil {
					t.Fatal("LangPath() error = nil, want error")
				}

				if !errors.Is(err, testCase.wantError) {
					t.Fatalf("LangPath() error = %v, want %v", err, testCase.wantError)
				}

				return
			}

			if err != nil {
				t.Fatalf("LangPath() error = %v", err)
			}

			if got != testCase.wantPath {
				t.Fatalf("LangPath() = %q, want %q", got, testCase.wantPath)
			}
		})
	}
}
