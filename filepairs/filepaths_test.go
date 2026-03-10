package filepairs_test

import (
	"errors"
	"testing"

	"github.com/dkarczmarski/go-kweb-lang/filepairs"
)

func TestNew(t *testing.T) {
	t.Parallel()

	fp := filepairs.New()

	if fp == nil {
		t.Fatal("New() returned nil")
	}
}

func TestFilePaths_CheckPath_MatchEnPath(t *testing.T) {
	t.Parallel()

	fp := filepairs.New()

	info, err := fp.CheckPath("content/en/docs/page.md")
	if err != nil {
		t.Fatalf("CheckPath() error = %v", err)
	}

	if info == nil {
		t.Fatal("CheckPath() returned nil info")
	}

	if info.Path != "content/en/docs/page.md" {
		t.Fatalf("Path = %q, want %q", info.Path, "content/en/docs/page.md")
	}

	if info.PairMatcherName != "content" {
		t.Fatalf("PairMatcherName = %q, want %q", info.PairMatcherName, "content")
	}

	if info.LangCode != "en" {
		t.Fatalf("LangCode = %q, want %q", info.LangCode, "en")
	}

	if !info.IsEnPath() {
		t.Fatal("IsEnPath() = false, want true")
	}
}

func TestFilePaths_CheckPath_MatchLangPath(t *testing.T) {
	t.Parallel()

	fp := filepairs.New()

	info, err := fp.CheckPath("content/pl/docs/page.md")
	if err != nil {
		t.Fatalf("CheckPath() error = %v", err)
	}

	if info == nil {
		t.Fatal("CheckPath() returned nil info")
	}

	if info.Path != "content/pl/docs/page.md" {
		t.Fatalf("Path = %q, want %q", info.Path, "content/pl/docs/page.md")
	}

	if info.PairMatcherName != "content" {
		t.Fatalf("PairMatcherName = %q, want %q", info.PairMatcherName, "content")
	}

	if info.LangCode != "pl" {
		t.Fatalf("LangCode = %q, want %q", info.LangCode, "pl")
	}

	if info.IsEnPath() {
		t.Fatal("IsEnPath() = true, want false")
	}
}

func TestFilePaths_CheckPath_NormalizesPath(t *testing.T) {
	t.Parallel()

	fp := filepairs.New()

	info, err := fp.CheckPath("./content/en/../en/docs/page.md")
	if err != nil {
		t.Fatalf("CheckPath() error = %v", err)
	}

	if info == nil {
		t.Fatal("CheckPath() returned nil info")
	}

	if info.Path != "content/en/docs/page.md" {
		t.Fatalf("Path = %q, want %q", info.Path, "content/en/docs/page.md")
	}

	if info.LangCode != "en" {
		t.Fatalf("LangCode = %q, want %q", info.LangCode, "en")
	}
}

func TestFilePaths_CheckPath_NotFound(t *testing.T) {
	t.Parallel()

	fp := filepairs.New()

	info, err := fp.CheckPath("static/logo.svg")
	if err == nil {
		t.Fatal("CheckPath() error = nil, want error")
	}

	if info != nil {
		t.Fatalf("CheckPath() info = %#v, want nil", info)
	}

	if !errors.Is(err, filepairs.ErrPairMatcherNotFound) {
		t.Fatalf("CheckPath() error = %v, want ErrPairMatcherNotFound", err)
	}
}
