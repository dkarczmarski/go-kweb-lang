package config_test

import (
	"errors"
	"os"
	"path/filepath"
	"testing"

	"github.com/dkarczmarski/go-kweb-lang/appinit/config"
)

func TestValidate_OK(t *testing.T) {
	t.Parallel()

	cfg := config.Config{
		RepoDir:     "/tmp/repo",
		CacheDir:    "/tmp/cache",
		WebHTTPAddr: ":8080",
	}

	err := config.Validate(cfg)
	if err != nil {
		t.Fatalf("Validate returned error: %v", err)
	}
}

func TestValidate_MissingRepoDir(t *testing.T) {
	t.Parallel()

	cfg := config.Config{
		CacheDir:    "/tmp/cache",
		WebHTTPAddr: ":8080",
	}

	err := config.Validate(cfg)

	if err == nil {
		t.Fatal("expected error")
	}

	if !errors.Is(err, config.ErrBadConfiguration) {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestValidate_NoWebAllowsEmptyAddr(t *testing.T) {
	t.Parallel()

	cfg := config.Config{
		RepoDir:  "/tmp/repo",
		CacheDir: "/tmp/cache",
		NoWeb:    true,
	}

	err := config.Validate(cfg)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestReadGitHubTokenFile_TwoLines(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	file := filepath.Join(dir, "token.txt")

	err := os.WriteFile(file, []byte("token\nagent\n"), 0o600)
	if err != nil {
		t.Fatal(err)
	}

	cfg := config.Config{
		GitHubTokenFile: file,
	}

	err = config.ReadGitHubTokenFile(&cfg, false, false)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if cfg.GitHubToken != "token" {
		t.Fatalf("unexpected token: %q", cfg.GitHubToken)
	}

	if cfg.GitHubUserAgent != "agent" {
		t.Fatalf("unexpected user agent: %q", cfg.GitHubUserAgent)
	}
}

func TestReadGitHubTokenFile_EmptyFileSkipEmpty(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	file := filepath.Join(dir, "token.txt")

	err := os.WriteFile(file, []byte(""), 0o600)
	if err != nil {
		t.Fatal(err)
	}

	cfg := config.Config{
		GitHubTokenFile: file,
		GitHubToken:     "existing",
	}

	err = config.ReadGitHubTokenFile(&cfg, false, true)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if cfg.GitHubToken != "existing" {
		t.Fatalf("token was overridden: %q", cfg.GitHubToken)
	}
}
