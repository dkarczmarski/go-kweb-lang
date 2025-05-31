//go:build e2e_test

package git_test

import (
	"context"
	"reflect"
	"testing"

	"go-kweb-lang/git"
)

func TestGit_ListFilesInCommit_E2E(t *testing.T) {
	ctx := context.Background()
	gitRepo := git.NewRepo(t.TempDir())

	if err := gitRepo.Create(ctx, "https://github.com/kubernetes/website"); err != nil {
		t.Fatal(err)
	}

	for _, tc := range []struct {
		name     string
		commitID string
		expected []string
	}{
		{
			name:     "merge pull request commit",
			commitID: "fe0abbcb4b7886ea8ce25f31042037a331b1423b",
			expected: nil,
		},
		{
			name:     "regular commit that is on the main branch",
			commitID: "f4cc9ccc61d6faef60e2e162658d7ff02a42435b",
			expected: []string{
				"content/ja/blog/_posts/2025-03-04-sig-etcd-spotlight/index.md",
				"content/ja/blog/_posts/2025-03-04-sig-etcd-spotlight/stats.png",
				"content/ja/blog/_posts/2025-03-04-sig-etcd-spotlight/stats2.png",
			},
		},
		{
			name:     "regular commit that is not on the main branch",
			commitID: "18eee9913cf7eccd5b68348e5de1164630f13d34",
			expected: []string{
				"content/pl/docs/concepts/workloads/controllers/_index.md",
			},
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			files, err := gitRepo.ListFilesInCommit(ctx, tc.commitID)
			if err != nil {
				t.Errorf("unexpected error: %v", err)
				return
			}

			if !reflect.DeepEqual(tc.expected, files) {
				t.Errorf("unexpected result: %v", files)
			}
		})
	}
}
