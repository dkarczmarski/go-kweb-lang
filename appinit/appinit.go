package appinit

import (
	"errors"
	"fmt"
	"log"
	"os"
	"strconv"
	"strings"
	"time"

	"go-kweb-lang/dashboard"
	"go-kweb-lang/web"

	"go-kweb-lang/git"
	"go-kweb-lang/githist"
	"go-kweb-lang/github"
	"go-kweb-lang/githubmon"
	"go-kweb-lang/gitseek"
	"go-kweb-lang/langcnt"
	"go-kweb-lang/pullreq"
	"go-kweb-lang/store"
	"go-kweb-lang/tasks"
)

type Config struct {
	RepoDir              string
	CacheDir             string
	LangCodes            []string
	RunOnce              bool
	RunInterval          int
	GitHubToken          string
	GitHubUserAgent      string
	GitHubTokenFile      string
	LangCodesProvider    *langcnt.LangCodesProvider
	GitRepo              *git.Git
	CacheStore           *store.FileStore
	DashboardStore       *dashboard.Store
	GitRepoHist          *githist.GitHist
	GitSeek              *gitseek.GitSeek
	GitHub               *github.GitHub
	FilePRFinder         *pullreq.FilePRFinder
	RefreshRepoTask      *tasks.RefreshRepoTask
	RefreshPRTask        *tasks.RefreshPRTask
	RefreshTask          *tasks.RefreshTask
	RefreshDashboardTask *dashboard.RefreshTask
	GitHubMonitor        *githubmon.Monitor
	SkipGitChecking      bool
	SkipPRChecking       bool
	NoWeb                bool
	WebHTTPAddr          string
	Server               *web.Server
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
		config.WebHTTPAddr = ":8080"

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
		getEnvValue("NO_WEB", func(value string) {
			config.NoWeb, _ = strconv.ParseBool(value)
		})
		getEnvValue("WEB_HTTP_ADDR", func(value string) {
			config.WebHTTPAddr = value
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
	flagSkipGit *bool,
	flagSkipPR *bool,
	flagNoWeb *bool,
	flagWebHTTPAddr *string,
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

		if *flagSkipGit {
			config.SkipPRChecking = *flagSkipGit
		}

		if *flagSkipPR {
			config.SkipPRChecking = *flagSkipPR
		}

		if *flagNoWeb {
			config.NoWeb = *flagNoWeb
		}

		if len(*flagWebHTTPAddr) > 0 {
			config.WebHTTPAddr = *flagWebHTTPAddr
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
			log.Printf("SKIP_GIT: %v", config.SkipGitChecking)
			log.Printf("SKIP_PR: %v", config.SkipPRChecking)
			log.Printf("NO_WEB: %v", config.NoWeb)
			log.Printf("WEB_HTTP_ADDR: %s", config.WebHTTPAddr)
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

			lines := strings.SplitN(string(b), "\n", 3)

			token := strings.TrimSpace(lines[0])
			if skipEmptyFile && len(token) == 0 {
				log.Printf("github token file is empty")
			} else {
				config.GitHubToken = token
			}

			if len(lines) > 1 {
				userAgent := strings.TrimSpace(lines[1])
				if len(userAgent) != 0 {
					config.GitHubUserAgent = userAgent
				}
			}
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

func NewCacheStore() func(*Config) error {
	return func(config *Config) error {
		cacheDirPath := config.CacheDir
		if len(cacheDirPath) == 0 {
			return fmt.Errorf("param CacheDir is not set: %w", ErrBadConfiguration)
		}

		config.CacheStore = store.NewFileStore(cacheDirPath)

		return nil
	}
}

func NewDashboardStore() func(*Config) error {
	return func(config *Config) error {
		cacheStore := config.CacheStore
		if cacheStore == nil {
			return fmt.Errorf("param CacheStore is not set: %w", ErrBadConfiguration)
		}

		config.DashboardStore = dashboard.NewStore(cacheStore)

		return nil
	}
}

func NewGitRepoHist() func(*Config) error {
	return func(config *Config) error {
		gitRepo := config.GitRepo
		if gitRepo == nil {
			return fmt.Errorf("param GitRepo is not set: %w", ErrBadConfiguration)
		}

		cacheStore := config.CacheStore
		if cacheStore == nil {
			return fmt.Errorf("param CacheStore is not set: %w", ErrBadConfiguration)
		}

		config.GitRepoHist = githist.New(gitRepo, cacheStore)

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

		cacheStore := config.CacheStore
		if cacheStore == nil {
			return fmt.Errorf("param CacheStore is not set: %w", ErrBadConfiguration)
		}

		config.GitSeek = gitseek.New(gitRepo, gitRepoHist, cacheStore)

		return nil
	}
}

func RegisterGitSeekInvalidator() func(*Config) error {
	return func(config *Config) error {
		gitRepoHist := config.GitRepoHist
		if gitRepoHist == nil {
			return fmt.Errorf("param GitRepoHist is not set: %w", ErrBadConfiguration)
		}

		gitSeeker := config.GitSeek
		if gitSeeker == nil {
			return fmt.Errorf("param GitSeek is not set: %w", ErrBadConfiguration)
		}

		gitRepoHist.RegisterInvalidator(gitSeeker)

		return nil
	}
}

func NewGitHub() func(*Config) error {
	return func(config *Config) error {
		config.GitHub = github.NewGitHub(
			github.WithDefaults(),
			github.WithAuthorization(config.GitHubToken, config.GitHubUserAgent),
			// todo: adjust this value when no authorization token is used
			//
			// magic number,
			// with authorization github allows at most 30 calls per minute, so
			// for safety we use a 3-second delay between requests
			github.WithThrottle(3*time.Second),
		)

		return nil
	}
}

func NewFilePRFinder() func(*Config) error {
	return func(config *Config) error {
		gitHub := config.GitHub
		if gitHub == nil {
			return fmt.Errorf("param GitHub is not set: %w", ErrBadConfiguration)
		}

		cacheStore := config.CacheStore
		if cacheStore == nil {
			return fmt.Errorf("param CacheStore is not set: %w", ErrBadConfiguration)
		}

		config.FilePRFinder = pullreq.NewFilePRFinder(gitHub, cacheStore)

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

func NewRefreshDashboardTask() func(*Config) error {
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

		dashboardStore := config.DashboardStore
		if dashboardStore == nil {
			return fmt.Errorf("param DashboardStore is not set: %w", ErrBadConfiguration)
		}

		config.RefreshDashboardTask = dashboard.NewRefreshTask(
			langCodesProvider,
			gitSeeker,
			filePRFinder,
			dashboardStore,
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

		refreshDashboardTask := config.RefreshDashboardTask
		if refreshDashboardTask == nil {
			return fmt.Errorf("param RefreshDashboardTask is not set: %w", ErrBadConfiguration)
		}

		config.RefreshTask = tasks.NewRefreshTask(refreshRepoTask, refreshPRTask, refreshDashboardTask)

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

		cacheStore := config.CacheStore
		if cacheStore == nil {
			return fmt.Errorf("param CacheStore is not set: %w", ErrBadConfiguration)
		}

		config.GitHubMonitor = githubmon.NewMonitor(
			gh,
			langCodesProvider,
			githubmon.NewMonitorFileStorage(cacheStore),
			config.SkipGitChecking,
			config.SkipPRChecking,
		)

		return nil
	}
}

func NewServer() func(*Config) error {
	return func(config *Config) error {
		if config.NoWeb {
			return nil
		}

		dashboardStore := config.DashboardStore
		if dashboardStore == nil {
			return fmt.Errorf("param DashboardStore is not set: %w", ErrBadConfiguration)
		}

		webHTTPAddr := config.WebHTTPAddr
		if len(webHTTPAddr) == 0 {
			return fmt.Errorf("param WebHTTPAddr is not set: %w", ErrBadConfiguration)
		}

		config.Server = web.NewServer(webHTTPAddr, dashboardStore)

		return nil
	}
}
