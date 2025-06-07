package pullreq_test

import (
	"context"
	"reflect"
	"testing"

	"go-kweb-lang/github"
	"go-kweb-lang/pullreq"
	"go-kweb-lang/pullreq/internal/mocks"
	"go-kweb-lang/testing/storetests"

	"go.uber.org/mock/gomock"
)

func TestFilePRFinder_LangIndex(t *testing.T) {
	langCode := "pl"

	for _, tc := range []struct {
		name              string
		init              func(t *testing.T, storageMock *mocks.MockFilePRFinderStorage)
		expectedLangIndex map[string][]int
	}{
		{
			name: "handle case when index does not exist yet",
			init: func(t *testing.T, storageMock *mocks.MockFilePRFinderStorage) {
				t.Helper()

				storageMock.EXPECT().
					LangIndex(langCode).
					Return(nil, nil)
			},
			expectedLangIndex: nil,
		},
		{
			name: "handle case when the index exists and contains data",
			init: func(t *testing.T, storageMock *mocks.MockFilePRFinderStorage) {
				t.Helper()

				storageMock.EXPECT().
					LangIndex(langCode).
					Return(map[string][]int{
						"/file1": {100, 200},
						"/file2": {300},
					}, nil)
			},
			expectedLangIndex: map[string][]int{
				"/file1": {100, 200},
				"/file2": {300},
			},
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			gitHubMock := mocks.NewMockGitHub(ctrl)
			cacheStore := mocks.NewMockCacheStore(ctrl)
			storageMock := mocks.NewMockFilePRFinderStorage(ctrl)

			filePRFinder := pullreq.NewFilePRFinder(
				gitHubMock,
				cacheStore,
				func(config *pullreq.FilePRFinderConfig) {
					config.Storage = storageMock
				},
			)

			tc.init(t, storageMock)

			langIndex, err := filePRFinder.LangIndex(langCode)
			if err != nil {
				t.Errorf("unexpected error: %v", err)
			}

			if !reflect.DeepEqual(tc.expectedLangIndex, langIndex) {
				t.Errorf("unexptected result\nexpected: %v\nactual  : %v", tc.expectedLangIndex, langIndex)
			}
		})
	}
}

func TestFilePRFinder_Update(t *testing.T) {
	ctx := context.Background()
	langCode := "pl"

	for _, tc := range []struct {
		name string
		init func(
			cacheDir string,
			gitHubMock *mocks.MockGitHub,
			cacheStore *mocks.MockCacheStore,
			storageMock *mocks.MockFilePRFinderStorage,
		)
	}{
		{
			name: "first run",
			init: func(
				cacheDir string,
				gitHubMock *mocks.MockGitHub,
				cacheStore *mocks.MockCacheStore,
				storageMock *mocks.MockFilePRFinderStorage,
			) {
				gitHubMock.EXPECT().
					PRSearch(
						ctx,
						github.PRSearchFilter{
							LangCode:    langCode,
							UpdatedFrom: "",
							OnlyOpen:    true,
						},
						github.PageRequest{
							Sort:    "updated",
							Order:   "asc",
							PerPage: 2,
						},
					).
					Return(
						&github.PRSearchResult{
							Items: []github.PRItem{
								{
									Number:    12,
									UpdatedAt: "D001",
								},
								{
									Number:    14,
									UpdatedAt: "D003",
								},
							},
							TotalCount: 3,
						},
						nil,
					).Times(1)

				gitHubMock.EXPECT().
					PRSearch(
						ctx,
						github.PRSearchFilter{
							LangCode:    langCode,
							UpdatedFrom: "D003",
							OnlyOpen:    true,
						},
						github.PageRequest{
							Sort:    "updated",
							Order:   "asc",
							PerPage: 2,
						},
					).
					Return(
						&github.PRSearchResult{
							Items: []github.PRItem{
								{
									Number:    15,
									UpdatedAt: "D004",
								},
							},
							TotalCount: 3,
						},
						nil,
					).Times(1)

				gitHubMock.EXPECT().
					PRSearch(
						ctx,
						github.PRSearchFilter{
							LangCode:    langCode,
							UpdatedFrom: "D004",
							OnlyOpen:    true,
						},
						github.PageRequest{
							Sort:    "updated",
							Order:   "asc",
							PerPage: 2,
						},
					).
					Return(
						&github.PRSearchResult{
							Items:      []github.PRItem{},
							TotalCount: 3,
						},
						nil,
					).Times(1)

				gitHubMock.EXPECT().GetPRCommits(ctx, 12).Return([]string{"C1"}, nil)
				gitHubMock.EXPECT().GetPRCommits(ctx, 14).Return([]string{"C2", "C3"}, nil)
				gitHubMock.EXPECT().GetPRCommits(ctx, 15).Return([]string{"C4"}, nil)

				cacheStore.EXPECT().
					Read(gomock.Any(), gomock.Any(), gomock.Any()).
					AnyTimes().
					DoAndReturn(storetests.MockReadNotFound())

				cacheStore.EXPECT().
					Write(gomock.Any(), gomock.Any(), gomock.Any()).Return(nil).
					AnyTimes().
					Return(nil)

				gitHubMock.EXPECT().GetCommitFiles(ctx, "C1").Return(&github.CommitFiles{Files: []string{"F1"}}, nil)
				gitHubMock.EXPECT().GetCommitFiles(ctx, "C2").Return(&github.CommitFiles{Files: []string{"F2", "F3"}}, nil)
				gitHubMock.EXPECT().GetCommitFiles(ctx, "C3").Return(&github.CommitFiles{Files: []string{"F1", "F4"}}, nil)
				gitHubMock.EXPECT().GetCommitFiles(ctx, "C4").Return(&github.CommitFiles{Files: []string{"F5"}}, nil)

				storageMock.EXPECT().StoreLangIndex(langCode, map[string][]int{
					"F1": {14, 12}, "F2": {14}, "F3": {14}, "F4": {14}, "F5": {15},
				}).Return(nil)
			},
		},
		{
			name: "there is no update",
			init: func(
				cacheDir string,
				gitHubMock *mocks.MockGitHub,
				cacheStore *mocks.MockCacheStore,
				storageMock *mocks.MockFilePRFinderStorage,
			) {
				gitHubMock.EXPECT().
					PRSearch(
						ctx,
						github.PRSearchFilter{
							LangCode:    langCode,
							UpdatedFrom: "",
							OnlyOpen:    true,
						},
						github.PageRequest{
							Sort:    "updated",
							Order:   "asc",
							PerPage: 2,
						},
					).
					Return(
						&github.PRSearchResult{
							Items: []github.PRItem{
								{
									Number:    12,
									UpdatedAt: "D001",
								},
								{
									Number:    14,
									UpdatedAt: "D003",
								},
							},
							TotalCount: 3,
						},
						nil,
					).Times(1)

				gitHubMock.EXPECT().
					PRSearch(
						ctx,
						github.PRSearchFilter{
							LangCode:    langCode,
							UpdatedFrom: "D003",
							OnlyOpen:    true,
						},
						github.PageRequest{
							Sort:    "updated",
							Order:   "asc",
							PerPage: 2,
						},
					).
					Return(
						&github.PRSearchResult{
							Items: []github.PRItem{
								{
									Number:    15,
									UpdatedAt: "D004",
								},
							},
							TotalCount: 3,
						},
						nil,
					).Times(1)

				gitHubMock.EXPECT().
					PRSearch(
						ctx,
						github.PRSearchFilter{
							LangCode:    langCode,
							UpdatedFrom: "D004",
							OnlyOpen:    true,
						},
						github.PageRequest{
							Sort:    "updated",
							Order:   "asc",
							PerPage: 2,
						},
					).
					Return(
						&github.PRSearchResult{
							Items:      []github.PRItem{},
							TotalCount: 3,
						},
						nil,
					).Times(1)

				cacheStore.EXPECT().
					Read("lang/pl/pr-pr-commits", "12", gomock.Any()).
					DoAndReturn(storetests.MockReadReturn(
						true,
						pullreq.PRCommits{
							UpdatedAt: "D001",
							CommitIds: []string{"C1"},
						}, nil,
					))

				cacheStore.EXPECT().
					Read("lang/pl/pr-commit-files", "C1", gomock.Any()).
					DoAndReturn(storetests.MockReadReturn(
						true,
						&github.CommitFiles{
							CommitID: "C1",
							Files:    []string{"F1"},
						}, nil))

				cacheStore.EXPECT().
					Read("lang/pl/pr-pr-commits", "14", gomock.Any()).
					DoAndReturn(storetests.MockReadReturn(
						true,
						pullreq.PRCommits{
							UpdatedAt: "D003",
							CommitIds: []string{"C2", "C3"},
						}, nil,
					))

				cacheStore.EXPECT().
					Read("lang/pl/pr-commit-files", "C2", gomock.Any()).
					DoAndReturn(storetests.MockReadReturn(
						true,
						&github.CommitFiles{
							CommitID: "C2",
							Files:    []string{"F2", "F3"},
						}, nil))
				cacheStore.EXPECT().
					Read("lang/pl/pr-commit-files", "C3", gomock.Any()).
					DoAndReturn(storetests.MockReadReturn(
						true,
						&github.CommitFiles{
							CommitID: "C3",
							Files:    []string{"F1", "F4"},
						}, nil))

				cacheStore.EXPECT().
					Read("lang/pl/pr-pr-commits", "15", gomock.Any()).
					DoAndReturn(storetests.MockReadReturn(
						true,
						pullreq.PRCommits{
							UpdatedAt: "D004",
							CommitIds: []string{"C4"},
						}, nil,
					))

				cacheStore.EXPECT().
					Read("lang/pl/pr-commit-files", "C4", gomock.Any()).
					DoAndReturn(storetests.MockReadReturn(
						true,
						&github.CommitFiles{
							CommitID: "C4",
							Files:    []string{"F5"},
						}, nil))

				storageMock.EXPECT().StoreLangIndex(langCode, map[string][]int{
					"F1": {14, 12}, "F2": {14}, "F3": {14}, "F4": {14}, "F5": {15},
				}).Return(nil)
			},
		},
		{
			name: "handle updates where a new file needs to be added to the index",
			init: func(
				cacheDir string,
				gitHubMock *mocks.MockGitHub,
				cacheStore *mocks.MockCacheStore,
				storageMock *mocks.MockFilePRFinderStorage,
			) {
				t.Helper()

				gitHubMock.EXPECT().
					PRSearch(
						ctx,
						github.PRSearchFilter{
							LangCode:    langCode,
							UpdatedFrom: "",
							OnlyOpen:    true,
						},
						github.PageRequest{
							Sort:    "updated",
							Order:   "asc",
							PerPage: 2,
						},
					).
					Return(
						&github.PRSearchResult{
							Items: []github.PRItem{
								{
									Number:    12,
									UpdatedAt: "D001",
								},
								{
									Number:    14,
									UpdatedAt: "D003",
								},
							},
							TotalCount: 3,
						},
						nil,
					).Times(1)

				gitHubMock.EXPECT().
					PRSearch(
						ctx,
						github.PRSearchFilter{
							LangCode:    langCode,
							UpdatedFrom: "D003",
							OnlyOpen:    true,
						},
						github.PageRequest{
							Sort:    "updated",
							Order:   "asc",
							PerPage: 2,
						},
					).
					Return(
						&github.PRSearchResult{
							Items: []github.PRItem{
								{
									Number:    15,
									UpdatedAt: "D005", // here
								},
							},
							TotalCount: 3,
						},
						nil,
					).Times(1)

				gitHubMock.EXPECT().
					PRSearch(
						ctx,
						github.PRSearchFilter{
							LangCode:    langCode,
							UpdatedFrom: "D005",
							OnlyOpen:    true,
						},
						github.PageRequest{
							Sort:    "updated",
							Order:   "asc",
							PerPage: 2,
						},
					).
					Return(
						&github.PRSearchResult{
							Items:      []github.PRItem{},
							TotalCount: 3,
						},
						nil,
					).Times(1)

				cacheStore.EXPECT().
					Read("lang/pl/pr-pr-commits", "12", gomock.Any()).
					DoAndReturn(storetests.MockReadReturn(
						true,
						pullreq.PRCommits{
							UpdatedAt: "D001",
							CommitIds: []string{"C1"},
						}, nil,
					))

				cacheStore.EXPECT().
					Read("lang/pl/pr-commit-files", "C1", gomock.Any()).
					DoAndReturn(storetests.MockReadReturn(
						true,
						&github.CommitFiles{
							CommitID: "C1",
							Files:    []string{"F1"},
						}, nil))

				cacheStore.EXPECT().
					Read("lang/pl/pr-pr-commits", "14", gomock.Any()).
					DoAndReturn(storetests.MockReadReturn(
						true,
						pullreq.PRCommits{
							UpdatedAt: "D003",
							CommitIds: []string{"C2", "C3"},
						}, nil,
					))

				cacheStore.EXPECT().
					Read("lang/pl/pr-commit-files", "C2", gomock.Any()).
					DoAndReturn(storetests.MockReadReturn(
						true,
						&github.CommitFiles{
							CommitID: "C2",
							Files:    []string{"F2", "F3"},
						}, nil))

				cacheStore.EXPECT().
					Read("lang/pl/pr-pr-commits", "15", gomock.Any()).
					DoAndReturn(storetests.MockReadReturn(
						true,
						pullreq.PRCommits{
							UpdatedAt: "D004",
							CommitIds: []string{"C4"},
						}, nil,
					))
				gitHubMock.EXPECT().GetPRCommits(ctx, 15).Return([]string{"C5"}, nil)
				cacheStore.EXPECT().Write(
					"lang/pl/pr-pr-commits",
					"15",
					pullreq.PRCommits{
						UpdatedAt: "D005",
						CommitIds: []string{"C5"},
					},
				)
				cacheStore.EXPECT().
					Read("lang/pl/pr-commit-files", "C5", gomock.Any()).
					DoAndReturn(storetests.MockReadNotFound())

				cacheStore.EXPECT().
					Read("lang/pl/pr-commit-files", "C3", gomock.Any()).
					DoAndReturn(storetests.MockReadReturn(
						true,
						&github.CommitFiles{
							CommitID: "C3",
							Files:    []string{"F1", "F4"},
						}, nil))

				gitHubMock.EXPECT().GetCommitFiles(ctx, "C5").Return(&github.CommitFiles{Files: []string{"F5", "F6"}}, nil)

				cacheStore.EXPECT().
					Write(
						"lang/pl/pr-commit-files",
						"C5",
						gomock.Any(),
					).Return(nil)

				storageMock.EXPECT().StoreLangIndex(langCode, map[string][]int{
					"F1": {14, 12}, "F2": {14}, "F3": {14}, "F4": {14}, "F5": {15}, "F6": {15},
				}).Return(nil)
			},
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			gitHub := mocks.NewMockGitHub(ctrl)
			prFinderStorage := mocks.NewMockFilePRFinderStorage(ctrl)
			cacheStore := mocks.NewMockCacheStore(ctrl)
			cacheDir := t.TempDir()

			tc.init(cacheDir, gitHub, cacheStore, prFinderStorage)

			filePRFinder := pullreq.NewFilePRFinder(
				gitHub,
				cacheStore,
				func(config *pullreq.FilePRFinderConfig) {
					config.Storage = prFinderStorage
					config.PerPage = 2
				},
			)

			err := filePRFinder.Update(ctx, langCode)
			if err != nil {
				t.Fatal(err)
			}
		})
	}
}
