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
	}{
		{
			name:     "fork is found",
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

			mergeCommit, err := gc.FindForkCommit(ctx, tc.commitID)
			if err != nil {
				t.Errorf("unexpected error: %v", err)
			}

			if !reflect.DeepEqual(mergeCommit, tc.expected) {
				t.Errorf("unexpected outcome\nactual   : %+v\nexpected: %+v", mergeCommit, tc.expected)
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
	}{
		{
			name:     "merge commit is found",
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

			mergeCommit, err := gc.FindMergeCommit(ctx, tc.commitID)
			if err != nil {
				t.Errorf("unexpected error: %v", err)
			}

			if !reflect.DeepEqual(mergeCommit, tc.expected) {
				t.Errorf("unexpected outcome\nactual   : %+v\nexpected: %+v", mergeCommit, tc.expected)
			}
		})
	}
}

func TestGitHist_PullRefresh(t *testing.T) {
	ctx := context.Background()

	const categoryMainBranchCommits = "git-main-branch-commits"

	for _, tc := range []struct {
		name     string
		initMock func(
			t *testing.T,
			gitRepoMock *mocks.MockRepo,
			invalidatorMock *mocks.MockInvalidator,
			cacheDir string,
			ctx context.Context,
		)
	}{
		{
			name: "no fresh commits",
			initMock: func(
				t *testing.T,
				gitRepoMock *mocks.MockRepo,
				invalidatorMock *mocks.MockInvalidator,
				cacheDir string,
				ctx context.Context,
			) {
				t.Helper()

				gitRepoMock.EXPECT().ListFreshCommits(ctx).Return(nil, nil)
				invalidatorMock.EXPECT().InvalidateFiles(nil).Return(nil)
			},
		},
		{
			name: "two fresh commits on the main branch",
			initMock: func(
				t *testing.T,
				gitRepoMock *mocks.MockRepo,
				invalidatorMock *mocks.MockInvalidator,
				cacheDir string,
				ctx context.Context,
			) {
				t.Helper()

				gitRepoMock.EXPECT().ListFreshCommits(ctx).Return(
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
				gitRepoMock.EXPECT().ListFilesInCommit(ctx, "CID1").Return([]string{"F10", "F11"}, nil)
				gitRepoMock.EXPECT().ListFilesInCommit(ctx, "CID2").Return([]string{"F11", "F12"}, nil)
				invalidatorMock.EXPECT().InvalidateFiles([]string{"F10", "F11", "F12"}).Return(nil)

				mustProxyCachePut(t, cacheDir, categoryMainBranchCommits, "")
			},
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			gitRepoMock := mocks.NewMockRepo(ctrl)
			invalidatorMock := mocks.NewMockInvalidator(ctrl)

			cacheDir := t.TempDir()

			gitRepoMock.EXPECT().Fetch(ctx).Return(nil)
			tc.initMock(t, gitRepoMock, invalidatorMock, cacheDir, ctx)
			gitRepoMock.EXPECT().Pull(ctx).Return(nil)

			gitRepoHist := githist.New(gitRepoMock, cacheDir)
			gitRepoHist.RegisterInvalidator(invalidatorMock)

			if err := gitRepoHist.PullRefresh(ctx); err != nil {
				t.Error(err)
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
