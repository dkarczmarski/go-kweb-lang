//nolint:goconst,paralleltest
package config_test

import (
	"errors"
	"reflect"
	"testing"

	"github.com/dkarczmarski/go-kweb-lang/appinit/config"
)

func TestFromEnv_LoadsValues(t *testing.T) {
	t.Setenv("REPO_DIR", "/tmp/repo")
	t.Setenv("CACHE_DIR", "/tmp/cache")
	t.Setenv("LANG_CODES", "pl, de ,pl,,fr")
	t.Setenv("RUN_ONCE", "true")
	t.Setenv("RUN_INTERVAL", "15")
	t.Setenv("GITHUB_TOKEN", "secret-token")
	t.Setenv("GITHUB_TOKEN_FILE", "/tmp/token.txt")
	t.Setenv("NO_WEB", "true")
	t.Setenv("WEB_HTTP_ADDR", ":9090")

	var cfg config.Config

	err := config.FromEnv(&cfg)
	if err != nil {
		t.Fatalf("FromEnv returned error: %v", err)
	}

	if cfg.RepoDir != "/tmp/repo" {
		t.Fatalf("unexpected RepoDir: %q", cfg.RepoDir)
	}

	if cfg.CacheDir != "/tmp/cache" {
		t.Fatalf("unexpected CacheDir: %q", cfg.CacheDir)
	}

	wantLangs := []string{"pl", "de", "fr"}
	if !reflect.DeepEqual(cfg.LangCodes, wantLangs) {
		t.Fatalf("unexpected LangCodes: got %#v, want %#v", cfg.LangCodes, wantLangs)
	}

	if !cfg.RunOnce {
		t.Fatalf("expected RunOnce=true")
	}

	if cfg.RunInterval != 15 {
		t.Fatalf("unexpected RunInterval: %d", cfg.RunInterval)
	}

	if cfg.GitHubToken != "secret-token" {
		t.Fatalf("unexpected GitHubToken: %q", cfg.GitHubToken)
	}

	if cfg.GitHubTokenFile != "/tmp/token.txt" {
		t.Fatalf("unexpected GitHubTokenFile: %q", cfg.GitHubTokenFile)
	}

	if !cfg.NoWeb {
		t.Fatalf("expected NoWeb=true")
	}

	if cfg.WebHTTPAddr != ":9090" {
		t.Fatalf("unexpected WebHTTPAddr: %q", cfg.WebHTTPAddr)
	}
}

func TestFromEnv_InvalidBool(t *testing.T) {
	t.Setenv("RUN_ONCE", "not-bool")

	var cfg config.Config
	err := config.FromEnv(&cfg)

	if err == nil {
		t.Fatal("expected error, got nil")
	}

	if !errors.Is(err, config.ErrInvalidEnvVars) {
		t.Fatalf("expected ErrInvalidEnvVars, got %v", err)
	}
}

func TestFromEnv_InvalidInt(t *testing.T) {
	t.Setenv("RUN_INTERVAL", "-5")

	var cfg config.Config
	err := config.FromEnv(&cfg)

	if err == nil {
		t.Fatal("expected error, got nil")
	}

	if !errors.Is(err, config.ErrInvalidEnvVars) {
		t.Fatalf("expected ErrInvalidEnvVars, got %v", err)
	}
}

func TestFromEnv_DoesNotChangeValuesWhenVarsMissing(t *testing.T) {
	cfg := config.Config{
		RepoDir:     "/existing/repo",
		CacheDir:    "/existing/cache",
		RunOnce:     true,
		RunInterval: 7,
		NoWeb:       true,
		WebHTTPAddr: ":1234",
	}

	err := config.FromEnv(&cfg)
	if err != nil {
		t.Fatalf("FromEnv returned error: %v", err)
	}

	want := config.Config{
		RepoDir:     "/existing/repo",
		CacheDir:    "/existing/cache",
		RunOnce:     true,
		RunInterval: 7,
		NoWeb:       true,
		WebHTTPAddr: ":1234",
	}

	if !reflect.DeepEqual(cfg, want) {
		t.Fatalf("config changed unexpectedly: got %#v, want %#v", cfg, want)
	}
}
