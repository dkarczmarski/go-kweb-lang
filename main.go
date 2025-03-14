package main

import (
	"encoding/json"
	"fmt"
	"go-kweb-lang/git"
	"go-kweb-lang/gitcache"
	"go-kweb-lang/seek"
	"go-kweb-lang/web"
	"log"
	"os"
	"path/filepath"
)

func ListFiles(root string) ([]string, error) {
	var files []string

	err := filepath.WalkDir(root, func(path string, d os.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if !d.IsDir() {
			relPath, err := filepath.Rel(root, path)
			if err != nil {
				return err
			}
			files = append(files, relPath)
		}
		return nil
	})

	if err != nil {
		return nil, err
	}

	return files, nil
}

var repoDirPath = "../kubernetes-website"

func Run() {
	langRelPaths, err := ListFiles(repoDirPath + "/content/pl")
	if err != nil {
		log.Fatal(err)
	}

	gitRepo := git.NewRepo(repoDirPath)
	gitRepoCache := gitcache.New(gitRepo, "cache")
	seeker := seek.NewGitLangSeeker(gitRepoCache)

	if err := gitRepoCache.PullRefresh(); err != nil {
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
