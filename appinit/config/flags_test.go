package config_test

import (
	"reflect"
	"testing"

	"github.com/dkarczmarski/go-kweb-lang/appinit/config"
)

func TestApplyFlags_AppliesValues(t *testing.T) {
	t.Parallel()

	cfg := config.Config{
		RepoDir:     "old-repo",
		CacheDir:    "old-cache",
		LangCodes:   []string{"en"},
		RunInterval: 5,
		WebHTTPAddr: ":8080",
	}

	repoDir := "new-repo"
	cacheDir := "new-cache"
	langCodes := "pl, de ,pl,,fr"
	runOnce := true
	runInterval := 30
	noWeb := true
	webHTTPAddr := ":9090"

	config.ApplyFlags(
		&cfg,
		&repoDir,
		&cacheDir,
		&langCodes,
		&runOnce,
		&runInterval,
		nil,
		nil,
		nil,
		nil,
		&noWeb,
		&webHTTPAddr,
	)

	if cfg.RepoDir != "new-repo" {
		t.Fatalf("unexpected RepoDir: %q", cfg.RepoDir)
	}

	if cfg.CacheDir != "new-cache" {
		t.Fatalf("unexpected CacheDir: %q", cfg.CacheDir)
	}

	wantLangs := []string{"pl", "de", "fr"}
	if !reflect.DeepEqual(cfg.LangCodes, wantLangs) {
		t.Fatalf("unexpected LangCodes: got %#v", cfg.LangCodes)
	}

	if !cfg.RunOnce {
		t.Fatalf("expected RunOnce=true")
	}

	if cfg.RunInterval != 30 {
		t.Fatalf("unexpected RunInterval: %d", cfg.RunInterval)
	}

	if !cfg.NoWeb {
		t.Fatalf("expected NoWeb=true")
	}

	if cfg.WebHTTPAddr != ":9090" {
		t.Fatalf("unexpected WebHTTPAddr: %q", cfg.WebHTTPAddr)
	}
}

func TestApplyFlags_NilPointersIgnored(t *testing.T) {
	t.Parallel()

	cfg := config.Config{
		RepoDir:     "repo",
		CacheDir:    "cache",
		RunInterval: 10,
		WebHTTPAddr: ":8080",
	}

	config.ApplyFlags(
		&cfg,
		nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil,
	)

	want := config.Config{
		RepoDir:     "repo",
		CacheDir:    "cache",
		RunInterval: 10,
		WebHTTPAddr: ":8080",
	}

	if !reflect.DeepEqual(cfg, want) {
		t.Fatalf("config changed unexpectedly: got %#v, want %#v", cfg, want)
	}
}
