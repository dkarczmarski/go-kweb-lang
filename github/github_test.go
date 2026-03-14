//nolint:testpackage
package github

import (
	"errors"
	"testing"
)

func TestGitHubBuildPRSearchURL(t *testing.T) {
	t.Parallel()

	gh := &GitHub{
		baseURL: "https://api.github.com",
	}

	got, err := gh.buildPRSearchURL(
		PRSearchFilter{
			OnlyOpen:    true,
			LangCode:    "pl",
			UpdatedFrom: "2024-12-01",
		},
		PageRequest{
			Sort:    "updated",
			Order:   "asc",
			PerPage: 4,
		},
	)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	want := "https://api.github.com/search/issues?" +
		"q=repo:kubernetes/website+is:pr+state:open+label:language/pl+updated:>2024-12-01" +
		"&order=asc&page=1&per_page=4&sort=updated"

	if got != want {
		t.Fatalf("unexpected URL\nwant: %s\ngot : %s", want, got)
	}
}

func TestGitHubBuildPRSearchURL_InvalidBaseURL(t *testing.T) {
	t.Parallel()

	gh := &GitHub{
		baseURL: "https://api.github.com\nbad",
	}

	_, err := gh.buildPRSearchURL(
		PRSearchFilter{
			OnlyOpen: true,
			LangCode: "pl",
		},
		PageRequest{
			Sort:    "updated",
			Order:   "asc",
			PerPage: 4,
		},
	)
	if err == nil {
		t.Fatal("expected error, got nil")
	}

	if !errors.Is(err, ErrInvalidBaseURL) {
		t.Fatalf("expected ErrGitHubBaseURLParseFailed, got %v", err)
	}
}
