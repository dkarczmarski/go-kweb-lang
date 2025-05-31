//go:build e2e_test

package gitseek_test

import (
	"context"
	"os"
	"path/filepath"
	"reflect"
	"testing"

	"go-kweb-lang/git"
	"go-kweb-lang/githist"
	"go-kweb-lang/gitseek"
	"go-kweb-lang/store"
)

func TestGitSeek_CheckFiles_E2E_issue1(t *testing.T) {
	ctx := context.Background()
	testDir := t.TempDir()
	repoDir := filepath.Join(testDir, "repo")
	mustMkDir(t, repoDir)
	cacheDir := filepath.Join(testDir, "cache")
	mustMkDir(t, cacheDir)
	storeCache := store.NewFileStore(cacheDir)
	gitRepo := git.NewRepo(repoDir)
	gitRepoHist := githist.New(gitRepo, storeCache)

	if err := gitRepo.Create(ctx, "https://github.com/kubernetes/website"); err != nil {
		t.Fatal(err)
	}

	for _, tc := range []struct {
		name        string
		langRelPath string
		expected    []gitseek.FileInfo
	}{
		{
			name:        "origin has direct commit and lang has commit in a separate branch",
			langRelPath: "docs/reference/glossary/kubelet.md",
			expected: []gitseek.FileInfo{
				{
					LangRelPath: "docs/reference/glossary/kubelet.md",
					LangLastCommit: git.CommitInfo{
						CommitID: "94d89c5f4cfbb8a8c6ea8adb712237f72eb32ec1",
						DateTime: "2021-06-22T18:26:54+09:00",
						Comment:  "[pl] Remove exec permission on markdown files",
					},
					LangForkCommit: &git.CommitInfo{
						CommitID: "f193a5fdf61781618c7951001ca9d4c862e2173d",
						DateTime: "2021-06-21T21:45:58-07:00",
						Comment:  "Merge pull request #28532 from lcc3108/patch-1",
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
			name:        "origin has commit in a separate branch and lang has a direct commit",
			langRelPath: "docs/sitemap.md",
			expected: []gitseek.FileInfo{
				{
					LangRelPath: "docs/sitemap.md",
					LangLastCommit: git.CommitInfo{
						CommitID: "0adc7047a5b607cc7987322c3b8209119131869d",
						DateTime: "2020-01-14T06:51:24-08:00",
						Comment:  "Init Polish localization (#18419) (#18659)",
					},
					LangForkCommit:   nil,
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

			cacheDir := t.TempDir()
			storeCache := store.NewFileStore(cacheDir)
			gitSeek := gitseek.New(gitRepo, gitRepoHist, storeCache)

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

func mustMkDir(t testing.TB, path string) {
	t.Helper()

	if err := os.Mkdir(path, 0o700); err != nil {
		t.Fatal(err)
	}
}
