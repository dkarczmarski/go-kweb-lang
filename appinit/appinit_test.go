package appinit_test

import (
	"errors"
	"os"
	"reflect"
	"testing"

	"github.com/dkarczmarski/go-kweb-lang/appinit"
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

			assertField(t, "RepoDir", cfg.RepoDir, tt.want.RepoDir)
			assertField(t, "CacheDir", cfg.CacheDir, tt.want.CacheDir)
			assertSlice(t, "LangCodes", cfg.LangCodes, tt.want.LangCodes)
			assertField(t, "RunOnce", cfg.RunOnce, tt.want.RunOnce)
			assertField(t, "RunInterval", cfg.RunInterval, tt.want.RunInterval)
			assertField(t, "GitHubToken", cfg.GitHubToken, tt.want.GitHubToken)
			assertField(t, "GitHubTokenFile", cfg.GitHubTokenFile, tt.want.GitHubTokenFile)
			assertField(t, "NoWeb", cfg.NoWeb, tt.want.NoWeb)
			assertField(t, "WebHTTPAddr", cfg.WebHTTPAddr, tt.want.WebHTTPAddr)
		})
	}
}

func TestParseFlagParams(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string

		// flags
		flagRepoDir         string
		flagCacheDir        string
		flagLangCodes       string
		flagOnce            bool
		flagInterval        int
		flagGitHubToken     string
		flagGitHubTokenFile string
		flagSkipGit         bool
		flagSkipPR          bool
		flagNoWeb           bool
		flagWebHTTPAddr     string

		// initial config (to test "do not override" behavior)
		initial appinit.Config

		// expected config after applying ParseFlagParams
		want appinit.Config
	}{
		{
			name: "does not override when flags are empty/false/zero",
			initial: appinit.Config{
				RepoDir:         "repo0",
				CacheDir:        "cache0",
				LangCodes:       []string{"en"},
				RunOnce:         true,
				RunInterval:     10,
				GitHubToken:     "tok0",
				GitHubTokenFile: "file0",
				SkipGitChecking: true,
				SkipPRChecking:  true,
				NoWeb:           true,
				WebHTTPAddr:     ":9999",
			},
			// all zero values for flags => should keep initial
			want: appinit.Config{
				RepoDir:         "repo0",
				CacheDir:        "cache0",
				LangCodes:       []string{"en"},
				RunOnce:         true,
				RunInterval:     10,
				GitHubToken:     "tok0",
				GitHubTokenFile: "file0",
				SkipGitChecking: true,
				SkipPRChecking:  true,
				NoWeb:           true,
				WebHTTPAddr:     ":9999",
			},
		},
		{
			name:         "overrides repo and cache dirs when provided",
			flagRepoDir:  "repo1",
			flagCacheDir: "cache1",
			initial: appinit.Config{
				RepoDir:  "repo0",
				CacheDir: "cache0",
			},
			want: appinit.Config{
				RepoDir:  "repo1",
				CacheDir: "cache1",
			},
		},
		{
			name:          "parses LANG_CODES when provided",
			flagLangCodes: "pl, en, de ",
			initial: appinit.Config{
				LangCodes: []string{"xx"},
			},
			want: appinit.Config{
				LangCodes: []string{"pl", "en", "de"},
			},
		},
		{
			name:     "RUN_ONCE can only be overridden to true",
			flagOnce: true,
			initial: appinit.Config{
				RunOnce: false,
			},
			want: appinit.Config{
				RunOnce: true,
			},
		},
		{
			name:     "RUN_ONCE false flag does not override true initial value",
			flagOnce: false,
			initial: appinit.Config{
				RunOnce: true,
			},
			want: appinit.Config{
				RunOnce: true,
			},
		},
		{
			name:         "RUN_INTERVAL overrides only when > 0",
			flagInterval: 15,
			initial: appinit.Config{
				RunInterval: 3,
			},
			want: appinit.Config{
				RunInterval: 15,
			},
		},
		{
			name:         "RUN_INTERVAL does not override when 0",
			flagInterval: 0,
			initial: appinit.Config{
				RunInterval: 7,
			},
			want: appinit.Config{
				RunInterval: 7,
			},
		},
		{
			name:                "overrides github token and token file when provided",
			flagGitHubToken:     "tok1",
			flagGitHubTokenFile: "file1",
			initial: appinit.Config{
				GitHubToken:     "tok0",
				GitHubTokenFile: "file0",
			},
			want: appinit.Config{
				GitHubToken:     "tok1",
				GitHubTokenFile: "file1",
			},
		},
		{
			name:        "SKIP_GIT sets SkipGitChecking",
			flagSkipGit: true,
			initial: appinit.Config{
				SkipGitChecking: false,
				SkipPRChecking:  false,
			},
			want: appinit.Config{
				SkipGitChecking: true,
				SkipPRChecking:  false,
			},
		},
		{
			name:       "SKIP_PR sets SkipPRChecking",
			flagSkipPR: true,
			initial: appinit.Config{
				SkipGitChecking: false,
				SkipPRChecking:  false,
			},
			want: appinit.Config{
				SkipGitChecking: false,
				SkipPRChecking:  true,
			},
		},
		{
			name:      "NO_WEB sets NoWeb",
			flagNoWeb: true,
			initial: appinit.Config{
				NoWeb: false,
			},
			want: appinit.Config{
				NoWeb: true,
			},
		},
		{
			name:            "WEB_HTTP_ADDR overrides when provided",
			flagWebHTTPAddr: ":8081",
			initial: appinit.Config{
				WebHTTPAddr: ":8080",
			},
			want: appinit.Config{
				WebHTTPAddr: ":8081",
			},
		},
		{
			name:            "mixed overrides apply together",
			flagRepoDir:     "repoX",
			flagOnce:        true,
			flagInterval:    20,
			flagSkipGit:     true,
			flagSkipPR:      true,
			flagNoWeb:       true,
			flagWebHTTPAddr: ":1234",
			flagLangCodes:   "en,pl",
			initial: appinit.Config{
				RepoDir:         "repo0",
				CacheDir:        "cache0",
				LangCodes:       []string{"xx"},
				RunOnce:         false,
				RunInterval:     5,
				SkipGitChecking: false,
				SkipPRChecking:  false,
				NoWeb:           false,
				WebHTTPAddr:     ":8080",
			},
			want: appinit.Config{
				RepoDir:         "repoX",
				CacheDir:        "cache0", // unchanged
				LangCodes:       []string{"en", "pl"},
				RunOnce:         true,
				RunInterval:     20,
				SkipGitChecking: true,
				SkipPRChecking:  true,
				NoWeb:           true,
				WebHTTPAddr:     ":1234",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			cfg := tt.initial // copy

			// create pointer flags expected by ParseFlagParams
			repoDir := tt.flagRepoDir
			cacheDir := tt.flagCacheDir
			langCodes := tt.flagLangCodes
			once := tt.flagOnce
			interval := tt.flagInterval
			ghToken := tt.flagGitHubToken
			ghTokenFile := tt.flagGitHubTokenFile
			skipGit := tt.flagSkipGit
			skipPR := tt.flagSkipPR
			noWeb := tt.flagNoWeb
			webHTTPAddr := tt.flagWebHTTPAddr

			opt := appinit.ParseFlagParams(
				&repoDir,
				&cacheDir,
				&langCodes,
				&once,
				&interval,
				&ghToken,
				&ghTokenFile,
				&skipGit,
				&skipPR,
				&noWeb,
				&webHTTPAddr,
			)

			if err := opt(&cfg); err != nil {
				t.Fatalf("ParseFlagParams returned error: %v", err)
			}

			assertField(t, "RepoDir", cfg.RepoDir, tt.want.RepoDir)
			assertField(t, "CacheDir", cfg.CacheDir, tt.want.CacheDir)
			assertSlice(t, "LangCodes", cfg.LangCodes, tt.want.LangCodes)
			assertField(t, "RunOnce", cfg.RunOnce, tt.want.RunOnce)
			assertField(t, "RunInterval", cfg.RunInterval, tt.want.RunInterval)
			assertField(t, "GitHubToken", cfg.GitHubToken, tt.want.GitHubToken)
			assertField(t, "GitHubTokenFile", cfg.GitHubTokenFile, tt.want.GitHubTokenFile)
			assertField(t, "SkipGitChecking", cfg.SkipGitChecking, tt.want.SkipGitChecking)
			assertField(t, "SkipPRChecking", cfg.SkipPRChecking, tt.want.SkipPRChecking)
			assertField(t, "NoWeb", cfg.NoWeb, tt.want.NoWeb)
			assertField(t, "WebHTTPAddr", cfg.WebHTTPAddr, tt.want.WebHTTPAddr)
		})
	}
}

func assertField[T comparable](t *testing.T, name string, got, want T) {
	t.Helper()

	if got != want {
		t.Fatalf("%s: got %#v, want %#v", name, got, want)
	}
}

func assertSlice(t *testing.T, name string, got, want []string) {
	t.Helper()

	if !reflect.DeepEqual(got, want) {
		t.Fatalf("%s: got %#v, want %#v", name, got, want)
	}
}
