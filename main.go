package main

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"go-kweb-lang/appinit"
	"go-kweb-lang/git"
	"go-kweb-lang/github"
	"go-kweb-lang/tasks"
	"log"
	"net/http"
	"os"
	"os/signal"
	"path/filepath"
	"syscall"
	"time"
)

func createRepoIfNotExists(ctx context.Context, repoDirPath string, gitRepo git.Repo) error {
	exists, err := fileExists(filepath.Join(repoDirPath, ".git"))
	if err != nil {
		return fmt.Errorf("error while checking if a git repository exists: %w", err)
	}
	if !exists {
		log.Println("repository doest not exist yet")
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

func runTasks(
	ctx context.Context,
	refreshRepoTask *tasks.RefreshRepoTask,
	refreshPRTask *tasks.RefreshPRTask,
	refreshTemplateDataTask *tasks.RefreshTemplateDataTask,
) error {
	if err := refreshRepoTask.Run(ctx); err != nil {
		return fmt.Errorf("error while running the refresh repository task: %w", err)
	}
	if err := refreshPRTask.RunAll(ctx); err != nil {
		return fmt.Errorf("error while running the refresh repository task: %w", err) //todo
	}
	if err := refreshTemplateDataTask.Run(ctx); err != nil {
		return fmt.Errorf("error while running the refresh template data task: %w", err)
	}

	return nil
}

func runChecks(
	ctx context.Context,
	delay time.Duration,
	repoMonitor *github.RepoMonitor,
	prMonitor *github.PRMonitor,
) {
	for {
		if err := repoMonitor.CheckRepo(ctx); err != nil {
			log.Printf("error while checking github for changes: %v", err)
		}

		if err := prMonitor.Check(ctx); err != nil {
			log.Printf("error while cheking for PR updates: %v", err)
		}

		select {
		case <-ctx.Done():
			return
		case <-time.After(delay):
		}
	}

}

var flagOnce = flag.Bool("once", false, "run synchronization once at startup")
var flagInterval = flag.Int("interval", 0, "run repeatedly with delay of N minutes between runs")

func main() {
	flag.Parse()

	log.Printf("run once: %v, interval: %v", *flagOnce, *flagInterval)

	ctx, _ := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)

	cfg, err := appinit.Init(
		appinit.GetEnv(true),
		appinit.NewContent(),
		appinit.NewRepo(),
		appinit.NewRepoCache(),
		appinit.NewGitHub(),
		appinit.NewPullRequests(),
		appinit.NewTemplateData(),
		appinit.NewRefreshRepoTask(),
		appinit.NewRefreshTemplateDataTask(),
		appinit.NewRefreshPRTask(),
		appinit.NewRepoMonitor(),
		appinit.NewPRMonitor(),
		appinit.NewServer(),
	)
	if err != nil {
		log.Fatal(err)
	}

	repoDirPath := cfg.RepoDirPath
	gitRepo := cfg.GitRepo
	if err := createRepoIfNotExists(ctx, repoDirPath, gitRepo); err != nil {
		log.Fatal(err)
	}

	if *flagOnce {
		if err := runTasks(ctx, cfg.RefreshRepoTask, cfg.RefreshPRTask, cfg.RefreshTemplateDataTask); err != nil {
			log.Fatal(err)
		}
	}

	if *flagInterval > 0 {
		delay := time.Minute * time.Duration(*flagInterval)

		go runChecks(ctx, delay, cfg.RepoMonitor, cfg.PRMonitor)
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
		log.Fatal(fmt.Errorf("error while running http server: %w", err))
	}
}
