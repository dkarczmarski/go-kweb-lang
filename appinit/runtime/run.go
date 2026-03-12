package runtime

import (
	"context"
	"errors"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"path/filepath"
	"syscall"
	"time"

	"github.com/dkarczmarski/go-kweb-lang/appinit/bootstrap"
	"github.com/dkarczmarski/go-kweb-lang/githubmon"
)

const (
	defaultDirPerm      = 0o755
	defaultRetryDelay   = 15 * time.Second
	shutdownGracePeriod = 10 * time.Second
	repoURL             = "https://github.com/kubernetes/website"
)

type repoCreator interface {
	Create(ctx context.Context, repoURL string) error
}

type retryChecker interface {
	RetryCheck(ctx context.Context, retryDelay time.Duration, onUpdateTask githubmon.OnUpdateTask) error
	IntervalCheck(ctx context.Context, intervalDelay time.Duration, onUpdateTask githubmon.OnUpdateTask) error
}

type httpServer interface {
	ListenAndServe() error
	Shutdown(ctx context.Context) error
}

func Run(ctx context.Context, app *bootstrap.App) error {
	ctx, stop := signal.NotifyContext(ctx, syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	if err := runCheckAndRefresh(
		ctx,
		app.Config.RepoDir,
		app.Config.SkipGitChecking,
		app.Config.RunOnce,
		app.Config.RunInterval,
		app.Services.GitRepo,
		app.Services.GitHubMonitor,
		app.Services.RefreshTask,
	); err != nil {
		return err
	}

	return runWebServer(ctx, app.Services.Server)
}

func runCheckAndRefresh(
	ctx context.Context,
	repoDirPath string,
	skipGitChecking bool,
	runOnce bool,
	runInterval int,
	gitRepo repoCreator,
	gitHubMonitor retryChecker,
	refreshTask githubmon.OnUpdateTask,
) error {
	if !skipGitChecking {
		if err := createRepoIfNotExists(ctx, repoDirPath, gitRepo); err != nil {
			return err
		}
	}

	if runOnce {
		if err := gitHubMonitor.RetryCheck(ctx, defaultRetryDelay, refreshTask); err != nil {
			return fmt.Errorf("github monitor retry check failed: %w", err)
		}

		return nil
	}

	if runInterval > 0 {
		go func() {
			intervalDelay := time.Minute * time.Duration(runInterval)

			if err := gitHubMonitor.IntervalCheck(ctx, intervalDelay, refreshTask); err != nil {
				if errors.Is(err, context.Canceled) || errors.Is(err, context.DeadlineExceeded) {
					log.Printf("context cancelled or deadline exceeded: %v", err)

					return
				}

				log.Printf("interval check failed: %v", err)
			}
		}()
	}

	return nil
}

func runWebServer(ctx context.Context, server httpServer) error {
	if server == nil {
		<-ctx.Done()
		log.Println("application stopped")

		return nil
	}

	serverErrCh := make(chan error, 1)

	go func() {
		<-ctx.Done()
		log.Println("server is shutting down")

		shutdownCtx, cancel := context.WithTimeout(context.WithoutCancel(ctx), shutdownGracePeriod)
		defer cancel()

		if err := server.Shutdown(shutdownCtx); err != nil {
			log.Printf("failed to shutdown http server: %v", err)
		}
	}()

	go func() {
		log.Println("starting web server")

		err := server.ListenAndServe()
		if err != nil && !errors.Is(err, http.ErrServerClosed) {
			serverErrCh <- fmt.Errorf("failed to run http server: %w", err)

			return
		}

		serverErrCh <- nil
	}()

	select {
	case <-ctx.Done():
		return nil
	case err := <-serverErrCh:
		return err
	}
}

func createRepoIfNotExists(ctx context.Context, repoDirPath string, gitRepo repoCreator) error {
	exists, err := fileExists(filepath.Join(repoDirPath, ".git"))
	if err != nil {
		return fmt.Errorf("error while checking if a git repository exists: %w", err)
	}

	if exists {
		return nil
	}

	log.Println("repository does not exist yet. creating a new one...")

	if err := os.MkdirAll(repoDirPath, defaultDirPerm); err != nil {
		return fmt.Errorf("error while creating directory %s: %w", repoDirPath, err)
	}

	if err := gitRepo.Create(ctx, repoURL); err != nil {
		return fmt.Errorf("error while creating kubernetes repository: %w", err)
	}

	log.Println("repository was created")

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
