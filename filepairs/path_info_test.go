package filepairs_test

import (
	"errors"
	"testing"

	"github.com/dkarczmarski/go-kweb-lang/filepairs"
)

func TestPathInfo_LangPath_FromEnPath(t *testing.T) {
	t.Parallel()

	fp := filepairs.New()

	info, err := fp.CheckPath("content/en/docs/page.md")
	if err != nil {
		t.Fatalf("CheckPath() error = %v", err)
	}

	if info.LangCode != "en" {
		t.Fatalf("LangCode = %q, want %q", info.LangCode, "en")
	}

	if !info.IsEnPath() {
		t.Fatal("IsEnPath() = false, want true")
	}

	got, err := info.LangPath("pl")
	if err != nil {
		t.Fatalf("LangPath() error = %v", err)
	}

	want := "content/pl/docs/page.md"
	if got != want {
		t.Fatalf("LangPath() = %q, want %q", got, want)
	}
}

func TestPathInfo_LangPath_FromNonEnPath(t *testing.T) {
	t.Parallel()

	fp := filepairs.New()

	info, err := fp.CheckPath("content/pl/docs/page.md")
	if err != nil {
		t.Fatalf("CheckPath() error = %v", err)
	}

	if info.LangCode != "pl" {
		t.Fatalf("LangCode = %q, want %q", info.LangCode, "pl")
	}

	if info.IsEnPath() {
		t.Fatal("IsEnPath() = true, want false")
	}

	_, err = info.LangPath("de")
	if err == nil {
		t.Fatal("LangPath() error = nil, want error")
	}

	if !errors.Is(err, filepairs.ErrLangPathRequiresEnPath) {
		t.Fatalf("LangPath() error = %v, want ErrLangPathRequiresEnPath", err)
	}
}

func TestPathInfo_LangPath_NilPathInfo(t *testing.T) {
	t.Parallel()

	var info *filepairs.PathInfo

	_, err := info.LangPath("pl")
	if err == nil {
		t.Fatal("LangPath() error = nil, want error")
	}

	if !errors.Is(err, filepairs.ErrInvalidInput) {
		t.Fatalf("LangPath() error = %v, want ErrInvalidInput", err)
	}
}
