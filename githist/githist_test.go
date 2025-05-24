package githist_test

import (
	"context"
	"reflect"
	"testing"

	"go-kweb-lang/git"
	"go-kweb-lang/githist"
	"go-kweb-lang/mocks"
	"go-kweb-lang/proxycache"

	"go.uber.org/mock/gomock"
)

func TestGitHist_FindFileLastCommit(t *testing.T) {
	ctx := context.Background()

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
					FindFileLastCommit(ctx, path).
					Return(expectedCommit, nil).
					Times(1)
			},
			before: func(t *testing.T, cacheDir, category, key string) {
				t.Helper()

				if mustProxyCacheKeyExists(t, cacheDir, category, key) {
					t.Fatal("should be impossible")
				}
			},
			after: func(t *testing.T, cacheDir, category, key string) {
				t.Helper()

				if !mustProxyCacheKeyExists(t, cacheDir, category, key) {
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
			gc := githist.New(mock, cacheDir)

			category := "git-file-last-commit"
			key := path

			tc.before(t, cacheDir, category, key)

			commit, err := gc.FindFileLastCommit(ctx, path)
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

func TestGitHist_FindFileCommitsAfter(t *testing.T) {
	ctx := context.Background()

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
					FindFileCommitsAfter(ctx, path, commitID).
					Return(expectedCommits, nil).
					Times(1)
			},
			before: func(t *testing.T, cacheDir, category, key string) {
				t.Helper()

				if mustProxyCacheKeyExists(t, cacheDir, category, key) {
					t.Fatal("should be impossible")
				}
			},
			after: func(t *testing.T, cacheDir, category, key string) {
				t.Helper()

				if !mustProxyCacheKeyExists(t, cacheDir, category, key) {
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

			gc := githist.New(mock, cacheDir)

			category := "git-file-updates"
			key := path

			tc.before(t, cacheDir, category, key)

			commits, err := gc.FindFileCommitsAfter(ctx, path, commitID)
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

func TestGitHist_FindForkCommit(t *testing.T) {
	ctx := context.Background()

	mainBranchCommits := []git.CommitInfo{
		{CommitID: "ID40", DateTime: "DT40", Comment: "TEXT40"},
		{CommitID: "ID30", DateTime: "DT30", Comment: "TEXT30"},
		{CommitID: "ID20", DateTime: "DT20", Comment: "TEXT20"},
		{CommitID: "ID10", DateTime: "DT10", Comment: "TEXT10"},
	}

	for _, tc := range []struct {
		name     string
		commitID string
		expected *git.CommitInfo
		initMock func(repo *mocks.MockRepo, commitID string)
		before   func(t *testing.T, cacheDir, category, key string, expected *git.CommitInfo)
		after    func(t *testing.T, cacheDir, category, key string)
	}{
		{
			name:     "miss cache",
			commitID: "ID25",
			expected: &git.CommitInfo{CommitID: "ID20", DateTime: "DT20", Comment: "TEXT20"},
			initMock: func(m *mocks.MockRepo, commitID string) {
				m.EXPECT().ListMainBranchCommits(ctx).
					Return(mainBranchCommits, nil).
					Times(1)

				m.EXPECT().
					ListAncestorCommits(ctx, commitID).
					Return([]git.CommitInfo{
						{CommitID: "ID20", DateTime: "DT20", Comment: "TEXT20"},
						{CommitID: "ID10", DateTime: "DT10", Comment: "TEXT10"},
					}, nil).
					Times(1)
			},
			before: func(t *testing.T, cacheDir, category, key string, _ *git.CommitInfo) {
				t.Helper()

				if mustProxyCacheKeyExists(t, cacheDir, category, key) {
					t.Fatal("should be impossible")
				}
			},
			after: func(t *testing.T, cacheDir, category, key string) {
				t.Helper()

				if !mustProxyCacheKeyExists(t, cacheDir, category, key) {
					t.Errorf("cache key %s should exist", key)
				}
			},
		},
		{
			name:     "hit cache",
			commitID: "ID25",
			expected: &git.CommitInfo{CommitID: "ID20", DateTime: "DT20", Comment: "TEXT20"},
			before: func(t *testing.T, cacheDir, category, key string, expected *git.CommitInfo) {
				t.Helper()

				if err := proxycache.Put(cacheDir, category, key, expected); err != nil {
					t.Fatal(err)
				}
			},
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			ctrl := gomock.NewController(t)
			gitRepoMock := mocks.NewMockRepo(ctrl)

			if tc.initMock != nil {
				tc.initMock(gitRepoMock, tc.commitID)
			}

			cacheDir := t.TempDir()
			gc := githist.New(gitRepoMock, cacheDir)

			category := "git-fork-commit"
			key := tc.commitID

			tc.before(t, cacheDir, category, key, tc.expected)

			mergeCommit, err := gc.FindForkCommit(ctx, tc.commitID)
			if err != nil {
				t.Errorf("unexpected error: %v", err)
			}

			if !reflect.DeepEqual(mergeCommit, tc.expected) {
				t.Errorf("unexpected outcome\nactual   : %+v\nexpected: %+v", mergeCommit, tc.expected)
			}

			if tc.after != nil {
				tc.after(t, cacheDir, category, key)
			}
		})
	}
}

func TestGitHist_FindMergeCommit(t *testing.T) {
	ctx := context.Background()

	mainBranchCommits := []git.CommitInfo{
		{CommitID: "ID40", DateTime: "DT40", Comment: "TEXT40"},
		{CommitID: "ID30", DateTime: "DT30", Comment: "TEXT30"},
		{CommitID: "ID20", DateTime: "DT20", Comment: "TEXT20"},
		{CommitID: "ID10", DateTime: "DT10", Comment: "TEXT10"},
	}

	for _, tc := range []struct {
		name     string
		commitID string
		expected *git.CommitInfo
		initMock func(repo *mocks.MockRepo, commitID string)
		before   func(t *testing.T, cacheDir, category, key string, expected *git.CommitInfo)
		after    func(t *testing.T, cacheDir, category, key string)
	}{
		{
			name:     "miss cache",
			commitID: "ID25",
			expected: &git.CommitInfo{CommitID: "ID30", DateTime: "DT30", Comment: "TEXT30"},
			initMock: func(m *mocks.MockRepo, commitID string) {
				m.EXPECT().ListMainBranchCommits(ctx).
					Return(mainBranchCommits, nil).
					Times(1)

				m.EXPECT().
					ListMergePoints(ctx, commitID).
					Return([]git.CommitInfo{
						{CommitID: "ID30", DateTime: "DT30", Comment: "TEXT30"},
						{CommitID: "ID40", DateTime: "DT40", Comment: "TEXT40"},
					}, nil).
					Times(1)
			},
			before: func(t *testing.T, cacheDir, category, key string, _ *git.CommitInfo) {
				t.Helper()

				if mustProxyCacheKeyExists(t, cacheDir, category, key) {
					t.Fatal("should be impossible")
				}
			},
			after: func(t *testing.T, cacheDir, category, key string) {
				t.Helper()

				if !mustProxyCacheKeyExists(t, cacheDir, category, key) {
					t.Errorf("cache key %s should exist", key)
				}
			},
		},
		{
			name:     "hit cache",
			commitID: "ID25",
			expected: &git.CommitInfo{CommitID: "ID30", DateTime: "DT30", Comment: "TEXT30"},
			before: func(t *testing.T, cacheDir, category, key string, expected *git.CommitInfo) {
				t.Helper()

				if err := proxycache.Put(cacheDir, category, key, expected); err != nil {
					t.Fatal(err)
				}
			},
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			ctrl := gomock.NewController(t)
			gitRepoMock := mocks.NewMockRepo(ctrl)

			if tc.initMock != nil {
				tc.initMock(gitRepoMock, tc.commitID)
			}

			cacheDir := t.TempDir()
			gc := githist.New(gitRepoMock, cacheDir)

			category := "git-merge-commit"
			key := tc.commitID

			tc.before(t, cacheDir, category, key, tc.expected)

			mergeCommit, err := gc.FindMergeCommit(ctx, tc.commitID)
			if err != nil {
				t.Errorf("unexpected error: %v", err)
			}

			if !reflect.DeepEqual(mergeCommit, tc.expected) {
				t.Errorf("unexpected outcome\nactual   : %+v\nexpected: %+v", mergeCommit, tc.expected)
			}

			if tc.after != nil {
				tc.after(t, cacheDir, category, key)
			}
		})
	}
}

func TestGitHist_PullRefresh(t *testing.T) {
	ctx := context.Background()

	const categoryLastCommit = "git-file-last-commit"
	const categoryMainBranchCommits = "git-main-branch-commits"
	const categoryUpdates = "git-file-updates"

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
			name: "two fresh commits on the main branch",
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

				mustProxyCachePut(t, cacheDir, categoryLastCommit, "F10")
				mustProxyCachePut(t, cacheDir, categoryLastCommit, "F11")
				mustProxyCachePut(t, cacheDir, categoryUpdates, "F11")

				mustProxyCachePut(t, cacheDir, categoryMainBranchCommits, "")

				return []string{"F10", "F11", "F12"}
			},
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			mock := mocks.NewMockRepo(ctrl)

			cacheDir := t.TempDir()

			mock.EXPECT().Fetch(ctx).Return(nil)
			fetchedFiles := tc.initMock(t, mock, cacheDir, ctx)
			mock.EXPECT().Pull(ctx).Return(nil)

			gitCache := githist.New(mock, cacheDir)

			if err := gitCache.PullRefresh(ctx); err != nil {
				t.Error(err)
			}

			for _, file := range fetchedFiles {
				for _, category := range []string{categoryLastCommit, categoryUpdates} {
					exists, err := proxycache.KeyExists(cacheDir, category, file)
					if err != nil {
						t.Fatal(err)
					}

					if exists {
						t.Errorf("file %v should not exists", file)
					}
				}
			}

			if mustProxyCacheKeyExists(t, cacheDir, categoryMainBranchCommits, "") {
				t.Error("key should not exists")
			}
		})
	}
}

func mustProxyCachePut(t *testing.T, cacheDir, category, key string) {
	t.Helper()

	if err := proxycache.Put(cacheDir, category, key, struct{}{}); err != nil {
		t.Fatal(err)
	}
}

func mustProxyCacheKeyExists(t *testing.T, cacheDir, category, key string) bool {
	t.Helper()

	exists, err := proxycache.KeyExists(cacheDir, category, key)
	if err != nil {
		t.Fatal(err)
	}

	return exists
}
