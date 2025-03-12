package main

import (
	"fmt"
	"testing"
)

func Test_LocalGitRepo_FindFileLastCommit(t *testing.T) {
	gitRepo := &LocalGitRepo{
		repoDirPath,
	}

	commitInfo := gitRepo.FindFileLastCommit("content/pl/docs/_index.md")
	fmt.Printf("%+v", &commitInfo)
}
