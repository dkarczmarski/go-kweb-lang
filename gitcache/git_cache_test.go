package gitcache_test

import (
	"go-kweb-lang/git"
	"go-kweb-lang/gitcache"
	"go-kweb-lang/gitcache/internal"
	"go-kweb-lang/mocks"
	"go.uber.org/mock/gomock"
	"os"
	"path/filepath"
	"reflect"
	"testing"
)

func TestGitRepoCache_FindFileLastCommit(t *testing.T) {
	path := "/path1"
	expectedCommit := git.CommitInfo{CommitId: "ID1", DateTime: "DT1", Comment: "TEXT1"}

	for _, tc := range []struct {
		name     string
		initMock func(repo *mocks.MockRepo)
		before   func(t *testing.T, cachePath string)
		after    func(t *testing.T, cachePath string)
	}{
		{
			name: "miss cache",
			initMock: func(m *mocks.MockRepo) {
				m.EXPECT().
					FindFileLastCommit(path).
					Return(expectedCommit, nil).
					Times(1)
			},
			before: func(t *testing.T, cachePath string) {
				if internal.FileExists(cachePath) {
					t.Fatal("should be impossible")
				}
			},
			after: func(t *testing.T, cachePath string) {
				if !internal.FileExists(cachePath) {
					t.Errorf("cache file %s should exist", cachePath)
				}
			},
		},
		{
			name: "hit cache",
			initMock: func(m *mocks.MockRepo) {
			},
			before: func(t *testing.T, cachePath string) {
				if err := internal.EnsureDir(filepath.Dir(cachePath)); err != nil {
					t.Fatal(err)
				}
				if err := internal.WriteJsonToFile(cachePath, &expectedCommit); err != nil {
					t.Fatal(err)
				}
			},
			after: func(t *testing.T, cachePath string) {
			},
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			m := mocks.NewMockRepo(ctrl)

			tc.initMock(m)

			cacheDir, err := os.MkdirTemp("", "testdir_")
			if err != nil {
				t.Fatal(err)
			}
			defer os.RemoveAll(cacheDir)

			gc := gitcache.New(m, cacheDir)
			cachePath := filepath.Join(internal.FileLastCommitDir(cacheDir), internal.KeyFile(internal.KeyHash(path)))

			tc.before(t, cachePath)

			commit, err := gc.FindFileLastCommit(path)
			if err != nil {
				t.Errorf("unexpected error: %v", err)
			}
			if !reflect.DeepEqual(commit, expectedCommit) {
				t.Errorf("unexpected outcome\nactual   : %+v\nexpected: %+v", commit, expectedCommit)
			}

			tc.after(t, cachePath)
		})
	}
}

func TestGitRepoCache_FindFileCommitsAfter(t *testing.T) {
	path := "path1"
	commitId := "ID"
	expectedCommits := []git.CommitInfo{
		{CommitId: "ID1", DateTime: "DT1", Comment: "TEXT1"},
		{CommitId: "ID2", DateTime: "DT2", Comment: "TEXT2"},
	}

	for _, tc := range []struct {
		name     string
		initMock func(repo *mocks.MockRepo)
		before   func(t *testing.T, cachePath string)
		after    func(t *testing.T, cachePath string)
	}{
		{
			name: "miss cache",
			initMock: func(m *mocks.MockRepo) {
				m.EXPECT().
					FindFileCommitsAfter(path, commitId).
					Return(expectedCommits, nil).
					Times(1)
			},
			before: func(t *testing.T, cachePath string) {
				if internal.FileExists(cachePath) {
					t.Fatal("should be impossible")
				}
			},
			after: func(t *testing.T, cachePath string) {
				if !internal.FileExists(cachePath) {
					t.Errorf("cache file %s should exist", cachePath)
				}
			},
		},
		{
			name: "hit cache",
			initMock: func(m *mocks.MockRepo) {
			},
			before: func(t *testing.T, cachePath string) {
				if err := internal.EnsureDir(filepath.Dir(cachePath)); err != nil {
					t.Fatal(err)
				}
				if err := internal.WriteJsonToFile(cachePath, &expectedCommits); err != nil {
					t.Fatal(err)
				}
			},
			after: func(t *testing.T, cachePath string) {
			},
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			m := mocks.NewMockRepo(ctrl)

			tc.initMock(m)

			cacheDir, err := os.MkdirTemp("", "testdir_")
			if err != nil {
				t.Fatal(err)
			}
			defer os.RemoveAll(cacheDir)

			gc := gitcache.New(m, cacheDir)
			cachePath := filepath.Join(internal.FileUpdatesDir(cacheDir), internal.KeyFile(internal.KeyHash(path)))

			tc.before(t, cachePath)

			commits, err := gc.FindFileCommitsAfter(path, commitId)
			if err != nil {
				t.Errorf("unexpected error: %v", err)
			}
			if !reflect.DeepEqual(commits, expectedCommits) {
				t.Errorf("unexpected outcome\nactual   : %+v\nexpected: %+v", commits, expectedCommits)
			}

			tc.after(t, cachePath)
		})
	}
}

func TestGitRepoCache_FindMergePoints(t *testing.T) {
	commitId := "ID"
	expectedCommits := []git.CommitInfo{
		{CommitId: "ID1", DateTime: "DT1", Comment: "TEXT1"},
		{CommitId: "ID2", DateTime: "DT2", Comment: "TEXT2"},
	}

	for _, tc := range []struct {
		name     string
		initMock func(repo *mocks.MockRepo)
		before   func(t *testing.T, cachePath string)
		after    func(t *testing.T, cachePath string)
	}{
		{
			name: "miss cache",
			initMock: func(m *mocks.MockRepo) {
				m.EXPECT().
					FindMergePoints(commitId).
					Return(expectedCommits, nil).
					Times(1)
			},
			before: func(t *testing.T, cachePath string) {
				if internal.FileExists(cachePath) {
					t.Fatal("should be impossible")
				}
			},
			after: func(t *testing.T, cachePath string) {
				if !internal.FileExists(cachePath) {
					t.Errorf("cache file %s should exist", cachePath)
				}
			},
		},
		{
			name: "hit cache",
			initMock: func(m *mocks.MockRepo) {
			},
			before: func(t *testing.T, cachePath string) {
				if err := internal.EnsureDir(filepath.Dir(cachePath)); err != nil {
					t.Fatal(err)
				}
				if err := internal.WriteJsonToFile(cachePath, &expectedCommits); err != nil {
					t.Fatal(err)
				}
			},
			after: func(t *testing.T, cachePath string) {
			},
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			m := mocks.NewMockRepo(ctrl)

			tc.initMock(m)

			cacheDir, err := os.MkdirTemp("", "testdir_")
			if err != nil {
				t.Fatal(err)
			}
			defer os.RemoveAll(cacheDir)

			gc := gitcache.New(m, cacheDir)
			cachePath := filepath.Join(internal.MergePointsDir(cacheDir), internal.KeyFile(internal.KeyHash(commitId)))

			tc.before(t, cachePath)

			commits, err := gc.FindMergePoints(commitId)
			if err != nil {
				t.Errorf("unexpected error: %v", err)
			}
			if !reflect.DeepEqual(commits, expectedCommits) {
				t.Errorf("unexpected outcome\nactual   : %+v\nexpected: %+v", commits, expectedCommits)
			}

			tc.after(t, cachePath)
		})
	}
}

func TestGitRepoCache_InvalidatePath(t *testing.T) {
	path := "dir1/path1"
	for _, tc := range []struct {
		name   string
		before func(t *testing.T, prop map[string]any)
		after  func(t *testing.T, prop map[string]any)
	}{
		{
			name: "when cache file does not exist, it returns no error",
		},
		{
			name: "when 'file-last-commit' cache file exists, it removes it",
			before: func(t *testing.T, prop map[string]any) {
				cacheDir := prop["cacheDir"].(string)

				cacheFile := filepath.Join(internal.FileLastCommitDir(cacheDir), internal.KeyFile(internal.KeyHash(path)))
				prop["cacheFile"] = cacheFile

				if err := internal.EnsureDir(filepath.Dir(cacheFile)); err != nil {
					t.Fatal(err)
				}
				if err := internal.WriteJsonToFile(cacheFile, struct{}{}); err != nil {
					t.Fatal(err)
				}
			},
			after: func(t *testing.T, prop map[string]any) {
				cacheFile := prop["cacheFile"].(string)

				if internal.FileExists(cacheFile) {
					t.Errorf("cache file %s should not exist", cacheFile)
				}
			},
		},
		{
			name: "when 'file-updates' cache file exists, it removes it",
			before: func(t *testing.T, prop map[string]any) {
				cacheDir := prop["cacheDir"].(string)

				cacheFile := filepath.Join(internal.FileUpdatesDir(cacheDir), internal.KeyFile(internal.KeyHash(path)))
				prop["cacheFile"] = cacheFile

				if err := internal.EnsureDir(filepath.Dir(cacheFile)); err != nil {
					t.Fatal(err)
				}
				if err := internal.WriteJsonToFile(cacheFile, struct{}{}); err != nil {
					t.Fatal(err)
				}
			},
			after: func(t *testing.T, prop map[string]any) {
				cacheFile := prop["cacheFile"].(string)

				if internal.FileExists(cacheFile) {
					t.Errorf("cache file %s should not exist", cacheFile)
				}
			},
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			m := mocks.NewMockRepo(ctrl)

			prop := make(map[string]any)

			cacheDir, err := os.MkdirTemp("", "testdir_")
			if err != nil {
				t.Fatal(err)
			}
			defer os.RemoveAll(cacheDir)
			prop["cacheDir"] = cacheDir

			gc := gitcache.New(m, cacheDir)

			if tc.before != nil {
				tc.before(t, prop)
			}

			if err := gc.InvalidatePath(path); err != nil {
				t.Fatal(err)
			}

			if tc.after != nil {
				tc.after(t, prop)
			}
		})
	}
}
