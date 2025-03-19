package main

import (
	"context"
	"errors"
	"fmt"
	"go-kweb-lang/git"
	"go-kweb-lang/gitcache"
	"go-kweb-lang/github"
	"go-kweb-lang/langcnt"
	"go-kweb-lang/tasks"
	"go-kweb-lang/web"
	"log"
	"net/http"
	"os"
	"os/signal"
	"path/filepath"
	"strings"
	"syscall"
	"time"
)

func getEnvOrDefault(key, defaultValue string) string {
	value, ok := os.LookupEnv(key)
	if !ok {
		return defaultValue
	}
	return value
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

func CreateWebServer(ctx context.Context) (*web.Server, error) {
	repoDirPath := getEnvOrDefault("REPO_DIR", "../kubernetes-website")
	cacheDirPath := getEnvOrDefault("CACHE_DIR", "./cache")
	allowedLangs := getEnvOrDefault("ALLOWED_LANGS", "")

	log.Printf("REPO_DIR: %s", repoDirPath)
	log.Printf("CACHE_DIR: %s", cacheDirPath)
	log.Printf("ALLOWED_LANGS: %s", allowedLangs)

	content := &langcnt.Content{RepoDir: repoDirPath}
	content.SetAllowedLang(parseAllowedLangs(allowedLangs))

	gitRepo := git.NewRepo(repoDirPath)

	exists, err := fileExists(filepath.Join(repoDirPath, ".git"))
	if err != nil {
		return nil, fmt.Errorf("error while checking if a git repository exists: %w", err)
	}
	if !exists {
		log.Println("repository doest not exist yet")
		if err := gitRepo.Create(ctx, "https://github.com/kubernetes/website"); err != nil {
			return nil, fmt.Errorf("error while creating kubernetes repository: %w", err)
		}
		log.Println("repository was created")
	}

	gitRepoCache := gitcache.New(gitRepo, cacheDirPath)
	gitHub := github.New()

	templateData := web.NewTemplateData()

	// todo: refactor
	refreshRepoTask := tasks.NewRefreshRepoTask(gitRepoCache)
	refreshTemplateDataTask := tasks.NewRefreshTemplateDataTask(content, gitRepoCache, templateData)

	if err := refreshRepoTask.Run(ctx); err != nil {
		return nil, fmt.Errorf("error while running the refresh repository task: %w", err)
	}
	if err := refreshTemplateDataTask.Run(ctx); err != nil {
		return nil, fmt.Errorf("error while running the refresh template data task: %w", err)
	}

	monitor := github.NewMonitor(gitHub, []github.OnUpdateTask{refreshRepoTask, refreshTemplateDataTask})
	_ = monitor // todo

	server := web.NewServer(templateData)

	return server, nil
}

func main() {
	ctx, _ := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)

	server, err := CreateWebServer(ctx)
	if err != nil {
		log.Fatal(err)
	}

	go func() {
		<-ctx.Done()
		log.Println("server is shutting down")
		ctx, cancelCtx := context.WithTimeout(context.Background(), time.Second*10)
		defer cancelCtx()
		_ = server.Shutdown(ctx)
	}()

	log.Println("starting web server")
	if err := server.ListenAndServe(); err != nil && errors.Is(err, http.ErrServerClosed) {
		log.Fatal(err)
	}
}
