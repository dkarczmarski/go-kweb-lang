//go:build e2e_test

package pullreq_test

import (
	"context"
	"testing"

	"go-kweb-lang/github"
	"go-kweb-lang/pullreq"
)

func TestFilePRFinder_Update_E2E(t *testing.T) {
	gh := github.New()
	fpr := &pullreq.FilePRFinder{
		GitHub:  gh,
		PerPage: 10,
	}

	err := fpr.Update(context.Background(), "pl")
	if err != nil {
		t.Fatal(err)
	}
}
