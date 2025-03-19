package gitcache_test

import (
	"context"
	"go-kweb-lang/git"
	"go-kweb-lang/gitcache"
	"go-kweb-lang/gitcache/internal"
	"go-kweb-lang/mocks"
	"path/filepath"
	"reflect"
	"testing"

	"go.uber.org/mock/gomock"
)

func TestGitRepoCache_FindFileLastCommit(t *testing.T) {
	path := "/path1"
	expectedCommit := git.CommitInfo{CommitID: "ID1", DateTime: "DT1", Comment: "TEXT1"}

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
					FindFileLastCommit(gomock.Any(), path).
					Return(expectedCommit, nil).
					Times(1)
			},
			before: func(t *testing.T, cachePath string) {
				t.Helper()
				if internal.FileExists(cachePath) {
					t.Fatal("should be impossible")
				}
			},
			after: func(t *testing.T, cachePath string) {
				t.Helper()
				if !internal.FileExists(cachePath) {
					t.Errorf("cache file %s should exist", cachePath)
				}
			},
		},
		{
			name: "hit cache",
			before: func(t *testing.T, cachePath string) {
				t.Helper()
				if err := internal.EnsureDir(filepath.Dir(cachePath)); err != nil {
					t.Fatal(err)
				}
				if err := internal.WriteJSONToFile(cachePath, &expectedCommit); err != nil {
					t.Fatal(err)
				}
			},
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			ctrl := gomock.NewController(t)
			mock := mocks.NewMockRepo(ctrl)

			if tc.initMock != nil {
				tc.initMock(mock)
			}

			cacheDir := t.TempDir()
			gc := gitcache.New(mock, cacheDir)
			cachePath := filepath.Join(internal.FileLastCommitDir(cacheDir), internal.KeyFile(internal.KeyHash(path)))

			tc.before(t, cachePath)

			commit, err := gc.FindFileLastCommit(context.Background(), path)
			if err != nil {
				t.Errorf("unexpected error: %v", err)
			}

			if !reflect.DeepEqual(commit, expectedCommit) {
				t.Errorf("unexpected outcome\nactual   : %+v\nexpected: %+v", commit, expectedCommit)
			}

			if tc.after != nil {
				tc.after(t, cachePath)
			}
		})
	}
}

func TestGitRepoCache_FindFileCommitsAfter(t *testing.T) {
	path := "path1"
	commitID := "ID"
	expectedCommits := []git.CommitInfo{
		{CommitID: "ID1", DateTime: "DT1", Comment: "TEXT1"},
		{CommitID: "ID2", DateTime: "DT2", Comment: "TEXT2"},
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
					FindFileCommitsAfter(context.Background(), path, commitID).
					Return(expectedCommits, nil).
					Times(1)
			},
			before: func(t *testing.T, cachePath string) {
				t.Helper()
				if internal.FileExists(cachePath) {
					t.Fatal("should be impossible")
				}
			},
			after: func(t *testing.T, cachePath string) {
				t.Helper()
				if !internal.FileExists(cachePath) {
					t.Errorf("cache file %s should exist", cachePath)
				}
			},
		},
		{
			name: "hit cache",
			before: func(t *testing.T, cachePath string) {
				t.Helper()
				if err := internal.EnsureDir(filepath.Dir(cachePath)); err != nil {
					t.Fatal(err)
				}
				if err := internal.WriteJSONToFile(cachePath, &expectedCommits); err != nil {
					t.Fatal(err)
				}
			},
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			ctrl := gomock.NewController(t)
			mock := mocks.NewMockRepo(ctrl)

			if tc.initMock != nil {
				tc.initMock(mock)
			}

			cacheDir := t.TempDir()

			gc := gitcache.New(mock, cacheDir)
			cachePath := filepath.Join(internal.FileUpdatesDir(cacheDir), internal.KeyFile(internal.KeyHash(path)))

			tc.before(t, cachePath)

			commits, err := gc.FindFileCommitsAfter(context.Background(), path, commitID)
			if err != nil {
				t.Errorf("unexpected error: %v", err)
			}
			if !reflect.DeepEqual(commits, expectedCommits) {
				t.Errorf("unexpected outcome\nactual   : %+v\nexpected: %+v", commits, expectedCommits)
			}

			if tc.after != nil {
				tc.after(t, cachePath)
			}
		})
	}
}

func TestGitRepoCache_FindMergePoints(t *testing.T) {
	commitID := "ID"
	expectedCommits := []git.CommitInfo{
		{CommitID: "ID1", DateTime: "DT1", Comment: "TEXT1"},
		{CommitID: "ID2", DateTime: "DT2", Comment: "TEXT2"},
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
					FindMergePoints(context.Background(), commitID).
					Return(expectedCommits, nil).
					Times(1)
			},
			before: func(t *testing.T, cachePath string) {
				t.Helper()
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
			before: func(t *testing.T, cachePath string) {
				t.Helper()
				if err := internal.EnsureDir(filepath.Dir(cachePath)); err != nil {
					t.Fatal(err)
				}
				if err := internal.WriteJSONToFile(cachePath, &expectedCommits); err != nil {
					t.Fatal(err)
				}
			},
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			ctrl := gomock.NewController(t)
			mock := mocks.NewMockRepo(ctrl)

			if tc.initMock != nil {
				tc.initMock(mock)
			}

			cacheDir := t.TempDir()
			gc := gitcache.New(mock, cacheDir)
			cachePath := filepath.Join(internal.MergePointsDir(cacheDir), internal.KeyFile(internal.KeyHash(commitID)))

			tc.before(t, cachePath)

			commits, err := gc.FindMergePoints(context.Background(), commitID)
			if err != nil {
				t.Errorf("unexpected error: %v", err)
			}
			if !reflect.DeepEqual(commits, expectedCommits) {
				t.Errorf("unexpected outcome\nactual   : %+v\nexpected: %+v", commits, expectedCommits)
			}

			if tc.after != nil {
				tc.after(t, cachePath)
			}
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
				t.Helper()

				cacheDir := prop["cacheDir"].(string)

				cacheFile := filepath.Join(internal.FileLastCommitDir(cacheDir), internal.KeyFile(internal.KeyHash(path)))
				prop["cacheFile"] = cacheFile

				if err := internal.EnsureDir(filepath.Dir(cacheFile)); err != nil {
					t.Fatal(err)
				}
				if err := internal.WriteJSONToFile(cacheFile, struct{}{}); err != nil {
					t.Fatal(err)
				}
			},
			after: func(t *testing.T, prop map[string]any) {
				t.Helper()

				cacheFile := prop["cacheFile"].(string)

				if internal.FileExists(cacheFile) {
					t.Errorf("cache file %s should not exist", cacheFile)
				}
			},
		},
		{
			name: "when 'file-updates' cache file exists, it removes it",
			before: func(t *testing.T, prop map[string]any) {
				t.Helper()

				cacheDir := prop["cacheDir"].(string)

				cacheFile := filepath.Join(internal.FileUpdatesDir(cacheDir), internal.KeyFile(internal.KeyHash(path)))
				prop["cacheFile"] = cacheFile

				if err := internal.EnsureDir(filepath.Dir(cacheFile)); err != nil {
					t.Fatal(err)
				}
				if err := internal.WriteJSONToFile(cacheFile, struct{}{}); err != nil {
					t.Fatal(err)
				}
			},
			after: func(t *testing.T, prop map[string]any) {
				t.Helper()

				cacheFile := prop["cacheFile"].(string)

				if internal.FileExists(cacheFile) {
					t.Errorf("cache file %s should not exist", cacheFile)
				}
			},
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			ctrl := gomock.NewController(t)
			mock := mocks.NewMockRepo(ctrl)

			prop := make(map[string]any)

			cacheDir := t.TempDir()
			prop["cacheDir"] = cacheDir

			gitCache := gitcache.New(mock, cacheDir)

			if tc.before != nil {
				tc.before(t, prop)
			}

			if err := gitCache.InvalidatePath(path); err != nil {
				t.Fatal(err)
			}

			if tc.after != nil {
				tc.after(t, prop)
			}
		})
	}
}
