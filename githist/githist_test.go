//nolint:dupl
package githist_test

import (
	"context"
	"errors"
	"reflect"
	"testing"

	"github.com/dkarczmarski/go-kweb-lang/git"
	"github.com/dkarczmarski/go-kweb-lang/githist"
	"github.com/dkarczmarski/go-kweb-lang/githist/internal/mocks"
	"github.com/dkarczmarski/go-kweb-lang/testing/storetests"
	"go.uber.org/mock/gomock"
)

func TestGitHist_FindForkCommit(t *testing.T) {
	t.Parallel()

	ctx := t.Context()

	mainBranchCommits := []git.CommitInfo{
		{CommitID: "C-ID-40", DateTime: "DT-40", Comment: "TEXT-40"},
		{CommitID: "C-ID-30", DateTime: "DT-30", Comment: "TEXT-30"},
		{CommitID: "C-ID-20", DateTime: "DT-20", Comment: "TEXT-20"},
		{CommitID: "C-ID-10", DateTime: "DT-10", Comment: "TEXT-10"},
	}

	for _, tc := range []struct {
		name        string
		commitID    string
		expected    *git.CommitInfo
		expectedErr error
		initMock    func(gitRepo *mocks.MockGitRepo, commitID string)
	}{
		{
			name:     "fork is found",
			commitID: "C-ID-25",
			expected: &git.CommitInfo{CommitID: "C-ID-20", DateTime: "DT-20", Comment: "TEXT-20"},
			initMock: func(m *mocks.MockGitRepo, commitID string) {
				m.EXPECT().
					ListMainBranchCommits(ctx).
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
		{
			name:        "commit is already on main branch",
			commitID:    "C-ID-20",
			expected:    nil,
			expectedErr: githist.ErrCommitOnMainBranch,
			initMock: func(m *mocks.MockGitRepo, _ string) {
				m.EXPECT().
					ListMainBranchCommits(ctx).
					Return(mainBranchCommits, nil).
					Times(1)
			},
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			ctrl := gomock.NewController(t)
			gitRepo := mocks.NewMockGitRepo(ctrl)
			cache := mocks.NewMockCacheStorage(ctrl)

			cache.EXPECT().
				Read(githist.MainBranchCommitsCacheBucket(), githist.MainBranchCommitsCacheKey(), gomock.Any()).
				DoAndReturn(storetests.MockReadNotFound())

			cache.EXPECT().
				Write(githist.MainBranchCommitsCacheBucket(), githist.MainBranchCommitsCacheKey(), gomock.Any()).
				Return(nil)

			if tc.initMock != nil {
				tc.initMock(gitRepo, tc.commitID)
			}

			gc := githist.New(gitRepo, cache)

			forkCommit, err := gc.FindForkCommit(ctx, tc.commitID)
			if !errors.Is(err, tc.expectedErr) {
				t.Fatalf("unexpected error\nactual  : %v\nexpected: %v", err, tc.expectedErr)
			}

			if !reflect.DeepEqual(forkCommit, tc.expected) {
				t.Errorf("unexpected outcome\nactual  : %+v\nexpected: %+v", forkCommit, tc.expected)
			}
		})
	}
}

func TestGitHist_FindMergeCommit(t *testing.T) {
	t.Parallel()

	ctx := t.Context()

	mainBranchCommits := []git.CommitInfo{
		{CommitID: "C-ID-40", DateTime: "DT-40", Comment: "TEXT-40"},
		{CommitID: "C-ID-30", DateTime: "DT-30", Comment: "TEXT-30"},
		{CommitID: "C-ID-20", DateTime: "DT-20", Comment: "TEXT-20"},
		{CommitID: "C-ID-10", DateTime: "DT-10", Comment: "TEXT-10"},
	}

	for _, tc := range []struct {
		name        string
		commitID    string
		expected    *git.CommitInfo
		expectedErr error
		initMock    func(repo *mocks.MockGitRepo, commitID string)
	}{
		{
			name:     "merge commit is found",
			commitID: "C-ID-25",
			expected: &git.CommitInfo{CommitID: "C-ID-30", DateTime: "DT-30", Comment: "TEXT-30"},
			initMock: func(m *mocks.MockGitRepo, commitID string) {
				m.EXPECT().
					ListMainBranchCommits(ctx).
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
		{
			name:        "commit is already on main branch",
			commitID:    "C-ID-20",
			expected:    nil,
			expectedErr: githist.ErrCommitOnMainBranch,
			initMock: func(m *mocks.MockGitRepo, _ string) {
				m.EXPECT().
					ListMainBranchCommits(ctx).
					Return(mainBranchCommits, nil).
					Times(1)
			},
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			ctrl := gomock.NewController(t)
			gitRepo := mocks.NewMockGitRepo(ctrl)
			cache := mocks.NewMockCacheStorage(ctrl)

			cache.EXPECT().
				Read(githist.MainBranchCommitsCacheBucket(), githist.MainBranchCommitsCacheKey(), gomock.Any()).
				DoAndReturn(storetests.MockReadReturn[[]git.CommitInfo](false, nil, nil))

			cache.EXPECT().
				Write(githist.MainBranchCommitsCacheBucket(), githist.MainBranchCommitsCacheKey(), gomock.Any()).
				Return(nil)

			if tc.initMock != nil {
				tc.initMock(gitRepo, tc.commitID)
			}

			gc := githist.New(gitRepo, cache)

			mergeCommit, err := gc.FindMergeCommit(ctx, tc.commitID)
			if !errors.Is(err, tc.expectedErr) {
				t.Fatalf("unexpected error\nactual  : %v\nexpected: %v", err, tc.expectedErr)
			}

			if !reflect.DeepEqual(mergeCommit, tc.expected) {
				t.Errorf("unexpected outcome\nactual  : %+v\nexpected: %+v", mergeCommit, tc.expected)
			}
		})
	}
}

func TestGitHist_PullRefresh(t *testing.T) {
	t.Parallel()

	ctx := t.Context()

	for _, tc := range []struct {
		name          string
		expectedFiles []string
		initMock      func(context.Context, *mocks.MockGitRepo, *mocks.MockCacheStorage)
	}{
		{
			name:          "no fresh commits",
			expectedFiles: []string{},
			initMock: func(ctx context.Context, gitRepo *mocks.MockGitRepo, cache *mocks.MockCacheStorage) {
				gitRepo.EXPECT().
					ListMainBranchCommits(ctx).
					Return([]git.CommitInfo{
						{CommitID: "C-ID-1", DateTime: "DT-1", Comment: "Comment-1"},
					}, nil)

				cache.EXPECT().
					Read(githist.MainBranchCommitsCacheBucket(), githist.MainBranchCommitsCacheKey(), gomock.Any()).
					DoAndReturn(storetests.MockReadNotFound())

				cache.EXPECT().
					Write(githist.MainBranchCommitsCacheBucket(), githist.MainBranchCommitsCacheKey(), gomock.Any()).
					Return(nil)

				gitRepo.EXPECT().
					ListFreshCommits(ctx).
					Return(nil, nil)
			},
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			ctrl := gomock.NewController(t)
			gitRepo := mocks.NewMockGitRepo(ctrl)
			cache := mocks.NewMockCacheStorage(ctrl)

			gitRepo.EXPECT().Fetch(ctx).Return(nil)
			gitRepo.EXPECT().Pull(ctx).Return(nil)

			tc.initMock(ctx, gitRepo, cache)

			gitRepoHist := githist.New(gitRepo, cache)

			files, err := gitRepoHist.PullRefresh(ctx)
			if err != nil {
				t.Error(err)
			}

			if !reflect.DeepEqual(files, tc.expectedFiles) {
				t.Errorf("expected files %v, got %v", tc.expectedFiles, files)
			}
		})
	}
}

func TestGitHist_GetLastMainBranchCommit(t *testing.T) {
	t.Parallel()

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
				{CommitID: "C-ID-12", DateTime: "DT-12", Comment: "Comment-12"},
				{CommitID: "C-ID-11", DateTime: "DT-11", Comment: "Comment-11"},
				{CommitID: "C-ID-10", DateTime: "DT-10", Comment: "Comment-10"},
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

			ctx := t.Context()

			ctrl := gomock.NewController(t)
			gitRepo := mocks.NewMockGitRepo(ctrl)
			cache := mocks.NewMockCacheStorage(ctrl)

			gitRepoHist := githist.New(gitRepo, cache)

			cache.EXPECT().
				Read(githist.MainBranchCommitsCacheBucket(), githist.MainBranchCommitsCacheKey(), gomock.Any()).
				DoAndReturn(storetests.MockReadNotFound())

			cache.EXPECT().
				Write(githist.MainBranchCommitsCacheBucket(), githist.MainBranchCommitsCacheKey(), gomock.Any()).
				Return(nil)

			gitRepo.EXPECT().
				ListMainBranchCommits(ctx).
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
