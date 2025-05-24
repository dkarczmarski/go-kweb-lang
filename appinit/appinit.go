package appinit

import (
	"errors"
	"fmt"
	"log"
	"os"
	"strconv"
	"strings"

	"go-kweb-lang/gitseek"

	"go-kweb-lang/git"
	"go-kweb-lang/githist"
	"go-kweb-lang/github"
	"go-kweb-lang/langcnt"
	"go-kweb-lang/pullreq"
	"go-kweb-lang/tasks"
	"go-kweb-lang/web"
)

type Config struct {
	RepoDir                 string
	CacheDir                string
	LangCodes               []string
	RunOnce                 bool
	RunInterval             int
	GitHubToken             string
	GitHubTokenFile         string
	LangCodesProvider       *langcnt.LangCodesProvider
	GitRepo                 git.Repo
	TemplateData            *web.TemplateData
	GitRepoHist             *githist.GitHist
	GitSeek                 *gitseek.GitSeek
	GitHub                  github.GitHub
	FilePRFinder            *pullreq.FilePRFinder
	RefreshRepoTask         *tasks.RefreshRepoTask
	RefreshTemplateDataTask *tasks.RefreshTemplateDataTask
	RefreshPRTask           *tasks.RefreshPRTask
	RefreshTask             *tasks.RefreshTask
	GitHubMonitor           *github.Monitor
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

func SetDefaultParams() func(*Config) error {
	return func(config *Config) error {
		config.RepoDir = "./.appdata/kubernetes-website"
		config.CacheDir = "./.appdata/cache"
		config.GitHubTokenFile = ".github-token.txt"

		return nil
	}
}

func getEnvValue(key string, consumer func(value string)) {
	value, ok := os.LookupEnv(key)
	if ok {
		consumer(value)
	}
}

func ParseEnvParams() func(*Config) error {
	return func(config *Config) error {
		getEnvValue("REPO_DIR", func(value string) {
			config.RepoDir = value
		})
		getEnvValue("CACHE_DIR", func(value string) {
			config.CacheDir = value
		})
		getEnvValue("LANG_CODES", func(value string) {
			config.LangCodes = parseLangCodes(value)
		})
		getEnvValue("RUN_ONCE", func(value string) {
			config.RunOnce, _ = strconv.ParseBool(value)
		})
		getEnvValue("RUN_INTERVAL", func(value string) {
			config.RunInterval, _ = strconv.Atoi(value)
		})
		getEnvValue("GITHUB_TOKEN", func(value string) {
			config.GitHubToken = value
		})
		getEnvValue("GITHUB_TOKEN_FILE", func(value string) {
			config.GitHubTokenFile = value
		})

		return nil
	}
}

func ParseFlagParams(
	flagRepoDir *string,
	flagCacheDir *string,
	flagLangCodes *string,
	flagOnce *bool,
	flagInterval *int,
	flagGitHubToken *string,
	flagGitHubTokenFile *string,
) func(*Config) error {
	return func(config *Config) error {
		if len(*flagLangCodes) > 0 {
			config.LangCodes = parseLangCodes(*flagLangCodes)
		}

		if len(*flagRepoDir) > 0 {
			config.RepoDir = *flagRepoDir
		}

		if len(*flagCacheDir) > 0 {
			config.CacheDir = *flagCacheDir
		}

		// it can only be overridden to true
		if *flagOnce {
			config.RunOnce = *flagOnce
		}

		// it can only be overridden with a non-zero value
		if *flagInterval > 0 {
			config.RunInterval = *flagInterval
		}

		if len(*flagGitHubToken) > 0 {
			config.GitHubToken = *flagGitHubToken
		}

		if len(*flagGitHubTokenFile) > 0 {
			config.GitHubTokenFile = *flagGitHubTokenFile
		}

		return nil
	}
}

func ShowParams(withPrint bool) func(*Config) error {
	return func(config *Config) error {
		if withPrint {
			log.Printf("LANG_CODES: %s", config.LangCodes)
			log.Printf("REPO_DIR: %s", config.RepoDir)
			log.Printf("CACHE_DIR: %s", config.CacheDir)
			log.Printf("RUN_ONCE: %v", config.RunOnce)
			log.Printf("RUN_INTERVAL: %v", config.RunInterval)
			log.Printf("GITHUB_TOKEN: %s", config.GitHubToken)
			log.Printf("GITHUB_TOKEN_FILE: %s", config.GitHubTokenFile)
		}

		return nil
	}
}

func parseLangCodes(s string) []string {
	if len(strings.TrimSpace(s)) == 0 {
		return nil
	}

	subs := strings.Split(s, ",")
	allowedLangCodes := make([]string, 0, len(subs))

	for _, sub := range subs {
		allowedLangCodes = append(allowedLangCodes, strings.TrimSpace(sub))
	}

	return allowedLangCodes
}

func ReadGitHubTokenFile(skipFileNotExist, skipEmptyFile bool) func(*Config) error {
	return func(config *Config) error {
		if len(config.GitHubTokenFile) > 0 {
			b, err := os.ReadFile(config.GitHubTokenFile)
			if err != nil {
				if skipFileNotExist && os.IsNotExist(err) {
					log.Printf("github token file does not exist")

					return nil
				}

				return fmt.Errorf("error while reading github token file %v: %w",
					config.GitHubTokenFile, err)
			}

			value := string(b)

			if skipEmptyFile && len(strings.TrimSpace(value)) == 0 {
				log.Printf("github token file is empty")

				return nil
			}

			config.GitHubToken = value
		}

		return nil
	}
}

func NewLangCodesProvider() func(config *Config) error {
	return func(config *Config) error {
		repoDirPath := config.RepoDir
		if len(repoDirPath) == 0 {
			return fmt.Errorf("param RepoDir is not set: %w", ErrBadConfiguration)
		}

		langCodesProvider := &langcnt.LangCodesProvider{RepoDir: repoDirPath}
		langCodesProvider.SetLangCodesFilter(config.LangCodes)

		config.LangCodesProvider = langCodesProvider

		return nil
	}
}

func NewRepo() func(*Config) error {
	return func(config *Config) error {
		repoDirPath := config.RepoDir
		if len(repoDirPath) == 0 {
			return fmt.Errorf("param RepoDir is not set: %w", ErrBadConfiguration)
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

		cacheDirPath := config.CacheDir
		if len(cacheDirPath) == 0 {
			return fmt.Errorf("param CacheDir is not set: %w", ErrBadConfiguration)
		}

		config.GitRepoHist = githist.New(gitRepo, cacheDirPath)

		return nil
	}
}

func NewGitSeek() func(*Config) error {
	return func(config *Config) error {
		gitRepo := config.GitRepo
		if gitRepo == nil {
			return fmt.Errorf("param GitRepo is not set: %w", ErrBadConfiguration)
		}

		gitRepoHist := config.GitRepoHist
		if gitRepoHist == nil {
			return fmt.Errorf("param GitRepoHist is not set: %w", ErrBadConfiguration)
		}

		config.GitSeek = gitseek.New(gitRepo, gitRepoHist)

		return nil
	}
}

func NewGitHub() func(*Config) error {
	return func(config *Config) error {
		config.GitHub = github.New(func(githubConfig *github.ClientConfig) {
			githubConfig.Token = config.GitHubToken
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

		cacheDirPath := config.CacheDir
		if len(cacheDirPath) == 0 {
			return fmt.Errorf("param CacheDir is not set: %w", ErrBadConfiguration)
		}

		config.FilePRFinder = pullreq.NewFilePRFinder(gitHub, cacheDirPath)

		return nil
	}
}

func NewRefreshRepoTask() func(*Config) error {
	return func(config *Config) error {
		gitRepoHist := config.GitRepoHist
		if gitRepoHist == nil {
			return fmt.Errorf("param GitHist is not set: %w", ErrBadConfiguration)
		}

		config.RefreshRepoTask = tasks.NewRefreshRepoTask(gitRepoHist)

		return nil
	}
}

func NewRefreshTemplateDataTask() func(*Config) error {
	return func(config *Config) error {
		langCodesProvider := config.LangCodesProvider
		if langCodesProvider == nil {
			return fmt.Errorf("param LangCodesProvider is not set: %w", ErrBadConfiguration)
		}

		gitSeeker := config.GitSeek
		if gitSeeker == nil {
			return fmt.Errorf("param GitSeek is not set: %w", ErrBadConfiguration)
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
			langCodesProvider,
			gitSeeker,
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

		langCodesProvider := config.LangCodesProvider
		if langCodesProvider == nil {
			return fmt.Errorf("param LangCodesProvider is not set: %w", ErrBadConfiguration)
		}

		config.RefreshPRTask = tasks.NewRefreshPRTask(filePRFinder, langCodesProvider)

		return nil
	}
}

func NewRefreshTask() func(*Config) error {
	return func(config *Config) error {
		refreshRepoTask := config.RefreshRepoTask
		if refreshRepoTask == nil {
			return fmt.Errorf("param RefreshRepoTask is not set: %w", ErrBadConfiguration)
		}

		refreshPRTask := config.RefreshPRTask
		if refreshPRTask == nil {
			return fmt.Errorf("param RefreshPRTask is not set: %w", ErrBadConfiguration)
		}

		refreshTemplateDataTask := config.RefreshTemplateDataTask
		if refreshTemplateDataTask == nil {
			return fmt.Errorf("param RefreshTemplateDataTask is not set: %w", ErrBadConfiguration)
		}

		config.RefreshTask = tasks.NewRefreshTask(refreshRepoTask, refreshPRTask, refreshTemplateDataTask)

		return nil
	}
}

func NewGitHubMonitor() func(*Config) error {
	return func(config *Config) error {
		gh := config.GitHub
		if gh == nil {
			return fmt.Errorf("param GitHub is not set: %w", ErrBadConfiguration)
		}

		langCodesProvider := config.LangCodesProvider
		if langCodesProvider == nil {
			return fmt.Errorf("param LangCodesProvider is not set: %w", ErrBadConfiguration)
		}

		cacheDirPath := config.CacheDir
		if len(cacheDirPath) == 0 {
			return fmt.Errorf("param CacheDir is not set: %w", ErrBadConfiguration)
		}

		config.GitHubMonitor = github.NewMonitor(
			gh,
			langCodesProvider,
			github.NewMonitorFileStorage(cacheDirPath),
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
