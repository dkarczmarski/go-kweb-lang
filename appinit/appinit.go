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
var ErrInvalidEnvVars = errors.New("invalid environment variables")

const (
	githubTokenFileLineLimit = 3
	githubThrottleDelay      = 3 * time.Second
)

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

func ParseEnvParams() func(*Config) error {
	return func(config *Config) error {
		var errs []string

		if v, ok := env("REPO_DIR"); ok {
			config.RepoDir = v
		}

		if v, ok := env("CACHE_DIR"); ok {
			config.CacheDir = v
		}

		if v, ok := env("LANG_CODES"); ok {
			config.LangCodes = parseLangCodes(v)
		}

		errs = parseEnvBool("RUN_ONCE", &config.RunOnce, errs)
		errs = parseEnvInt("RUN_INTERVAL", &config.RunInterval, errs)

		if v, ok := env("GITHUB_TOKEN"); ok {
			config.GitHubToken = v
		}

		if v, ok := env("GITHUB_TOKEN_FILE"); ok {
			config.GitHubTokenFile = v
		}

		errs = parseEnvBool("NO_WEB", &config.NoWeb, errs)

		if v, ok := env("WEB_HTTP_ADDR"); ok {
			config.WebHTTPAddr = v
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
		applyFlagLangCodes(flagLangCodes, &config.LangCodes)
		applyFlagString(flagRepoDir, &config.RepoDir)
		applyFlagString(flagCacheDir, &config.CacheDir)
		applyFlagBoolTrue(flagOnce, &config.RunOnce)
		applyFlagIntPositive(flagInterval, &config.RunInterval)
		applyFlagString(flagGitHubToken, &config.GitHubToken)
		applyFlagString(flagGitHubTokenFile, &config.GitHubTokenFile)
		applyFlagBoolTrue(flagSkipGit, &config.SkipGitChecking)
		applyFlagBoolTrue(flagSkipPR, &config.SkipPRChecking)
		applyFlagBoolTrue(flagNoWeb, &config.NoWeb)
		applyFlagString(flagWebHTTPAddr, &config.WebHTTPAddr)

		return nil
	}
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
		*target = parseLangCodes(*flag)
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

func ShowParams(withPrint bool) func(*Config) error {
	return func(config *Config) error {
		if withPrint {
			log.Printf("LANG_CODES: %s", strings.Join(config.LangCodes, ","))
			log.Printf("REPO_DIR: %s", config.RepoDir)
			log.Printf("CACHE_DIR: %s", config.CacheDir)
			log.Printf("RUN_ONCE: %v", config.RunOnce)
			log.Printf("RUN_INTERVAL: %v", config.RunInterval)

			if config.GitHubToken != "" {
				log.Printf("GITHUB_TOKEN: (set, len=%d)", len(config.GitHubToken))
			} else {
				log.Printf("GITHUB_TOKEN: (empty)")
			}

			log.Printf("GITHUB_TOKEN_FILE: %s", config.GitHubTokenFile)
			log.Printf("SKIP_GIT: %v", config.SkipGitChecking)
			log.Printf("SKIP_PR: %v", config.SkipPRChecking)
			log.Printf("NO_WEB: %v", config.NoWeb)
			log.Printf("WEB_HTTP_ADDR: %s", config.WebHTTPAddr)
		}

		return nil
	}
}

func parseLangCodes(langCodesStr string) []string {
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

func ReadGitHubTokenFile(skipFileNotExist, skipEmptyFile bool) func(*Config) error {
	return func(config *Config) error {
		if len(config.GitHubTokenFile) == 0 {
			return nil
		}

		tokenData, err := os.ReadFile(config.GitHubTokenFile)
		if err != nil {
			if skipFileNotExist && os.IsNotExist(err) {
				log.Printf("github token file does not exist")

				return nil
			}

			return fmt.Errorf("error while reading github token file %v: %w",
				config.GitHubTokenFile, err)
		}

		parseGitHubTokenData(config, tokenData, skipEmptyFile)

		return nil
	}
}

func parseGitHubTokenData(config *Config, tokenData []byte, skipEmptyFile bool) {
	lines := strings.SplitN(string(tokenData), "\n", githubTokenFileLineLimit)

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
			github.WithThrottle(githubThrottleDelay),
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
		gitHubClient := config.GitHub
		if gitHubClient == nil {
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
			gitHubClient,
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
