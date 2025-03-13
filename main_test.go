package main

import (
	"fmt"
	"go-kweb-lang/git"
	"testing"
)

func Test_LocalGitRepo_FindFileLastCommit(t *testing.T) {
	gitRepo := git.NewRepo(repoDirPath)

	commitInfo := gitRepo.FindFileLastCommit("content/pl/docs/_index.md")
	fmt.Printf("%+v", &commitInfo)
}
