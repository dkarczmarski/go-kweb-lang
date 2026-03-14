package config

import (
	"errors"
	"fmt"
	"log"
	"os"
	"strings"
)

var (
	ErrBadConfiguration = errors.New("bad configuration")
	ErrInvalidEnvVars   = errors.New("invalid environment variables")
)

func Validate(cfg Config) error {
	if len(cfg.RepoDir) == 0 {
		return fmt.Errorf("param RepoDir is not set: %w", ErrBadConfiguration)
	}

	if len(cfg.CacheDir) == 0 {
		return fmt.Errorf("param CacheDir is not set: %w", ErrBadConfiguration)
	}

	if !cfg.NoWeb && len(cfg.WebHTTPAddr) == 0 {
		return fmt.Errorf("param WebHTTPAddr is not set: %w", ErrBadConfiguration)
	}

	return nil
}

func ReadGitHubTokenFile(cfg *Config, skipFileNotExist, skipEmptyFile bool) error {
	if len(cfg.GitHubTokenFile) == 0 {
		return nil
	}

	tokenData, err := os.ReadFile(cfg.GitHubTokenFile)
	if err != nil {
		if skipFileNotExist && os.IsNotExist(err) {
			log.Printf("github token file does not exist")

			return nil
		}

		return fmt.Errorf("error while reading github token file %v: %w",
			cfg.GitHubTokenFile, err)
	}

	parseGitHubTokenData(cfg, tokenData, skipEmptyFile)

	return nil
}

func parseGitHubTokenData(cfg *Config, tokenData []byte, skipEmptyFile bool) {
	const githubTokenFileLineLimit = 3

	lines := strings.SplitN(string(tokenData), "\n", githubTokenFileLineLimit)

	token := strings.TrimSpace(lines[0])
	if skipEmptyFile && len(token) == 0 {
		log.Printf("github token file is empty")
	} else {
		cfg.GitHubToken = token
	}

	if len(lines) > 1 {
		userAgent := strings.TrimSpace(lines[1])
		if len(userAgent) != 0 {
			cfg.GitHubUserAgent = userAgent
		}
	}
}
