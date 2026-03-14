package config

import (
	"log"
	"strings"
)

func ApplyFlags(
	cfg *Config,
	flagRepoDir *string,
	flagCacheDir *string,
	flagLangCodes *string,
	flagOnce *bool,
	flagInterval *int,
	flagGitHubToken *string,
	flagGitHubTokenFile *string,
	flagSkipGit *bool,
	flagSkipPR *bool,
	flagNoWeb *bool,
	flagWebHTTPAddr *string,
) {
	applyFlagLangCodes(flagLangCodes, &cfg.LangCodes)
	applyFlagString(flagRepoDir, &cfg.RepoDir)
	applyFlagString(flagCacheDir, &cfg.CacheDir)
	applyFlagBoolTrue(flagOnce, &cfg.RunOnce)
	applyFlagIntPositive(flagInterval, &cfg.RunInterval)
	applyFlagString(flagGitHubToken, &cfg.GitHubToken)
	applyFlagString(flagGitHubTokenFile, &cfg.GitHubTokenFile)
	applyFlagBoolTrue(flagSkipGit, &cfg.SkipGitChecking)
	applyFlagBoolTrue(flagSkipPR, &cfg.SkipPRChecking)
	applyFlagBoolTrue(flagNoWeb, &cfg.NoWeb)
	applyFlagString(flagWebHTTPAddr, &cfg.WebHTTPAddr)
}

func Show(cfg Config, withPrint bool) {
	if !withPrint {
		return
	}

	log.Printf("LANG_CODES: %s", strings.Join(cfg.LangCodes, ","))
	log.Printf("REPO_DIR: %s", cfg.RepoDir)
	log.Printf("CACHE_DIR: %s", cfg.CacheDir)
	log.Printf("RUN_ONCE: %v", cfg.RunOnce)
	log.Printf("RUN_INTERVAL: %v", cfg.RunInterval)

	if cfg.GitHubToken != "" {
		log.Printf("GITHUB_TOKEN: (set, len=%d)", len(cfg.GitHubToken))
	} else {
		log.Printf("GITHUB_TOKEN: (empty)")
	}

	log.Printf("GITHUB_TOKEN_FILE: %s", cfg.GitHubTokenFile)
	log.Printf("SKIP_GIT: %v", cfg.SkipGitChecking)
	log.Printf("SKIP_PR: %v", cfg.SkipPRChecking)
	log.Printf("NO_WEB: %v", cfg.NoWeb)
	log.Printf("WEB_HTTP_ADDR: %s", cfg.WebHTTPAddr)
}

func applyFlagString(flag *string, target *string) {
	if flag == nil {
		return
	}

	if v := strings.TrimSpace(*flag); v != "" {
		*target = v
	}
}

func applyFlagLangCodes(flag *string, target *[]string) {
	if flag == nil {
		return
	}

	if strings.TrimSpace(*flag) != "" {
		*target = ParseLangCodes(*flag)
	}
}

// only allows overriding the value to true.
func applyFlagBoolTrue(flag *bool, target *bool) {
	if flag == nil {
		return
	}

	if *flag {
		*target = true
	}
}

// only overrides when the value is positive.
func applyFlagIntPositive(flag *int, target *int) {
	if flag == nil {
		return
	}

	if *flag > 0 {
		*target = *flag
	}
}
