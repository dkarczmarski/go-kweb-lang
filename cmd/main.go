package main

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"path/filepath"
	"syscall"
	"time"

	"go-kweb-lang/github"
	"go-kweb-lang/langcnt"
	"go-kweb-lang/tasks"

	"go-kweb-lang/appinit"
	"go-kweb-lang/git"
)

var (
	flagRepoDir  = flag.String("repo-dir", "", "kubernetes website repository directory path")
	flagCacheDir = flag.String("cache-dir", "",
		"cache directory path")
	flagLangCodes       = flag.String("lang-codes", "", "allowed lang codes")
	flagRunOnce         = flag.Bool("run-once", false, "run synchronization once at startup")
	flagRunInterval     = flag.Int("run-interval", 0, "run repeatedly with delay of N minutes between runs")
	flagGitHubToken     = flag.String("github-token", "", "github api access token")
	flagGitHubTokenFile = flag.String("github-token-file", "", "file path with github api access token")
)

func createRepoIfNotExists(ctx context.Context, repoDirPath string, gitRepo git.Repo) error {
	exists, err := fileExists(filepath.Join(repoDirPath, ".git"))
	if err != nil {
		return fmt.Errorf("error while checking if a git repository exists: %w", err)
	}
	if !exists {
		log.Println("repository does not exist yet. creating a new one...")

		if err := os.MkdirAll(repoDirPath, 0o755); err != nil {
			return fmt.Errorf("error while creating directory %s: %w", repoDirPath, err)
		}

		if err := gitRepo.Create(ctx, "https://github.com/kubernetes/website"); err != nil {
			return fmt.Errorf("error while creating kubernetes repository: %w", err)
		}

		log.Println("repository was created")
	}

	return nil
}

func fileExists(path string) (bool, error) {
	_, err := os.Stat(path)
	if os.IsNotExist(err) {
		return false, nil
	}
	if err != nil {
		return false, fmt.Errorf("error while checking whether file exists: %w", err)
	}

	return true, nil
}

func runOnce(ctx context.Context, langCodesProvider *langcnt.LangCodesProvider, refreshTask *tasks.RefreshTask) {
	langCodes, err := langCodesProvider.LangCodes()
	if err != nil {
		log.Fatal(fmt.Errorf("error while getting available languages: %w", err))
	}

	if err := refreshTask.OnUpdate(ctx, true, langCodes); err != nil {
		log.Fatal(err)
	}
}

func runInterval(
	ctx context.Context,
	gitHubMonitor *github.Monitor,
	refreshTask *tasks.RefreshTask,
	delay time.Duration,
) {
	if err := gitHubMonitor.StartIntervalCheck(ctx, delay, refreshTask); err != nil {
		if !errors.Is(err, context.Canceled) && !errors.Is(err, context.DeadlineExceeded) {
			log.Fatal(err)
		}

		log.Println("context cancelled or deadline exceeded:", err)
	}
}

func Run(ctx context.Context, cfg *appinit.Config) error {
	ctx, _ = signal.NotifyContext(ctx, syscall.SIGINT, syscall.SIGTERM)

	if err := createRepoIfNotExists(ctx, cfg.RepoDir, cfg.GitRepo); err != nil {
		return err
	}

	if cfg.RunOnce && cfg.RunInterval == 0 {
		runOnce(ctx, cfg.LangCodesProvider, cfg.RefreshTask)
	}

	if cfg.RunInterval > 0 {
		if err := cfg.RefreshTemplateDataTask.Run(ctx); err != nil {
			return fmt.Errorf("error while refreshing web model: %w", err)
		}

		go runInterval(
			ctx,
			cfg.GitHubMonitor,
			cfg.RefreshTask,
			time.Minute*time.Duration(cfg.RunInterval),
		)
	}

	server := cfg.Server

	go func() {
		<-ctx.Done()
		log.Println("server is shutting down")
		ctx, cancelCtx := context.WithTimeout(context.Background(), time.Second*10)
		defer cancelCtx()
		_ = server.Shutdown(ctx)
	}()

	log.Println("starting web server")
	if err := server.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
		return fmt.Errorf("error while running http server: %w", err)
	}

	return nil
}

func main() {
	flag.Parse()

	cfg, err := appinit.Init(
		// read params
		appinit.SetDefaultParams(),
		appinit.ParseEnvParams(),
		appinit.ParseFlagParams(
			flagRepoDir,
			flagCacheDir,
			flagLangCodes,
			flagRunOnce,
			flagRunInterval,
			flagGitHubToken,
			flagGitHubTokenFile,
		),
		appinit.ShowParams(true),
		appinit.ReadGitHubTokenFile(true, true),

		// create components
		appinit.NewLangCodesProvider(),
		appinit.NewRepo(),
		appinit.NewRepoCache(),
		appinit.NewGitHub(),
		appinit.NewFilePRFinder(),
		appinit.NewTemplateData(),
		appinit.NewRefreshRepoTask(),
		appinit.NewRefreshTemplateDataTask(),
		appinit.NewRefreshPRTask(),
		appinit.NewRefreshTask(),
		appinit.NewGitHubMonitor(),
		appinit.NewServer(),
	)
	if err != nil {
		log.Fatal(fmt.Errorf("error while application configuration initialization: %w", err))
	}

	if err := Run(context.Background(), cfg); err != nil {
		log.Fatal(err)
	}
}
