//go:build e2e_test

package pullreq_test

import (
	"go-kweb-lang/github"
	"go-kweb-lang/pullreq"
	"testing"
)

func TestPullRequests_Update_E2E(t *testing.T) {
	gh := github.New()
	fpr := &pullreq.PullRequests{
		GitHub:  gh,
		PerPage: 10,
	}

	err := fpr.Update("pl")
	if err != nil {
		t.Fatal(err)
	}
}
