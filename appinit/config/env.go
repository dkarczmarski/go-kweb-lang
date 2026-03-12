package config

import (
	"fmt"
	"os"
	"strconv"
	"strings"
)

func FromEnv(cfg *Config) error {
	var errs []string

	if v, ok := env("REPO_DIR"); ok {
		cfg.RepoDir = v
	}

	if v, ok := env("CACHE_DIR"); ok {
		cfg.CacheDir = v
	}

	if v, ok := env("LANG_CODES"); ok {
		cfg.LangCodes = ParseLangCodes(v)
	}

	errs = parseEnvBool("RUN_ONCE", &cfg.RunOnce, errs)
	errs = parseEnvInt("RUN_INTERVAL", &cfg.RunInterval, errs)

	if v, ok := env("GITHUB_TOKEN"); ok {
		cfg.GitHubToken = v
	}

	if v, ok := env("GITHUB_TOKEN_FILE"); ok {
		cfg.GitHubTokenFile = v
	}

	errs = parseEnvBool("NO_WEB", &cfg.NoWeb, errs)

	if v, ok := env("WEB_HTTP_ADDR"); ok {
		cfg.WebHTTPAddr = v
	}

	if len(errs) > 0 {
		return fmt.Errorf(
			"%w:\n - %s",
			ErrInvalidEnvVars,
			strings.Join(errs, "\n - "),
		)
	}

	return nil
}

func env(key string) (string, bool) {
	return os.LookupEnv(key)
}

func parseEnvBool(key string, target *bool, errs []string) []string {
	value, ok := os.LookupEnv(key)
	if !ok {
		return errs
	}

	parsed, err := strconv.ParseBool(value)
	if err != nil {
		return append(errs, fmt.Sprintf("%s must be boolean, got %q", key, value))
	}

	*target = parsed

	return errs
}

func parseEnvInt(key string, target *int, errs []string) []string {
	value, ok := os.LookupEnv(key)
	if !ok {
		return errs
	}

	parsed, err := strconv.Atoi(value)
	if err != nil || parsed < 0 {
		return append(errs, fmt.Sprintf("%s must be non-negative int, got %q", key, value))
	}

	*target = parsed

	return errs
}

func ParseLangCodes(langCodesStr string) []string {
	langCodesStr = strings.TrimSpace(langCodesStr)
	if langCodesStr == "" {
		return nil
	}

	seen := map[string]struct{}{}
	subs := strings.Split(langCodesStr, ",")
	allowedLangCodes := make([]string, 0, len(subs))

	for _, sub := range subs {
		sub = strings.TrimSpace(sub)
		if sub == "" {
			continue
		}

		if _, ok := seen[sub]; ok {
			continue
		}

		seen[sub] = struct{}{}

		allowedLangCodes = append(allowedLangCodes, sub)
	}

	if len(allowedLangCodes) == 0 {
		return nil
	}

	return allowedLangCodes
}
