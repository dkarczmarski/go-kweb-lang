package appinit

import (
	"context"
	"errors"
	"fmt"
	"log"
	"os"
	"strings"

	"go-kweb-lang/git"
	"go-kweb-lang/github"
	"go-kweb-lang/gitpc"
	"go-kweb-lang/langcnt"
	"go-kweb-lang/pullreq"
	"go-kweb-lang/tasks"
	"go-kweb-lang/web"
)

type Config struct {
	RepoDirPath             string
	CacheDirPath            string
	AllowedLangCodes        []string
	Content                 *langcnt.Content
	GitRepo                 git.Repo
	TemplateData            *web.TemplateData
	GitRepoProxyCache       *gitpc.ProxyCache
	GitHub                  github.GitHub
	FilePRFinder            *pullreq.FilePRFinder
	RefreshRepoTask         *tasks.RefreshRepoTask
	RefreshTemplateDataTask *tasks.RefreshTemplateDataTask
	RefreshPRTask           *tasks.RefreshPRTask
	RepoMonitor             *github.RepoMonitor
	PRMonitor               *github.PRMonitor
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
		config.AllowedLangCodes = parseAllowedLangs(allowedLangs)

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
		content.SetAllowedLangCodes(config.AllowedLangCodes)

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

		config.GitRepoProxyCache = gitpc.New(gitRepo, cacheDirPath)

		return nil
	}
}

func NewGitHub() func(*Config) error {
	return func(config *Config) error {
		config.GitHub = github.New(func(config *github.ClientConfig) {
			config.Token = os.Getenv("GITHUB_TOKEN")
		})

		return nil
	}
}

func NewFilePRFinder() func(*Config) error {
	return func(config *Config) error {
		gitHub := config.GitHub
		if gitHub == nil {
			return fmt.Errorf("param GitHub is not set: %w", ErrBadConfiguration)
		}

		cacheDirPath := config.CacheDirPath
		if len(cacheDirPath) == 0 {
			return fmt.Errorf("param CacheDirPath is not set: %w", ErrBadConfiguration)
		}

		config.FilePRFinder = pullreq.NewFilePRFinder(gitHub, cacheDirPath)

		return nil
	}
}

func NewRefreshRepoTask() func(*Config) error {
	return func(config *Config) error {
		gitRepoProxyCache := config.GitRepoProxyCache
		if gitRepoProxyCache == nil {
			return fmt.Errorf("param ProxyCache is not set: %w", ErrBadConfiguration)
		}

		config.RefreshRepoTask = tasks.NewRefreshRepoTask(gitRepoProxyCache)

		return nil
	}
}

func NewRefreshTemplateDataTask() func(*Config) error {
	return func(config *Config) error {
		content := config.Content
		if content == nil {
			return fmt.Errorf("param Content is not set: %w", ErrBadConfiguration)
		}

		gitRepoProxyCache := config.GitRepoProxyCache
		if gitRepoProxyCache == nil {
			return fmt.Errorf("param ProxyCache is not set: %w", ErrBadConfiguration)
		}

		filePRFinder := config.FilePRFinder
		if filePRFinder == nil {
			return fmt.Errorf("param FilePRFinder is not set: %w", ErrBadConfiguration)
		}

		templateData := config.TemplateData
		if templateData == nil {
			return fmt.Errorf("param TemplateData is not set: %w", ErrBadConfiguration)
		}

		config.RefreshTemplateDataTask = tasks.NewRefreshTemplateDataTask(
			content,
			gitRepoProxyCache,
			filePRFinder,
			templateData,
		)

		return nil
	}
}

func NewRefreshPRTask() func(*Config) error {
	return func(config *Config) error {
		filePRFinder := config.FilePRFinder
		if filePRFinder == nil {
			return fmt.Errorf("param FilePRFinder is not set: %w", ErrBadConfiguration)
		}

		content := config.Content
		if content == nil {
			return fmt.Errorf("param Content is not set: %w", ErrBadConfiguration)
		}

		config.RefreshPRTask = tasks.NewRefreshPRTask(filePRFinder, content)

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

type githubOnPRUpdateTaskAdapter func(ctx context.Context) error

func (f githubOnPRUpdateTaskAdapter) Run(ctx context.Context, langCode string) error {
	return f(ctx)
}

func NewPRMonitor() func(*Config) error {
	return func(config *Config) error {
		gitHub := config.GitHub
		if gitHub == nil {
			return fmt.Errorf("param GitHub is not set: %w", ErrBadConfiguration)
		}

		cacheDirPath := config.CacheDirPath
		if len(cacheDirPath) == 0 {
			return fmt.Errorf("param CacheDirPath is not set: %w", ErrBadConfiguration)
		}

		content := config.Content
		if content == nil {
			return fmt.Errorf("param Content is not set: %w", ErrBadConfiguration)
		}

		refreshPRTask := config.RefreshPRTask
		if refreshPRTask == nil {
			return fmt.Errorf("param RefreshPRTask is not set: %w", ErrBadConfiguration)
		}

		refreshTemplateDataTask := config.RefreshTemplateDataTask
		if refreshTemplateDataTask == nil {
			return fmt.Errorf("param RefreshTemplateDataTask is not set: %w", ErrBadConfiguration)
		}

		config.PRMonitor = github.NewPRMonitor(
			gitHub,
			cacheDirPath,
			content,
			[]github.OnPRUpdateTask{
				refreshPRTask,
				githubOnPRUpdateTaskAdapter(refreshTemplateDataTask.Run),
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
