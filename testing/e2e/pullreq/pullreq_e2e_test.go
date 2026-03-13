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

func TestFilePRIndex_RefreshIndex_E2E(t *testing.T) {
	gitHub := github.NewGitHub(
		github.WithDefaults(),
		github.WithThrottle(3*time.Second),
	)
	cacheStore := store.NewFileStore(t.TempDir())
	filePRIndex := pullreq.NewFilePRIndex(
		gitHub,
		cacheStore,
		10,
	)

	err := filePRIndex.RefreshIndex(context.Background(), "pl")
	if err != nil {
		t.Fatal(err)
	}
}
