//go:build e2e_test

package pullreq_test

import (
	"context"
	"testing"
	"time"

	"github.com/dkarczmarski/go-kweb-lang/github"
	"github.com/dkarczmarski/go-kweb-lang/pullreq"
	"github.com/dkarczmarski/go-kweb-lang/store"
)

func TestFilePRFinder_Update_E2E(t *testing.T) {
	gitHub := github.NewGitHub(
		github.WithDefaults(),
		github.WithThrottle(3*time.Second),
	)
	cacheStore := store.NewFileStore(t.TempDir())
	filePRFinder := pullreq.NewFilePRFinder(
		gitHub,
		cacheStore,
		func(config *pullreq.FilePRFinderConfig) {
			config.PerPage = 10
		},
	)

	err := filePRFinder.Update(context.Background(), "pl")
	if err != nil {
		t.Fatal(err)
	}
}
