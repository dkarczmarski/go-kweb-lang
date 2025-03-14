package main

import (
	"go-kweb-lang/git"
	"go-kweb-lang/gitcache"
	"go-kweb-lang/github"
	"go-kweb-lang/tasks"
	"go-kweb-lang/web"
	"log"
)

var repoDirPath = "../kubernetes-website"

func Run() {
	gitRepoCache := gitcache.New(git.NewRepo(repoDirPath), "cache")
	gitHub := github.New()

	templateData := &web.TemplateData{}

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

	server := web.NewServer(templateData)
	if err := server.ListenAndServe(); err != nil {
		log.Fatal(err)
	}
}

func main() {
	Run()
}
