package main

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
	"net/http"
	"os"
	"path/filepath"
	"strings"
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

func Run() {
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
		log.Fatal(err)
	}
	if !exists {
		log.Println("repository doest not exist yet")
		if err := gitRepo.Create("https://github.com/kubernetes/website"); err != nil {
			log.Fatal("error while creating kubernetes repository")
		}
		log.Println("repository was created")
	}

	gitRepoCache := gitcache.New(gitRepo, cacheDirPath)
	gitHub := github.New()

	templateData := web.NewTemplateData()

	refreshRepoTask := tasks.NewRefreshRepoTask(gitRepoCache)
	refreshTemplateDataTask := tasks.NewRefreshTemplateDataTask(content, gitRepoCache, templateData)

	if err := refreshRepoTask.Run(); err != nil {
		log.Fatal(err)
	}
	if err := refreshTemplateDataTask.Run(); err != nil {
		log.Fatal(err)
	}

	monitor := github.NewMonitor(gitHub, []github.OnUpdateTask{refreshRepoTask, refreshTemplateDataTask})
	_ = monitor // todo

	log.Println("starting web server")

	server := web.NewServer(templateData)
	if err := server.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
		log.Fatalf("HTTP server failed: %v", err)
	}
}

func main() {
	Run()
}
