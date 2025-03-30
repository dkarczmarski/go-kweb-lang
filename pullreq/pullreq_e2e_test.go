//go:build e2e_test

package pullreq_test

import (
	"context"
	"testing"

	"go-kweb-lang/github"
	"go-kweb-lang/pullreq"
)

func TestFilePRFinder_Update_E2E(t *testing.T) {
	gitHub := github.New()
	filePRFinder := pullreq.NewFilePRFinder(
		gitHub,
		t.TempDir(),
		func(config *pullreq.FilePRFinderConfig) {
			config.PerPage = 10
		},
	)

	err := filePRFinder.Update(context.Background(), "pl")
	if err != nil {
		t.Fatal(err)
	}
}
