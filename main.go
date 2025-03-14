package main

import (
	"encoding/json"
	"fmt"
	"go-kweb-lang/git"
	"go-kweb-lang/gitcache"
	"go-kweb-lang/seek"
	"go-kweb-lang/web"
	"log"
)

var repoDirPath = "../kubernetes-website"

func Run() {
	gitRepo := git.NewRepo(repoDirPath)
	gitRepoCache := gitcache.New(gitRepo, "cache")
	seeker := seek.NewGitLangSeeker(gitRepoCache)

	if err := gitRepoCache.PullRefresh(); err != nil {
		log.Fatal(err)
	}

	langRelPaths, err := gitRepoCache.ListFiles("/content/pl")
	if err != nil {
		log.Fatal(err)
	}

	fileInfos := seeker.CheckFiles(langRelPaths)

	b, err := json.MarshalIndent(&fileInfos, "", "\t")
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println(string(b))

	model := web.BuildTableModel(fileInfos)

	templateData := &web.TemplateData{}
	templateData.Set(model)

	server := web.NewServer(templateData)
	if err := server.ListenAndServe(); err != nil {
		log.Fatal(err)
	}
}

func main() {
	Run()
}
