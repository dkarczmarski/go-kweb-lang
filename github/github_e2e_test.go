//go:build e2e_test

package github_test

import (
	"log"
	"testing"

	"go-kweb-lang/github"
)

func TestGitHub_GetCommitFiles_E2E(t *testing.T) {
	gh := github.New()

	files, err := gh.GetCommitFiles("f9ef60a9cf2ce7fdc4e242c292d8ed728deab912")
	if err != nil {
		t.Fatal(err)
	}

	t.Log(files)
}

func TestGitHub_GetPRCommits_E2E(t *testing.T) {
	gh := github.New()

	commitIds, err := gh.GetPRCommits(50193)
	if err != nil {
		t.Fatal(err)
	}

	t.Log(commitIds)
}

func TestGitHub_PRSearch_E2E(t *testing.T) {
	gh := github.New()

	var prs []github.PRItem

	var maxUpdatedAt string
	for safetyCounter := 10; safetyCounter > 0; safetyCounter-- {
		result, err := gh.PRSearch(
			github.PRSearchFilter{
				LangCode:    "pl",
				UpdatedFrom: maxUpdatedAt,
				OnlyOpen:    true,
			},
			github.PageRequest{
				Sort:    "updated",
				Order:   "asc",
				PerPage: 4,
			},
		)
		if err != nil {
			t.Fatal(err)
		}
		if len(result.Items) == 0 {
			break
		}

		for _, it := range result.Items {
			log.Printf("%v", it)
		}

		prs = append(prs, result.Items...)

		maxUpdatedAt = result.Items[len(result.Items)-1].UpdatedAt

		t.Log(result)
	}

	t.Log(prs)
}
