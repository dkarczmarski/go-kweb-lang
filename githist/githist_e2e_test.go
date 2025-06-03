//go:build e2e_test

package githist_test

import (
	"context"
	"os"
	"path/filepath"
	"reflect"
	"testing"

	"go-kweb-lang/git"
	"go-kweb-lang/githist"
	"go-kweb-lang/store"
)

func TestGitHist_MergeCommitFiles_E2E(t *testing.T) {
	ctx := context.Background()
	testDir := t.TempDir()
	repoDir := filepath.Join(testDir, "repo")
	mustMkDir(t, repoDir)
	cacheDir := filepath.Join(testDir, "cache")
	mustMkDir(t, cacheDir)
	cacheStore := store.NewFileStore(cacheDir)
	gitRepo := git.NewRepo(repoDir)
	gitRepoHist := githist.New(gitRepo, cacheStore)

	if err := gitRepo.Create(ctx, "https://github.com/kubernetes/website"); err != nil {
		t.Fatal(err)
	}

	for _, tc := range []struct {
		name     string
		commitID string
		expected []string
	}{
		{
			name:     "single parent commit on the main branch",
			commitID: "e9776e07fe959ec1d28009bebc978eb5c3685874",
			expected: []string{},
		},
		{
			name:     "merge commit of the single commit branch 2db8f134af73020f9c33fa514cbb680ee1b20ee9",
			commitID: "2db8f134af73020f9c33fa514cbb680ee1b20ee9",
			expected: []string{
				"content/ja/docs/tasks/run-application/run-single-instance-stateful-application.md",
				"content/ja/examples/application/mysql/mysql-deployment.yaml",
				"content/ja/examples/application/wordpress/mysql-deployment.yaml",
			},
		},
		{
			name:     "merge commit of multi commit branch dfd3e59aa2ba837b587b06a9172a614e1b3b4160",
			commitID: "dfd3e59aa2ba837b587b06a9172a614e1b3b4160",
			expected: []string{
				"content/hi/examples/service/networking/namespaced-params.yaml",
				"content/hi/examples/service/networking/network-policy-allow-all-egress.yaml",
				"content/hi/examples/service/networking/network-policy-allow-all-ingress.yaml",
				"content/hi/examples/service/networking/network-policy-default-deny-all.yaml",
				"content/hi/examples/service/networking/network-policy-default-deny-egress.yaml",
				"content/hi/examples/service/networking/network-policy-default-deny-ingress.yaml",
				"content/hi/examples/service/networking/networkpolicy-multiport-egress.yaml",
				"content/hi/examples/service/networking/networkpolicy.yaml",
			},
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			files, err := gitRepoHist.MergeCommitFiles(ctx, tc.commitID)
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

func TestGitHist_FindForkCommit_E2E(t *testing.T) {
	ctx := context.Background()
	testDir := t.TempDir()
	repoDir := filepath.Join(testDir, "repo")
	mustMkDir(t, repoDir)
	cacheDir := filepath.Join(testDir, "cache")
	mustMkDir(t, cacheDir)
	cacheStore := store.NewFileStore(cacheDir)
	gitRepo := git.NewRepo(repoDir)
	gitRepoHist := githist.New(gitRepo, cacheStore)

	if err := gitRepo.Create(ctx, "https://github.com/kubernetes/website"); err != nil {
		t.Fatal(err)
	}

	for _, tc := range []struct {
		name     string
		commitID string
		expected *git.CommitInfo
	}{
		{
			name:     "commitID is on the main branch",
			commitID: "f4cc9ccc61d6faef60e2e162658d7ff02a42435b",
			expected: nil,
		},
		{
			name:     "commitID that is not on the main branch",
			commitID: "26e9bb02e959ffb52b3609f7c794382339b47060",
			expected: &git.CommitInfo{
				CommitID: "5d9e5d4d764f6a9b4ee172e737b54b377724f8f0",
				DateTime: "2025-05-12T23:21:16-07:00",
				Comment:  "Merge pull request #50879 from jayeshmahajan/jm/hi-example-pods-security-sec-alpha",
			},
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			files, err := gitRepoHist.FindForkCommit(ctx, tc.commitID)
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

func TestGitHist_FindMergeCommit_E2E(t *testing.T) {
	ctx := context.Background()
	testDir := t.TempDir()
	repoDir := filepath.Join(testDir, "repo")
	mustMkDir(t, repoDir)
	cacheDir := filepath.Join(testDir, "cache")
	mustMkDir(t, cacheDir)
	cacheStore := store.NewFileStore(cacheDir)
	gitRepo := git.NewRepo(repoDir)
	gitRepoHist := githist.New(gitRepo, cacheStore)

	if err := gitRepo.Create(ctx, "https://github.com/kubernetes/website"); err != nil {
		t.Fatal(err)
	}

	for _, tc := range []struct {
		name     string
		commitID string
		expected *git.CommitInfo
	}{
		{
			name:     "commitID is on the main branch",
			commitID: "f4cc9ccc61d6faef60e2e162658d7ff02a42435b",
			expected: nil,
		},
		{
			name:     "commitID that is not on the main branch",
			commitID: "26e9bb02e959ffb52b3609f7c794382339b47060",
			expected: &git.CommitInfo{
				CommitID: "91c9ff2e4a9f746cc30721485122d8bfb3024b1f",
				DateTime: "2025-05-20T10:47:15-07:00",
				Comment:  "Merge pull request #50889 from yanai-tomohiro/make_ja_job",
			},
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			files, err := gitRepoHist.FindMergeCommit(ctx, tc.commitID)
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

func TestGitHist_GetLastMainBranchCommit_E2E(t *testing.T) {
	ctx := context.Background()
	testDir := t.TempDir()
	repoDir := filepath.Join(testDir, "repo")
	mustMkDir(t, repoDir)
	cacheDir := filepath.Join(testDir, "cache")
	mustMkDir(t, cacheDir)
	cacheStore := store.NewFileStore(cacheDir)
	gitRepo := git.NewRepo(repoDir)
	gitRepoHist := githist.New(gitRepo, cacheStore)

	if err := gitRepo.Create(ctx, "https://github.com/kubernetes/website"); err != nil {
		t.Fatal(err)
	}

	commit, err := gitRepoHist.GetLastMainBranchCommit(ctx)
	if err != nil {
		t.Fatal(err)
	}

	t.Logf("%+v", commit)
}

func mustMkDir(t testing.TB, path string) {
	t.Helper()

	if err := os.Mkdir(path, 0o700); err != nil {
		t.Fatal(err)
	}
}
