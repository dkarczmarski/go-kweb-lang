package gitpc_test

import (
	"context"
	"reflect"
	"testing"

	"go-kweb-lang/git"
	"go-kweb-lang/gitpc"
	"go-kweb-lang/mocks"
	"go-kweb-lang/proxycache"

	"go.uber.org/mock/gomock"
)

func TestProxyCache_FindFileLastCommit(t *testing.T) {
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

				if mustProxycacheKeyExists(t, cacheDir, category, key) {
					t.Fatal("should be impossible")
				}
			},
			after: func(t *testing.T, cacheDir, category, key string) {
				t.Helper()

				if !mustProxycacheKeyExists(t, cacheDir, category, key) {
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
			gc := gitpc.New(mock, cacheDir)

			category := gitpc.CategoryLastCommit
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

func TestProxyCache_FindFileCommitsAfter(t *testing.T) {
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

				if mustProxycacheKeyExists(t, cacheDir, category, key) {
					t.Fatal("should be impossible")
				}
			},
			after: func(t *testing.T, cacheDir, category, key string) {
				t.Helper()

				if !mustProxycacheKeyExists(t, cacheDir, category, key) {
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

			gc := gitpc.New(mock, cacheDir)

			category := gitpc.CategoryUpdates
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

func TestProxyCache_ListMergePoints(t *testing.T) {
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
					ListMergePoints(context.Background(), commitID).
					Return(expectedCommits, nil).
					Times(1)
			},
			before: func(t *testing.T, cacheDir, category, key string) {
				t.Helper()

				if mustProxycacheKeyExists(t, cacheDir, category, key) {
					t.Fatal("should be impossible")
				}
			},
			after: func(t *testing.T, cacheDir, category, key string) {
				t.Helper()

				if !mustProxycacheKeyExists(t, cacheDir, category, key) {
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
			gc := gitpc.New(mock, cacheDir)

			category := gitpc.CategoryMergePoints
			key := commitID

			tc.before(t, cacheDir, category, key)

			commits, err := gc.ListMergePoints(context.Background(), commitID)
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

func TestProxyCache_InvalidatePath(t *testing.T) {
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

				if err := proxycache.Put(cacheDir, gitpc.CategoryLastCommit, key, struct{}{}); err != nil {
					t.Fatal(err)
				}
			},
			after: func(t *testing.T, cacheDir, key string) {
				t.Helper()

				if mustProxycacheKeyExists(t, cacheDir, "", key) {
					t.Error("cache key should not exist")
				}
			},
		},
		{
			name: "when 'file-updates' cache file exists, it removes it",
			before: func(t *testing.T, cacheDir, key string) {
				t.Helper()

				if err := proxycache.Put(cacheDir, gitpc.CategoryUpdates, key, struct{}{}); err != nil {
					t.Fatal(err)
				}
			},
			after: func(t *testing.T, cacheDir, key string) {
				t.Helper()

				if mustProxycacheKeyExists(t, cacheDir, "", key) {
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

			gitCache := gitpc.New(mock, cacheDir)

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

func TestProxyCache_PullRefresh(t *testing.T) {
	for _, tc := range []struct {
		name     string
		initMock func(t *testing.T, mock *mocks.MockRepo, cacheDir string, ctx context.Context) []string
	}{
		{
			name: "no fresh commits",
			initMock: func(t *testing.T, mock *mocks.MockRepo, cacheDir string, ctx context.Context) []string {
				t.Helper()

				mock.EXPECT().ListFreshCommits(ctx).Return(nil, nil)

				return []string{}
			},
		},
		{
			name: "two fresh commits",
			initMock: func(t *testing.T, mock *mocks.MockRepo, cacheDir string, ctx context.Context) []string {
				t.Helper()

				mock.EXPECT().ListFreshCommits(ctx).Return(
					[]git.CommitInfo{
						{
							CommitID: "CID1",
							DateTime: "D1",
							Comment:  "Comment1",
						},
						{
							CommitID: "CID2",
							DateTime: "D2",
							Comment:  "Comment2",
						},
					}, nil,
				)
				mock.EXPECT().ListFilesInCommit(ctx, "CID1").Return([]string{"F10", "F11"}, nil)
				mock.EXPECT().ListFilesInCommit(ctx, "CID2").Return([]string{"F11", "F12"}, nil)

				mustProxycachePut(t, cacheDir, gitpc.CategoryLastCommit, "F10")
				mustProxycachePut(t, cacheDir, gitpc.CategoryLastCommit, "F11")
				mustProxycachePut(t, cacheDir, gitpc.CategoryUpdates, "F11")

				mustProxycachePut(t, cacheDir, gitpc.CategoryMainBranchCommits, "")

				return []string{"F10", "F11", "F12"}
			},
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			ctx := context.Background()

			ctrl := gomock.NewController(t)
			mock := mocks.NewMockRepo(ctrl)

			cacheDir := t.TempDir()

			mock.EXPECT().Fetch(ctx).Return(nil)
			fetchedFiles := tc.initMock(t, mock, cacheDir, ctx)
			mock.EXPECT().Pull(ctx).Return(nil)

			gitCache := gitpc.New(mock, cacheDir)

			if err := gitCache.PullRefresh(ctx); err != nil {
				t.Error(err)
			}

			for _, file := range fetchedFiles {
				for _, category := range []string{gitpc.CategoryLastCommit, gitpc.CategoryUpdates} {
					exists, err := proxycache.KeyExists(cacheDir, category, file)
					if err != nil {
						t.Fatal(err)
					}

					if exists {
						t.Errorf("file %v should not exists", file)
					}
				}
			}

			if mustProxycacheKeyExists(t, cacheDir, gitpc.CategoryMainBranchCommits, "") {
				t.Error("key should not exists")
			}
		})
	}
}

func mustProxycachePut(t *testing.T, cacheDir, category, key string) {
	t.Helper()

	if err := proxycache.Put(cacheDir, category, key, struct{}{}); err != nil {
		t.Fatal(err)
	}
}

func mustProxycacheKeyExists(t *testing.T, cacheDir, category, key string) bool {
	t.Helper()

	exists, err := proxycache.KeyExists(cacheDir, category, key)
	if err != nil {
		t.Fatal(err)
	}

	return exists
}
