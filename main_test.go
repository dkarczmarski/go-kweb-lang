package main

import (
	"fmt"
	"go-kweb-lang/git"
	"testing"
)

func Test_LocalGitRepo_FindFileLastCommit(t *testing.T) {
	gitRepo := git.NewRepo(repoDirPath)

	commitInfo, err := gitRepo.FindFileLastCommit("content/pl/docs/_index.md")
	if err != nil {
		t.Fatal(err)
	}
	fmt.Printf("%+v", &commitInfo)
}
