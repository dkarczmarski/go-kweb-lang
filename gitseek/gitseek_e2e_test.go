//go:build e2e_test

package gitseek_test

import (
	"context"
	"reflect"
	"testing"

	"go-kweb-lang/git"

	"go-kweb-lang/gitseek"
)

func TestGitSeek_CheckFiles_E2E_issue1(t *testing.T) {
	ctx := context.Background()
	repoDir := t.TempDir()
	gitRepo := git.NewRepo(repoDir)

	if err := gitRepo.Create(ctx, "https://github.com/kubernetes/website"); err != nil {
		t.Fatal(err)
	}

	for _, tc := range []struct {
		name        string
		langRelPath string
		expected    []gitseek.FileInfo
	}{
		{
			name:        "direct commit",
			langRelPath: "docs/reference/glossary/kubelet.md",
			expected: []gitseek.FileInfo{
				{
					LangRelPath: "docs/reference/glossary/kubelet.md",
					LangCommit: git.CommitInfo{
						CommitID: "94d89c5f4cfbb8a8c6ea8adb712237f72eb32ec1",
						DateTime: "2021-06-22T18:26:54+09:00",
						Comment:  "[pl] Remove exec permission on markdown files",
					},
					OriginFileStatus: "MODIFIED",
					OriginUpdates: []gitseek.OriginUpdate{
						{
							Commit: git.CommitInfo{
								CommitID: "2a4a506919b01acaffbd33fc09928ae217454b97",
								DateTime: "2023-09-23T16:20:57-07:00",
								Comment:  "add link for kubelet and cloud-controller-manager (#40931)",
							},
							MergePoint: nil,
						},
					},
				},
			},
		},
		{
			name:        "commit in a separate branch",
			langRelPath: "docs/sitemap.md",
			expected: []gitseek.FileInfo{
				{
					LangRelPath: "docs/sitemap.md",
					LangCommit: git.CommitInfo{
						CommitID: "0adc7047a5b607cc7987322c3b8209119131869d",
						DateTime: "2020-01-14T06:51:24-08:00",
						Comment:  "Init Polish localization (#18419) (#18659)",
					},
					OriginFileStatus: "NOT_EXIST",
					OriginUpdates: []gitseek.OriginUpdate{
						{
							Commit: git.CommitInfo{
								CommitID: "b1fb333fad2301d4d6c30548fb001687f42e24c1",
								DateTime: "2020-09-30T20:30:50+05:00",
								Comment:  "Delete sitemap.md",
							},
							MergePoint: &git.CommitInfo{
								CommitID: "e2233f570a5e3e3ffdc152050d5273e9cdcfe5b1",
								DateTime: "2020-09-30T12:04:54-07:00",
								Comment:  "Merge pull request #24270 from roshnaeem/patch-1",
							},
						},
					},
				},
			},
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			if err := gitRepo.Checkout(ctx, "5cd0e54824949460b8985e68de9f49f9cd5052d9"); err != nil {
				t.Fatal(err)
			}

			gitSeek := gitseek.New(gitRepo)

			fileInfos, err := gitSeek.CheckFiles(ctx, []string{tc.langRelPath}, "pl")
			if err != nil {
				t.Fatal(err)
			}

			if !reflect.DeepEqual(tc.expected, fileInfos) {
				t.Errorf("unexpected result\nexpected: %+v\nactual  : %+v", tc.expected, fileInfos)
			}
		})
	}
}
