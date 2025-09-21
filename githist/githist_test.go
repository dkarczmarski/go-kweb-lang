package githist_test

import (
	"context"
	"reflect"
	"testing"

	"go-kweb-lang/git"
	"go-kweb-lang/githist"
	"go-kweb-lang/githist/internal/mocks"
	"go-kweb-lang/testing/storetests"

	"go.uber.org/mock/gomock"
)

const bucketMainBranchCommits = "git-main-branch-commits"

func TestGitHist_FindForkCommit(t *testing.T) {
	ctx := context.Background()

	mainBranchCommits := []git.CommitInfo{
		{CommitID: "C-ID-40", DateTime: "DT-40", Comment: "TEXT-40"},
		{CommitID: "C-ID-30", DateTime: "DT-30", Comment: "TEXT-30"},
		{CommitID: "C-ID-20", DateTime: "DT-20", Comment: "TEXT-20"},
		{CommitID: "C-ID-10", DateTime: "DT-10", Comment: "TEXT-10"},
	}

	for _, tc := range []struct {
		name     string
		commitID string
		expected *git.CommitInfo
		initMock func(gitRepo *mocks.MockGitRepo, commitID string)
	}{
		{
			name:     "fork is found",
			commitID: "C-ID-25",
			expected: &git.CommitInfo{CommitID: "C-ID-20", DateTime: "DT-20", Comment: "TEXT-20"},
			initMock: func(m *mocks.MockGitRepo, commitID string) {
				m.EXPECT().ListMainBranchCommits(ctx).
					Return(mainBranchCommits, nil).
					Times(1)

				m.EXPECT().
					ListAncestorCommits(ctx, commitID).
					Return([]git.CommitInfo{
						{CommitID: "C-ID-20", DateTime: "DT-20", Comment: "TEXT-20"},
						{CommitID: "C-ID-10", DateTime: "DT-10", Comment: "TEXT-10"},
					}, nil).
					Times(1)
			},
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			ctrl := gomock.NewController(t)
			gitRepo := mocks.NewMockGitRepo(ctrl)
			cacheStore := mocks.NewMockCacheStore(ctrl)

			cacheStore.EXPECT().
				Read(bucketMainBranchCommits, "", gomock.Any()).
				DoAndReturn(storetests.MockReadNotFound())

			cacheStore.EXPECT().
				Write(gomock.Any(), gomock.Any(), gomock.Any()).
				Return(nil)

			if tc.initMock != nil {
				tc.initMock(gitRepo, tc.commitID)
			}

			gc := githist.New(gitRepo, cacheStore)

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
		{CommitID: "C-ID-40", DateTime: "DT-40", Comment: "TEXT-40"},
		{CommitID: "C-ID-30", DateTime: "DT-30", Comment: "TEXT-30"},
		{CommitID: "C-ID-20", DateTime: "DT-20", Comment: "TEXT-20"},
		{CommitID: "C-ID-10", DateTime: "DT-10", Comment: "TEXT-10"},
	}

	for _, tc := range []struct {
		name     string
		commitID string
		expected *git.CommitInfo
		initMock func(repo *mocks.MockGitRepo, commitID string)
	}{
		{
			name:     "merge commit is found",
			commitID: "C-ID-25",
			expected: &git.CommitInfo{CommitID: "C-ID-30", DateTime: "DT-30", Comment: "TEXT-30"},
			initMock: func(m *mocks.MockGitRepo, commitID string) {
				m.EXPECT().ListMainBranchCommits(ctx).
					Return(mainBranchCommits, nil).
					Times(1)

				m.EXPECT().
					ListMergePoints(ctx, commitID).
					Return([]git.CommitInfo{
						{CommitID: "C-ID-30", DateTime: "DT-30", Comment: "TEXT-30"},
						{CommitID: "C-ID-40", DateTime: "DT-40", Comment: "TEXT-40"},
					}, nil).
					Times(1)
			},
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			ctrl := gomock.NewController(t)
			gitRepo := mocks.NewMockGitRepo(ctrl)
			cacheStore := mocks.NewMockCacheStore(ctrl)

			cacheStore.EXPECT().
				Read(bucketMainBranchCommits, "", gomock.Any()).
				DoAndReturn(storetests.MockReadReturn[[]git.CommitInfo](
					false, nil, nil),
				)

			cacheStore.EXPECT().
				Write(gomock.Any(), gomock.Any(), gomock.Any()).Return(nil)

			if tc.initMock != nil {
				tc.initMock(gitRepo, tc.commitID)
			}

			gc := githist.New(gitRepo, cacheStore)

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

	for _, tc := range []struct {
		name     string
		initMock func(
			ctx context.Context,
			gitRepo *mocks.MockGitRepo,
			cacheStore *mocks.MockCacheStore,
			invalidator *mocks.MockInvalidator,
		)
	}{
		{
			name: "no fresh commits",
			initMock: func(
				ctx context.Context,
				gitRepo *mocks.MockGitRepo,
				cacheStore *mocks.MockCacheStore,
				invalidator *mocks.MockInvalidator,
			) {
				gitRepo.EXPECT().ListMainBranchCommits(ctx).
					Return([]git.CommitInfo{
						{CommitID: "C-ID-1", DateTime: "DT-1", Comment: "Comment-1"},
					}, nil).
					Times(1)

				cacheStore.EXPECT().
					Read(bucketMainBranchCommits, "", gomock.Any()).
					DoAndReturn(storetests.MockReadNotFound()).
					Times(1)

				cacheStore.EXPECT().
					Write(bucketMainBranchCommits, "", gomock.Any()).
					Return(nil).
					Times(1)

				cacheStore.EXPECT().
					Read(bucketMainBranchCommits, "", gomock.Any()).
					DoAndReturn(storetests.MockReadReturn[[]git.CommitInfo](
						true,
						[]git.CommitInfo{
							{CommitID: "C-ID-1", DateTime: "DT-1", Comment: "Comment-1"},
						},
						nil,
					)).
					Times(1)

				gitRepo.EXPECT().ListFreshCommits(ctx).Return(nil, nil)
			},
		},
		{
			name: "fresh commits on the main branch",
			initMock: func(
				ctx context.Context,
				gitRepo *mocks.MockGitRepo,
				cacheStore *mocks.MockCacheStore,
				invalidator *mocks.MockInvalidator,
			) {
				mainBranch1 := []git.CommitInfo{
					{CommitID: "C-ID-0", DateTime: "DT-0", Comment: "Comment-0"},
				}
				mainBranch2 := []git.CommitInfo{
					{CommitID: "C-ID-2", DateTime: "DT-2", Comment: "Comment-2"},
					{CommitID: "C-ID-1", DateTime: "DT-1", Comment: "Comment-1"},
					{CommitID: "C-ID-0", DateTime: "DT-0", Comment: "Comment-0"},
				}

				fresh := []git.CommitInfo{
					{CommitID: "C-ID-1", DateTime: "DT-1", Comment: "Comment-1"},
					{CommitID: "C-ID-2", DateTime: "DT-2", Comment: "Comment-2"},
				}
				filesC1 := []string{"content/fr/file1", "content/en/file2"}
				filesC2 := []string{"dir/file3", "content/en/dir/file4"}

				gomock.InOrder(
					// GetLastMainBranchCommit
					// first Read -> not found -> ListMainBranchCommits -> Write
					cacheStore.EXPECT().
						Read(bucketMainBranchCommits, "", gomock.Any()).
						DoAndReturn(storetests.MockReadNotFound()),

					gitRepo.EXPECT().
						ListMainBranchCommits(ctx).
						Return(mainBranch1, nil),

					cacheStore.EXPECT().
						Write(bucketMainBranchCommits, "", gomock.Any()).
						Return(nil),

					// fresh commits
					gitRepo.EXPECT().
						ListFreshCommits(ctx).
						Return(fresh, nil),

					// delete cache entry because fresh commits were found so main branch commits have changed
					cacheStore.EXPECT().
						Delete(bucketMainBranchCommits, "").
						Return(nil),

					// listMainBranchCommits after Delete()
					// second Read -> not found -> ListMainBranchCommits -> Write
					cacheStore.EXPECT().
						Read(bucketMainBranchCommits, "", gomock.Any()).
						DoAndReturn(storetests.MockReadNotFound()),

					gitRepo.EXPECT().
						ListMainBranchCommits(ctx).
						Return(mainBranch2, nil),

					cacheStore.EXPECT().
						Write(bucketMainBranchCommits, "", gomock.Any()).
						Return(nil),

					//
					gitRepo.EXPECT().
						ListFilesInCommit(ctx, "C-ID-1").
						Return(filesC1, nil),

					// file invalidation
					invalidator.EXPECT().InvalidateFile("content/fr/file1").Return(nil),
					invalidator.EXPECT().InvalidateFile("content/en/file2").Return(nil),

					//
					gitRepo.EXPECT().
						ListFilesInCommit(ctx, "C-ID-2").
						Return(filesC2, nil),

					// file invalidation
					invalidator.EXPECT().InvalidateFile("dir/file3").Return(nil),
					invalidator.EXPECT().InvalidateFile("content/en/dir/file4").Return(nil),
				)
			},
		},
		{
			name: "fresh merge commits",
			initMock: func(
				ctx context.Context,
				gitRepo *mocks.MockGitRepo,
				cacheStore *mocks.MockCacheStore,
				invalidator *mocks.MockInvalidator,
			) {
				mainBranch1 := []git.CommitInfo{
					{CommitID: "C-ID-0", DateTime: "DT-0", Comment: "Comment-0"},
				}

				fresh := []git.CommitInfo{
					{CommitID: "C-ID-1", DateTime: "DT-1", Comment: "Comment-1"},
					{CommitID: "C-ID-2", DateTime: "DT-2", Comment: "Comment-2"},
				}

				mainBranch2 := []git.CommitInfo{
					{CommitID: "C-ID-2", DateTime: "DT-2", Comment: "Comment-2"},
					{CommitID: "C-ID-1", DateTime: "DT-1", Comment: "Comment-1"},
					{CommitID: "C-ID-10", DateTime: "DT-10", Comment: "Comment-10"},
					{CommitID: "C-ID-13", DateTime: "DT-13", Comment: "Comment-13"},
					{CommitID: "C-ID-103", DateTime: "DT-103", Comment: "Comment-103"},
					{CommitID: "C-ID-102", DateTime: "DT-102", Comment: "Comment-102"},
					{CommitID: "C-ID-0", DateTime: "DT-0", Comment: "Comment-0"},
				}

				gomock.InOrder(
					// GetLastMainBranchCommit
					// first Read -> not found -> ListMainBranchCommits -> Write
					cacheStore.EXPECT().
						Read(bucketMainBranchCommits, "", gomock.Any()).
						DoAndReturn(storetests.MockReadNotFound()),

					gitRepo.EXPECT().
						ListMainBranchCommits(ctx).
						Return(mainBranch1, nil),

					cacheStore.EXPECT().
						Write(bucketMainBranchCommits, "", gomock.Any()).
						Return(nil),

					// fresh commits
					gitRepo.EXPECT().
						ListFreshCommits(ctx).
						Return(fresh, nil),

					// delete cache entry because fresh commits were found so main branch commits have changed
					cacheStore.EXPECT().
						Delete(bucketMainBranchCommits, "").
						Return(nil),

					// listMainBranchCommits after Delete()
					// second Read -> not found -> ListMainBranchCommits -> Write
					cacheStore.EXPECT().
						Read(bucketMainBranchCommits, "", gomock.Any()).
						DoAndReturn(storetests.MockReadNotFound()),

					gitRepo.EXPECT().
						ListMainBranchCommits(ctx).
						Return(mainBranch2, nil),

					cacheStore.EXPECT().
						Write(bucketMainBranchCommits, "", gomock.Any()).
						Return(nil),

					//
					gitRepo.EXPECT().
						ListFilesInCommit(ctx, "C-ID-1").
						Return([]string{}, nil),

					gitRepo.EXPECT().
						ListCommitParents(ctx, "C-ID-1").
						Return([]string{"C-ID-10", "C-ID-11"}, nil),

					gitRepo.EXPECT().
						ListAncestorCommits(ctx, "C-ID-11").
						Return([]git.CommitInfo{
							{CommitID: "C-ID-11-2", DateTime: "DT-11-2", Comment: "Comment-11-2"},
							{CommitID: "C-ID-11-1", DateTime: "DT-11-1", Comment: "Comment-1-11"},
							{CommitID: "C-ID-102", DateTime: "DT-102", Comment: "Comment-102"}, // on main
						}, nil),

					gitRepo.EXPECT().
						ListFilesBetweenCommits(ctx, "C-ID-102", "C-ID-11").
						Return([]string{"dir/file1", "dir/file2"}, nil),

					// file invalidation
					invalidator.EXPECT().InvalidateFile("dir/file1").Return(nil),
					invalidator.EXPECT().InvalidateFile("dir/file2").Return(nil),

					//
					gitRepo.EXPECT().
						ListFilesInCommit(ctx, "C-ID-2").
						Return([]string{}, nil),

					gitRepo.EXPECT().
						ListCommitParents(ctx, "C-ID-2").
						Return([]string{"C-ID-12", "C-ID-13"}, nil),

					gitRepo.EXPECT().
						ListAncestorCommits(ctx, "C-ID-12").
						Return([]git.CommitInfo{
							{CommitID: "C-ID-12-1", DateTime: "DT-12-1", Comment: "Comment-12-1"},
							{CommitID: "C-ID-103", DateTime: "DT-103", Comment: "Comment-103"}, // on main
						}, nil),

					gitRepo.EXPECT().
						ListFilesBetweenCommits(ctx, "C-ID-103", "C-ID-12").
						Return([]string{"dir/file3"}, nil),

					// file invalidation
					invalidator.EXPECT().InvalidateFile("dir/file3").Return(nil),
				)
			},
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			gitRepo := mocks.NewMockGitRepo(ctrl)
			invalidator := mocks.NewMockInvalidator(ctrl)
			cacheStore := mocks.NewMockCacheStore(ctrl)

			gitRepo.EXPECT().Fetch(ctx).Return(nil)
			gitRepo.EXPECT().Pull(ctx).Return(nil)

			tc.initMock(ctx, gitRepo, cacheStore, invalidator)

			gitRepoHist := githist.New(gitRepo, cacheStore)
			gitRepoHist.RegisterInvalidator(invalidator)

			if err := gitRepoHist.PullRefresh(ctx); err != nil {
				t.Error(err)
			}
		})
	}
}

func TestGitHist_GetLastMainBranchCommit(t *testing.T) {
	for _, tc := range []struct {
		name       string
		mainBranch []git.CommitInfo
		expected   git.CommitInfo
	}{
		{
			name:       "empty list of the main branch commits",
			mainBranch: nil,
			expected:   git.CommitInfo{},
		},
		{
			name: "not empty list of the main branch commits",
			mainBranch: []git.CommitInfo{
				{
					CommitID: "C-ID-12",
					DateTime: "DT-12",
					Comment:  "Comment-12",
				},
				{
					CommitID: "C-ID-11",
					DateTime: "DT-11",
					Comment:  "Comment-11",
				},
				{
					CommitID: "C-ID-10",
					DateTime: "DT-10",
					Comment:  "Comment-10",
				},
			},
			expected: git.CommitInfo{
				CommitID: "C-ID-12",
				DateTime: "DT-12",
				Comment:  "Comment-12",
			},
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			ctx := context.Background()

			ctrl := gomock.NewController(t)
			gitRepo := mocks.NewMockGitRepo(ctrl)
			cacheStore := mocks.NewMockCacheStore(ctrl)
			gitRepoHist := githist.New(gitRepo, cacheStore)

			cacheStore.EXPECT().Read(gomock.Any(), gomock.Any(), gomock.Any()).
				DoAndReturn(storetests.MockReadNotFound())
			cacheStore.EXPECT().Write(gomock.Any(), gomock.Any(), gomock.Any()).
				Return(nil)

			gitRepo.EXPECT().ListMainBranchCommits(ctx).
				Return(tc.mainBranch, nil)

			commit, err := gitRepoHist.GetLastMainBranchCommit(ctx)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			if !reflect.DeepEqual(tc.expected, commit) {
				t.Errorf("unexpected outcome: %+v", commit)
			}
		})
	}
}
