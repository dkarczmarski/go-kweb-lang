package appinit

import (
	"errors"
	"fmt"
	"go-kweb-lang/git"
	"go-kweb-lang/gitcache"
	"go-kweb-lang/github"
	"go-kweb-lang/langcnt"
	"go-kweb-lang/tasks"
	"go-kweb-lang/web"
	"log"
	"os"
	"strings"
)

type Config struct {
	RepoDirPath             string
	CacheDirPath            string
	AllowedLangs            []string
	Content                 *langcnt.Content
	GitRepo                 git.Repo
	TemplateData            *web.TemplateData
	GitRepoCache            *gitcache.GitRepoCache
	GitHub                  *github.GitHub
	RefreshRepoTask         *tasks.RefreshRepoTask
	RefreshTemplateDataTask *tasks.RefreshTemplateDataTask
	RepoMonitor             *github.RepoMonitor
	Server                  *web.Server
}

var ErrBadConfiguration = errors.New("bad configuration")

func Init(opts ...func(config *Config) error) (*Config, error) {
	var config Config

	for _, opt := range opts {
		if err := opt(&config); err != nil {
			return &config, err
		}
	}

	return &config, nil
}

func getEnvOrDefault(key, defaultValue string) string {
	value, ok := os.LookupEnv(key)
	if !ok {
		return defaultValue
	}
	return value
}

func GetEnv(withPrint bool) func(*Config) error {
	return func(config *Config) error {
		repoDirPath := getEnvOrDefault("REPO_DIR", "../kubernetes-website")
		cacheDirPath := getEnvOrDefault("CACHE_DIR", "./cache")
		allowedLangs := getEnvOrDefault("ALLOWED_LANGS", "")

		if withPrint {
			log.Printf("REPO_DIR: %s", repoDirPath)
			log.Printf("CACHE_DIR: %s", cacheDirPath)
			log.Printf("ALLOWED_LANGS: %s", allowedLangs)
		}

		config.RepoDirPath = repoDirPath
		config.CacheDirPath = cacheDirPath
		config.AllowedLangs = parseAllowedLangs(allowedLangs)

		return nil
	}
}

func parseAllowedLangs(s string) []string {
	if len(strings.TrimSpace(s)) == 0 {
		return nil
	}

	subs := strings.Split(s, ",")
	allowedLangs := make([]string, 0, len(subs))

	for _, sub := range subs {
		allowedLangs = append(allowedLangs, strings.TrimSpace(sub))
	}

	return allowedLangs
}

func NewContent() func(config *Config) error {
	return func(config *Config) error {
		repoDirPath := config.RepoDirPath
		if len(repoDirPath) == 0 {
			return fmt.Errorf("param RepoDirPath is not set: %w", ErrBadConfiguration)
		}

		content := &langcnt.Content{RepoDir: repoDirPath}
		content.SetAllowedLang(config.AllowedLangs)

		config.Content = content

		return nil
	}
}

func NewRepo() func(*Config) error {
	return func(config *Config) error {
		repoDirPath := config.RepoDirPath
		if len(repoDirPath) == 0 {
			return fmt.Errorf("param RepoDirPath is not set: %w", ErrBadConfiguration)
		}

		config.GitRepo = git.NewRepo(repoDirPath)

		return nil
	}
}

func NewTemplateData() func(*Config) error {
	return func(config *Config) error {
		config.TemplateData = web.NewTemplateData()

		return nil
	}
}

func NewRepoCache() func(*Config) error {
	return func(config *Config) error {
		gitRepo := config.GitRepo
		if gitRepo == nil {
			return fmt.Errorf("param GitRepo is not set: %w", ErrBadConfiguration)
		}

		cacheDirPath := config.CacheDirPath
		if len(cacheDirPath) == 0 {
			return fmt.Errorf("param CacheDirPath is not set: %w", ErrBadConfiguration)
		}

		config.GitRepoCache = gitcache.New(gitRepo, cacheDirPath)

		return nil
	}
}

func NewGitHub() func(*Config) error {
	return func(config *Config) error {
		config.GitHub = github.New()

		return nil
	}
}

func NewRefreshRepoTask() func(*Config) error {
	return func(config *Config) error {
		gitRepoCache := config.GitRepoCache
		if gitRepoCache == nil {
			return fmt.Errorf("param GitRepoCache is not set: %w", ErrBadConfiguration)
		}

		config.RefreshRepoTask = tasks.NewRefreshRepoTask(gitRepoCache)

		return nil
	}
}

func NewRefreshTemplateDataTask() func(*Config) error {
	return func(config *Config) error {
		content := config.Content
		if content == nil {
			return fmt.Errorf("param Content is not set: %w", ErrBadConfiguration)
		}

		gitRepoCache := config.GitRepoCache
		if gitRepoCache == nil {
			return fmt.Errorf("param GitRepoCache is not set: %w", ErrBadConfiguration)
		}

		templateData := config.TemplateData
		if templateData == nil {
			return fmt.Errorf("param TemplateData is not set: %w", ErrBadConfiguration)
		}

		config.RefreshTemplateDataTask = tasks.NewRefreshTemplateDataTask(
			content,
			gitRepoCache,
			templateData,
		)

		return nil
	}
}

func NewRepoMonitor() func(*Config) error {
	return func(config *Config) error {
		gitHub := config.GitHub
		if gitHub == nil {
			return fmt.Errorf("param GitHub is not set: %w", ErrBadConfiguration)
		}

		refreshRepoTask := config.RefreshRepoTask
		if refreshRepoTask == nil {
			return fmt.Errorf("param RefreshRepoTask is not set: %w", ErrBadConfiguration)
		}

		refreshTemplateDataTask := config.RefreshTemplateDataTask
		if refreshTemplateDataTask == nil {
			return fmt.Errorf("param RefreshTemplateDataTask is not set: %w", ErrBadConfiguration)
		}

		config.RepoMonitor = github.NewRepoMonitor(
			gitHub,
			[]github.OnUpdateTask{
				refreshRepoTask,
				refreshTemplateDataTask,
			},
		)

		return nil
	}
}

func NewServer() func(*Config) error {
	return func(config *Config) error {
		templateData := config.TemplateData
		if templateData == nil {
			return fmt.Errorf("param TemplateData is not set: %w", ErrBadConfiguration)
		}

		config.Server = web.NewServer(templateData)

		return nil
	}
}
