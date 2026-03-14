package bootstrap

import (
	"fmt"
	"time"

	"github.com/dkarczmarski/go-kweb-lang/appinit/config"
	"github.com/dkarczmarski/go-kweb-lang/dashboard"
	"github.com/dkarczmarski/go-kweb-lang/filepairs"
	"github.com/dkarczmarski/go-kweb-lang/git"
	"github.com/dkarczmarski/go-kweb-lang/githist"
	"github.com/dkarczmarski/go-kweb-lang/github"
	"github.com/dkarczmarski/go-kweb-lang/githubmon"
	"github.com/dkarczmarski/go-kweb-lang/gitseek"
	"github.com/dkarczmarski/go-kweb-lang/langcnt"
	"github.com/dkarczmarski/go-kweb-lang/pullreq"
	"github.com/dkarczmarski/go-kweb-lang/store"
	"github.com/dkarczmarski/go-kweb-lang/tasks"
	"github.com/dkarczmarski/go-kweb-lang/web"
)

const (
	githubThrottleDelay = 3 * time.Second
	githubPerPage       = 100
)

type Services struct {
	LangCodesProvider    *langcnt.LangCodesProvider
	GitRepo              *git.Git
	CacheStore           *store.JSONFileStore
	DashboardStore       *dashboard.Store
	GitRepoHist          *githist.GitHist
	FilePaths            *filepairs.FilePaths
	PairProviders        *filepairs.PairProviders
	GitSeek              *gitseek.GitSeek
	GitHub               *github.GitHub
	FilePRIndex          *pullreq.FilePRIndex
	RefreshRepoTask      *tasks.RefreshRepoTask
	RefreshPRTask        *tasks.RefreshPRTask
	OnGitHubUpdateTask   *tasks.OnGitHubUpdateTask
	RefreshDashboardTask *tasks.RefreshDashboardTask
	GitHubMonitor        *githubmon.Monitor
	Server               *web.Server
}

func BuildServices(cfg config.Config) (*Services, error) {
	if err := config.Validate(cfg); err != nil {
		return nil, fmt.Errorf("validate config: %w", err)
	}

	//nolint:exhaustruct
	services := &Services{}

	buildCoreServices(cfg, services)
	buildTaskServices(cfg, services)

	if err := buildOptionalServices(cfg, services); err != nil {
		return nil, err
	}

	return services, nil
}

func buildCoreServices(cfg config.Config, services *Services) {
	langCodesProvider := &langcnt.LangCodesProvider{RepoDir: cfg.RepoDir}
	langCodesProvider.SetLangCodesFilter(cfg.LangCodes)
	services.LangCodesProvider = langCodesProvider

	services.GitRepo = git.NewRepo(cfg.RepoDir)
	services.CacheStore = store.NewFileStore(cfg.CacheDir)
	services.DashboardStore = dashboard.NewStore(services.CacheStore)
	services.GitRepoHist = githist.New(services.GitRepo, services.CacheStore)
	services.FilePaths = filepairs.New()

	services.PairProviders = filepairs.NewPairProviders(
		filepairs.NewContentPairProvider(services.GitRepo),
	)

	services.GitSeek = gitseek.New(services.GitRepo, services.GitRepoHist, services.CacheStore)

	services.GitHub = github.NewGitHub(
		github.WithDefaults(),
		github.WithAuthorization(cfg.GitHubToken, cfg.GitHubUserAgent),
		// with authorization github allows at most 30 calls per minute, so
		// for safety we use a 3-second delay between requests
		github.WithThrottle(githubThrottleDelay),
	)

	services.FilePRIndex = pullreq.NewFilePRIndex(services.GitHub, services.CacheStore, githubPerPage)
}

func buildTaskServices(cfg config.Config, services *Services) {
	services.RefreshRepoTask = tasks.NewRefreshRepoTask(
		services.GitRepoHist,
		services.FilePaths,
		services.LangCodesProvider,
		services.GitSeek,
	)

	services.RefreshDashboardTask = tasks.NewRefreshDashboardTask(
		services.LangCodesProvider,
		services.PairProviders,
		services.GitSeek,
		services.FilePRIndex,
		services.DashboardStore,
	)

	services.RefreshPRTask = tasks.NewRefreshPRTask(
		services.FilePRIndex,
		services.LangCodesProvider,
	)

	services.OnGitHubUpdateTask = tasks.NewOnGitHubUpdateTask(
		services.RefreshRepoTask,
		services.RefreshPRTask,
		services.RefreshDashboardTask,
	)

	services.GitHubMonitor = githubmon.NewMonitor(
		services.GitHub,
		services.LangCodesProvider,
		githubmon.NewMonitorFileStorage(services.CacheStore),
		cfg.SkipGitChecking,
		cfg.SkipPRChecking,
	)
}

func buildOptionalServices(cfg config.Config, services *Services) error {
	if cfg.NoWeb {
		return nil
	}

	if len(cfg.WebHTTPAddr) == 0 {
		return fmt.Errorf("param WebHTTPAddr is not set: %w", config.ErrBadConfiguration)
	}

	services.Server = web.NewServer(cfg.WebHTTPAddr, services.DashboardStore)

	return nil
}
