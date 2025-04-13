package pullreq_test

import (
	"context"
	"reflect"
	"testing"

	"go-kweb-lang/pullreq/internal"

	"go-kweb-lang/proxycache"

	"go-kweb-lang/github"
	"go-kweb-lang/mocks"
	"go-kweb-lang/pullreq"

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
			storageMock := mocks.NewMockFilePRFinderStorage(ctrl)

			filePRFinder := pullreq.NewFilePRFinder(
				gitHubMock,
				t.TempDir(),
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
			t *testing.T,
			cacheDir string,
			gitHubMock *mocks.MockGitHub,
			storageMock *mocks.MockFilePRFinderStorage,
		)
	}{
		{
			name: "first run",
			init: func(
				t *testing.T,
				cacheDir string,
				gitHubMock *mocks.MockGitHub,
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
				t *testing.T,
				cacheDir string,
				gitHubMock *mocks.MockGitHub,
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

				must(t, proxycache.Put(cacheDir, internal.CategoryPrCommits, "12",
					internal.PRCommits{
						UpdatedAt: "D001",
						CommitIds: []string{"C1"},
					},
				))
				must(t, proxycache.Put(cacheDir, internal.CategoryPrCommits, "14",
					internal.PRCommits{
						UpdatedAt: "D003",
						CommitIds: []string{"C2", "C3"},
					},
				))
				must(t, proxycache.Put(cacheDir, internal.CategoryPrCommits, "15",
					internal.PRCommits{
						UpdatedAt: "D004",
						CommitIds: []string{"C4"},
					},
				))

				must(t, proxycache.Put(cacheDir, internal.CategoryCommitFiles, "C1",
					&github.CommitFiles{
						CommitID: "C1",
						Files:    []string{"F1"},
					},
				))
				must(t, proxycache.Put(cacheDir, internal.CategoryCommitFiles, "C2",
					&github.CommitFiles{
						CommitID: "C2",
						Files:    []string{"F2", "F3"},
					},
				))
				must(t, proxycache.Put(cacheDir, internal.CategoryCommitFiles, "C3",
					&github.CommitFiles{
						CommitID: "C3",
						Files:    []string{"F1", "F4"},
					},
				))
				must(t, proxycache.Put(cacheDir, internal.CategoryCommitFiles, "C4",
					&github.CommitFiles{
						CommitID: "C4",
						Files:    []string{"F5"},
					},
				))

				storageMock.EXPECT().StoreLangIndex(langCode, map[string][]int{
					"F1": {14, 12}, "F2": {14}, "F3": {14}, "F4": {14}, "F5": {15},
				}).Return(nil)
			},
		},
		{
			name: "handle updates where a new file needs to be added to the index",
			init: func(
				t *testing.T,
				cacheDir string,
				gitHubMock *mocks.MockGitHub,
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

				must(t, proxycache.Put(cacheDir, internal.CategoryPrCommits, "12",
					internal.PRCommits{
						UpdatedAt: "D001",
						CommitIds: []string{"C1"},
					},
				))
				must(t, proxycache.Put(cacheDir, internal.CategoryPrCommits, "14",
					internal.PRCommits{
						UpdatedAt: "D003",
						CommitIds: []string{"C2", "C3"},
					},
				))
				must(t, proxycache.Put(cacheDir, internal.CategoryPrCommits, "15",
					internal.PRCommits{
						UpdatedAt: "D004",
						CommitIds: []string{"C4"},
					},
				))
				gitHubMock.EXPECT().GetPRCommits(ctx, 15).Return([]string{"C5"}, nil)

				must(t, proxycache.Put(cacheDir, internal.CategoryCommitFiles, "C1",
					&github.CommitFiles{
						CommitID: "C1",
						Files:    []string{"F1"},
					},
				))
				must(t, proxycache.Put(cacheDir, internal.CategoryCommitFiles, "C2",
					&github.CommitFiles{
						CommitID: "C2",
						Files:    []string{"F2", "F3"},
					},
				))
				must(t, proxycache.Put(cacheDir, internal.CategoryCommitFiles, "C3",
					&github.CommitFiles{
						CommitID: "C3",
						Files:    []string{"F1", "F4"},
					},
				))
				gitHubMock.EXPECT().GetCommitFiles(ctx, "C5").Return(&github.CommitFiles{Files: []string{"F5", "F6"}}, nil)

				storageMock.EXPECT().StoreLangIndex(langCode, map[string][]int{
					"F1": {14, 12}, "F2": {14}, "F3": {14}, "F4": {14}, "F5": {15}, "F6": {15},
				}).Return(nil)
			},
		},
		{
			name: "handle updates when the PR list changes for an existing file",
			init: func(
				t *testing.T,
				cacheDir string,
				gitHubMock *mocks.MockGitHub,
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

				must(t, proxycache.Put(cacheDir, internal.CategoryPrCommits, "12",
					internal.PRCommits{
						UpdatedAt: "D001",
						CommitIds: []string{"C1"},
					},
				))
				must(t, proxycache.Put(cacheDir, internal.CategoryPrCommits, "14",
					internal.PRCommits{
						UpdatedAt: "D003",
						CommitIds: []string{"C2", "C3"},
					},
				))
				must(t, proxycache.Put(cacheDir, internal.CategoryPrCommits, "15",
					internal.PRCommits{
						UpdatedAt: "D004",
						CommitIds: []string{"C4"},
					},
				))
				// C5 is the new commit
				gitHubMock.EXPECT().GetPRCommits(ctx, 15).Return([]string{"C4", "C5"}, nil)

				must(t, proxycache.Put(cacheDir, internal.CategoryCommitFiles, "C1",
					&github.CommitFiles{
						CommitID: "C1",
						Files:    []string{"F1"},
					},
				))
				must(t, proxycache.Put(cacheDir, internal.CategoryCommitFiles, "C2",
					&github.CommitFiles{
						CommitID: "C2",
						Files:    []string{"F2", "F3"},
					},
				))
				must(t, proxycache.Put(cacheDir, internal.CategoryCommitFiles, "C3",
					&github.CommitFiles{
						CommitID: "C3",
						Files:    []string{"F1", "F4"},
					},
				))
				must(t, proxycache.Put(cacheDir, internal.CategoryCommitFiles, "C4",
					&github.CommitFiles{
						CommitID: "C4",
						Files:    []string{"F5"},
					},
				))
				gitHubMock.EXPECT().GetCommitFiles(ctx, "C5").Return(&github.CommitFiles{Files: []string{"F2"}}, nil)

				storageMock.EXPECT().StoreLangIndex(langCode, map[string][]int{
					"F1": {14, 12}, "F2": {15, 14}, "F3": {14}, "F4": {14}, "F5": {15},
				}).Return(nil)
			},
		},
		{
			name: "handle updates where a file has no remaining PRs and should be removed",
			init: func(
				t *testing.T,
				cacheDir string,
				gitHubMock *mocks.MockGitHub,
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
							Items:      []github.PRItem{},
							TotalCount: 3,
						},
						nil,
					).Times(1)

				must(t, proxycache.Put(cacheDir, internal.CategoryPrCommits, "12",
					internal.PRCommits{
						UpdatedAt: "D001",
						CommitIds: []string{"C1"},
					},
				))
				must(t, proxycache.Put(cacheDir, internal.CategoryPrCommits, "14",
					internal.PRCommits{
						UpdatedAt: "D003",
						CommitIds: []string{"C2", "C3"},
					},
				))

				must(t, proxycache.Put(cacheDir, internal.CategoryCommitFiles, "C1",
					&github.CommitFiles{
						CommitID: "C1",
						Files:    []string{"F1"},
					},
				))
				must(t, proxycache.Put(cacheDir, internal.CategoryCommitFiles, "C2",
					&github.CommitFiles{
						CommitID: "C2",
						Files:    []string{"F2", "F3"},
					},
				))
				must(t, proxycache.Put(cacheDir, internal.CategoryCommitFiles, "C3",
					&github.CommitFiles{
						CommitID: "C3",
						Files:    []string{"F1", "F4"},
					},
				))
				must(t, proxycache.Put(cacheDir, internal.CategoryCommitFiles, "C4",
					&github.CommitFiles{
						CommitID: "C4",
						Files:    []string{"F5"},
					},
				))

				storageMock.EXPECT().StoreLangIndex(langCode, map[string][]int{
					"F1": {14, 12}, "F2": {14}, "F3": {14}, "F4": {14},
				}).Return(nil)
			},
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			gitHubMock := mocks.NewMockGitHub(ctrl)
			storageMock := mocks.NewMockFilePRFinderStorage(ctrl)
			cacheDir := t.TempDir()

			tc.init(t, cacheDir, gitHubMock, storageMock)

			filePRFinder := pullreq.NewFilePRFinder(
				gitHubMock,
				cacheDir,
				func(config *pullreq.FilePRFinderConfig) {
					config.Storage = storageMock
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

func must(t *testing.T, err error) {
	if err != nil {
		t.Fatal(err)
	}
}
