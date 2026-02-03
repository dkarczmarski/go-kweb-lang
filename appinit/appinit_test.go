package appinit_test

import (
	"errors"
	"os"
	"reflect"
	"testing"

	"go-kweb-lang/appinit"
)

func TestParseEnvParams(t *testing.T) {
	type tc struct {
		name    string
		env     map[string]string
		initial appinit.Config
		want    appinit.Config
		wantErr bool
	}

	tests := []tc{
		{
			name: "no env vars -> config unchanged",
			initial: appinit.Config{
				RepoDir:         "repo0",
				CacheDir:        "cache0",
				LangCodes:       []string{"en"},
				RunOnce:         true,
				RunInterval:     10,
				GitHubToken:     "tok0",
				GitHubTokenFile: "file0",
				NoWeb:           true,
				WebHTTPAddr:     ":9999",
			},
			want: appinit.Config{
				RepoDir:         "repo0",
				CacheDir:        "cache0",
				LangCodes:       []string{"en"},
				RunOnce:         true,
				RunInterval:     10,
				GitHubToken:     "tok0",
				GitHubTokenFile: "file0",
				NoWeb:           true,
				WebHTTPAddr:     ":9999",
			},
		},
		{
			name: "valid env vars override fields",
			env: map[string]string{
				"REPO_DIR":          "repo1",
				"CACHE_DIR":         "cache1",
				"LANG_CODES":        "pl, en, de",
				"RUN_ONCE":          "true",
				"RUN_INTERVAL":      "30",
				"GITHUB_TOKEN":      "tok1",
				"GITHUB_TOKEN_FILE": "file1",
				"NO_WEB":            "true",
				"WEB_HTTP_ADDR":     ":8081",
			},
			initial: appinit.Config{
				RepoDir:         "repo0",
				CacheDir:        "cache0",
				LangCodes:       []string{"xx"},
				RunOnce:         false,
				RunInterval:     5,
				GitHubToken:     "tok0",
				GitHubTokenFile: "file0",
				NoWeb:           false,
				WebHTTPAddr:     ":8080",
			},
			want: appinit.Config{
				RepoDir:         "repo1",
				CacheDir:        "cache1",
				LangCodes:       []string{"pl", "en", "de"},
				RunOnce:         true,
				RunInterval:     30,
				GitHubToken:     "tok1",
				GitHubTokenFile: "file1",
				NoWeb:           true,
				WebHTTPAddr:     ":8081",
			},
		},
		{
			name: "RUN_INTERVAL negative -> error and RunInterval not overridden",
			env: map[string]string{
				"RUN_INTERVAL": "-1",
			},
			initial: appinit.Config{
				RunInterval: 7,
			},
			want: appinit.Config{
				RunInterval: 7,
			},
			wantErr: true,
		},
		{
			name: "invalid booleans/ints -> error, invalid fields not overridden, valid fields still applied",
			env: map[string]string{
				"REPO_DIR":     "repoX",
				"RUN_ONCE":     "not-bool",
				"RUN_INTERVAL": "abc",
				"NO_WEB":       "maybe",
			},
			initial: appinit.Config{
				RepoDir:     "repo0",
				RunOnce:     true, // should remain unchanged due to invalid env
				RunInterval: 10,   // should remain unchanged due to invalid env
				NoWeb:       true, // should remain unchanged due to invalid env
			},
			want: appinit.Config{
				RepoDir:     "repoX", // valid env applied
				RunOnce:     true,    // unchanged (invalid)
				RunInterval: 10,      // unchanged (invalid)
				NoWeb:       true,    // unchanged (invalid)
			},
			wantErr: true,
		},
		{
			name: "LANG_CODES empty/whitespace -> parseLangCodes returns nil (overrides to nil)",
			env: map[string]string{
				"LANG_CODES": "   ",
			},
			initial: appinit.Config{
				LangCodes: []string{"en"},
			},
			want: appinit.Config{
				LangCodes: nil,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// ensure clean env for keys we may use
			for _, k := range []string{
				"REPO_DIR", "CACHE_DIR", "LANG_CODES", "RUN_ONCE", "RUN_INTERVAL",
				"GITHUB_TOKEN", "GITHUB_TOKEN_FILE", "NO_WEB", "WEB_HTTP_ADDR",
			} {
				os.Unsetenv(k)
			}

			for k, v := range tt.env {
				t.Setenv(k, v)
			}

			cfg := tt.initial

			opt := appinit.ParseEnvParams()
			err := opt(&cfg)

			if tt.wantErr {
				if err == nil {
					t.Fatalf("expected error, got nil")
				}

				if !errors.Is(err, appinit.ErrInvalidEnvVars) {
					t.Fatalf("expected errors.Is(err, ErrInvalidEnvVars)=true, got false; err=%v", err)
				}
			} else {
				if err != nil {
					t.Fatalf("unexpected error: %v", err)
				}
			}

			assertEqual(t, "RepoDir", cfg.RepoDir, tt.want.RepoDir)
			assertEqual(t, "CacheDir", cfg.CacheDir, tt.want.CacheDir)
			assertSliceEqual(t, "LangCodes", cfg.LangCodes, tt.want.LangCodes)
			assertEqual(t, "RunOnce", cfg.RunOnce, tt.want.RunOnce)
			assertEqual(t, "RunInterval", cfg.RunInterval, tt.want.RunInterval)
			assertEqual(t, "GitHubToken", cfg.GitHubToken, tt.want.GitHubToken)
			assertEqual(t, "GitHubTokenFile", cfg.GitHubTokenFile, tt.want.GitHubTokenFile)
			assertEqual(t, "NoWeb", cfg.NoWeb, tt.want.NoWeb)
			assertEqual(t, "WebHTTPAddr", cfg.WebHTTPAddr, tt.want.WebHTTPAddr)
		})
	}
}

func assertEqual[T comparable](t *testing.T, field string, got, want T) {
	t.Helper()

	if got != want {
		t.Fatalf("%s: got %#v, want %#v", field, got, want)
	}
}

func assertSliceEqual(t *testing.T, field string, got, want []string) {
	t.Helper()

	if !reflect.DeepEqual(got, want) {
		t.Fatalf("%s: got %#v, want %#v", field, got, want)
	}
}
