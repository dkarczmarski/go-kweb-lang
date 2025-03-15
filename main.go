package main

import (
	"go-kweb-lang/git"
	"go-kweb-lang/gitcache"
	"go-kweb-lang/github"
	"go-kweb-lang/tasks"
	"go-kweb-lang/web"
	"log"
	"os"
	"path/filepath"
)

var repoDirPath = "../kubernetes-website"

func fileExists(path string) (bool, error) {
	_, err := os.Stat(path)
	if os.IsNotExist(err) {
		return false, nil
	}
	if err != nil {
		return false, err
	}
	return true, nil
}

func Run() {
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

	gitRepoCache := gitcache.New(gitRepo, "cache")
	gitHub := github.New()

	templateData := web.NewTemplateData()

	refreshRepoTask := tasks.NewRefreshRepoTask(gitRepoCache)
	refreshTemplateDataTask := tasks.NewRefreshTemplateDataTask(gitRepoCache, templateData)

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
	if err := server.ListenAndServe(); err != nil {
		log.Fatal(err)
	}
}

func main() {
	Run()
}
