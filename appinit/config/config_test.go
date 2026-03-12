package config_test

import (
	"reflect"
	"testing"

	"github.com/dkarczmarski/go-kweb-lang/appinit/config"
)

func TestDefault(t *testing.T) {
	t.Parallel()

	cfg := config.Default()

	if cfg.RepoDir != "./.appdata/kubernetes-website" {
		t.Fatalf("unexpected RepoDir: %q", cfg.RepoDir)
	}

	if cfg.CacheDir != "./.appdata/cache" {
		t.Fatalf("unexpected CacheDir: %q", cfg.CacheDir)
	}

	if cfg.GitHubTokenFile != ".github-token.txt" {
		t.Fatalf("unexpected GitHubTokenFile: %q", cfg.GitHubTokenFile)
	}

	if cfg.WebHTTPAddr != ":8080" {
		t.Fatalf("unexpected WebHTTPAddr: %q", cfg.WebHTTPAddr)
	}
}

func TestParseLangCodes_Empty(t *testing.T) {
	t.Parallel()

	got := config.ParseLangCodes("   ")
	if got != nil {
		t.Fatalf("expected nil, got %#v", got)
	}
}

func TestParseLangCodes_DeduplicatesAndTrims(t *testing.T) {
	t.Parallel()

	got := config.ParseLangCodes("pl, de ,pl,,fr")
	want := []string{"pl", "de", "fr"}

	if !reflect.DeepEqual(got, want) {
		t.Fatalf("unexpected lang codes: got %#v, want %#v", got, want)
	}
}

func TestParseLangCodes_AllEmpty(t *testing.T) {
	t.Parallel()

	got := config.ParseLangCodes(" , , ")
	if got != nil {
		t.Fatalf("expected nil, got %#v", got)
	}
}
