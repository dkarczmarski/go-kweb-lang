package main

import (
	"encoding/json"
	"fmt"
	"go-kweb-lang/git"
	"go-kweb-lang/gitcache"
	"go-kweb-lang/seek"
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

	if err := gitRepo.Fetch(); err != nil {
		log.Fatal(err)
	}
	freshCommits, err := gitRepo.FreshCommits()
	if err != nil {
		log.Fatal(err)
	}
	for _, fc := range freshCommits {
		commitFiles, err := gitRepoCache.CommitFiles(fc.CommitId)
		if err != nil {
			log.Fatal(err)
		}
		for _, f := range commitFiles {
			if err := gitRepoCache.InvalidatePath(f); err != nil {
				log.Fatal(err)
			}
		}
	}
	if err := gitRepo.Pull(); err != nil {
		log.Fatal(err)
	}

	result := seeker.CheckFiles(langRelPaths)
	b, err := json.MarshalIndent(&result, "", "\t")
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println(string(b))
}

func main() {
	Run()
}
