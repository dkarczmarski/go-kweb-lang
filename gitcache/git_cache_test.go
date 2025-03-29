package gitcache_test

import (
	"context"
	"go-kweb-lang/git"
	"go-kweb-lang/gitcache"
	"go-kweb-lang/mocks"
	"go-kweb-lang/proxycache"
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
		before   func(t *testing.T, cacheDir, category, key string)
		after    func(t *testing.T, cacheDir, category, key string)
	}{
		{
			name: "miss cache",
			initMock: func(m *mocks.MockRepo) {
				m.EXPECT().
					FindFileLastCommit(gomock.Any(), path).
					Return(expectedCommit, nil).
					Times(1)
			},
			before: func(t *testing.T, cacheDir, category, key string) {
				t.Helper()
				if proxycacheKeyExists(t, cacheDir, category, key) {
					t.Fatal("should be impossible")
				}
			},
			after: func(t *testing.T, cacheDir, category, key string) {
				t.Helper()
				if !proxycacheKeyExists(t, cacheDir, category, key) {
					t.Errorf("cache key %s should exist", key)
				}
			},
		},
		{
			name: "hit cache",
			before: func(t *testing.T, cacheDir, category, key string) {
				t.Helper()

				if err := proxycache.Put(cacheDir, category, key, expectedCommit); err != nil {
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

			category := gitcache.CategoryLastCommit
			key := path

			tc.before(t, cacheDir, category, key)

			commit, err := gc.FindFileLastCommit(context.Background(), path)
			if err != nil {
				t.Errorf("unexpected error: %v", err)
			}

			if !reflect.DeepEqual(commit, expectedCommit) {
				t.Errorf("unexpected outcome\nactual   : %+v\nexpected: %+v", commit, expectedCommit)
			}

			if tc.after != nil {
				tc.after(t, cacheDir, category, key)
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
		before   func(t *testing.T, cacheDir, category, key string)
		after    func(t *testing.T, cacheDir, category, key string)
	}{
		{
			name: "miss cache",
			initMock: func(m *mocks.MockRepo) {
				m.EXPECT().
					FindFileCommitsAfter(context.Background(), path, commitID).
					Return(expectedCommits, nil).
					Times(1)
			},
			before: func(t *testing.T, cacheDir, category, key string) {
				t.Helper()
				if proxycacheKeyExists(t, cacheDir, category, key) {
					t.Fatal("should be impossible")
				}
			},
			after: func(t *testing.T, cacheDir, category, key string) {
				t.Helper()
				if !proxycacheKeyExists(t, cacheDir, category, key) {
					t.Errorf("cache key %s should exist", key)
				}
			},
		},
		{
			name: "hit cache",
			before: func(t *testing.T, cacheDir, category, key string) {
				t.Helper()

				if err := proxycache.Put(cacheDir, category, key, expectedCommits); err != nil {
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

			category := gitcache.CategoryUpdates
			key := path

			tc.before(t, cacheDir, category, key)

			commits, err := gc.FindFileCommitsAfter(context.Background(), path, commitID)
			if err != nil {
				t.Errorf("unexpected error: %v", err)
			}
			if !reflect.DeepEqual(commits, expectedCommits) {
				t.Errorf("unexpected outcome\nactual   : %+v\nexpected: %+v", commits, expectedCommits)
			}

			if tc.after != nil {
				tc.after(t, cacheDir, category, key)
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
		before   func(t *testing.T, cacheDir, category, key string)
		after    func(t *testing.T, cacheDir, category, key string)
	}{
		{
			name: "miss cache",
			initMock: func(m *mocks.MockRepo) {
				m.EXPECT().
					FindMergePoints(context.Background(), commitID).
					Return(expectedCommits, nil).
					Times(1)
			},
			before: func(t *testing.T, cacheDir, category, key string) {
				t.Helper()
				if proxycacheKeyExists(t, cacheDir, category, key) {
					t.Fatal("should be impossible")
				}
			},
			after: func(t *testing.T, cacheDir, category, key string) {
				if !proxycacheKeyExists(t, cacheDir, category, key) {
					t.Errorf("cache key %s should exist", key)
				}
			},
		},
		{
			name: "hit cache",
			before: func(t *testing.T, cacheDir, category, key string) {
				t.Helper()

				if err := proxycache.Put(cacheDir, category, key, expectedCommits); err != nil {
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

			category := gitcache.CategoryMergePoints
			key := commitID

			tc.before(t, cacheDir, category, key)

			commits, err := gc.FindMergePoints(context.Background(), commitID)
			if err != nil {
				t.Errorf("unexpected error: %v", err)
			}
			if !reflect.DeepEqual(commits, expectedCommits) {
				t.Errorf("unexpected outcome\nactual   : %+v\nexpected: %+v", commits, expectedCommits)
			}

			if tc.after != nil {
				tc.after(t, cacheDir, category, key)
			}
		})
	}
}

func TestGitRepoCache_InvalidatePath(t *testing.T) {
	for _, tc := range []struct {
		name   string
		before func(t *testing.T, cacheDir, key string)
		after  func(t *testing.T, cacheDir, key string)
	}{
		{
			name: "when cache file does not exist, it returns no error",
		},
		{
			name: "when 'file-last-commit' cache file exists, it removes it",
			before: func(t *testing.T, cacheDir, key string) {
				t.Helper()

				if err := proxycache.Put(cacheDir, gitcache.CategoryLastCommit, key, struct{}{}); err != nil {
					t.Fatal(err)
				}
			},
			after: func(t *testing.T, cacheDir, key string) {
				t.Helper()

				if proxycacheKeyExists(t, cacheDir, "", key) {
					t.Error("cache key should not exist")
				}
			},
		},
		{
			name: "when 'file-updates' cache file exists, it removes it",
			before: func(t *testing.T, cacheDir, key string) {
				t.Helper()

				if err := proxycache.Put(cacheDir, gitcache.CategoryUpdates, key, struct{}{}); err != nil {
					t.Fatal(err)
				}
			},
			after: func(t *testing.T, cacheDir, key string) {
				t.Helper()

				if proxycacheKeyExists(t, cacheDir, "", key) {
					t.Error("cache key should not exist")
				}
			},
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			ctrl := gomock.NewController(t)
			mock := mocks.NewMockRepo(ctrl)

			cacheDir := t.TempDir()

			gitCache := gitcache.New(mock, cacheDir)

			samplePath := "dir1/path1"

			if tc.before != nil {
				tc.before(t, cacheDir, samplePath)
			}

			if err := gitCache.InvalidatePath(samplePath); err != nil {
				t.Fatal(err)
			}

			if tc.after != nil {
				tc.after(t, cacheDir, samplePath)
			}
		})
	}
}

func proxycacheKeyExists(t *testing.T, cacheDir, category, key string) bool {
	t.Helper()

	exists, err := proxycache.KeyExists(cacheDir, category, key)
	if err != nil {
		t.Fatal(err)
	}

	return exists
}
