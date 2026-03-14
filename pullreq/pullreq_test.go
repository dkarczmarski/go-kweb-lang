//nolint:paralleltest
package pullreq_test

import (
	"errors"
	"reflect"
	"testing"

	"github.com/dkarczmarski/go-kweb-lang/github"
	"github.com/dkarczmarski/go-kweb-lang/pullreq"
	"github.com/dkarczmarski/go-kweb-lang/pullreq/internal/cachetypes"
	"github.com/dkarczmarski/go-kweb-lang/pullreq/internal/mocks"
	"github.com/dkarczmarski/go-kweb-lang/testing/storetests"
	"go.uber.org/mock/gomock"
)

func TestFilePRIndex_LangIndex(t *testing.T) {
	langCode := "pl"

	for _, tc := range []struct {
		name              string
		init              func(t *testing.T, cacheStore *mocks.MockCacheStorage)
		expectedLangIndex pullreq.FilePRIndexData
		expectedErr       error
	}{
		{
			name: "handle case when index does not exist yet",
			init: func(t *testing.T, cacheStore *mocks.MockCacheStorage) {
				t.Helper()

				cacheStore.EXPECT().
					Read(
						pullreq.FilePRsIndexCacheBucket(langCode),
						pullreq.FilePRsIndexCacheKey(langCode),
						gomock.Any(),
					).
					DoAndReturn(storetests.MockReadReturn[pullreq.FilePRIndexData](false, nil, nil))
			},
			expectedLangIndex: nil,
			expectedErr:       pullreq.ErrLangIndexNotFound,
		},
		{
			name: "handle case when the index exists and contains data",
			init: func(t *testing.T, cacheStore *mocks.MockCacheStorage) {
				t.Helper()

				cacheStore.EXPECT().
					Read(
						pullreq.FilePRsIndexCacheBucket(langCode),
						pullreq.FilePRsIndexCacheKey(langCode),
						gomock.Any(),
					).
					DoAndReturn(storetests.MockReadReturn(
						true,
						pullreq.FilePRIndexData{
							"/file1": {100, 200},
							"/file2": {300},
						},
						nil,
					))
			},
			expectedLangIndex: pullreq.FilePRIndexData{
				"/file1": {100, 200},
				"/file2": {300},
			},
			expectedErr: nil,
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			gitHubMock := mocks.NewMockGitHub(ctrl)
			cacheStore := mocks.NewMockCacheStorage(ctrl)

			filePRIndex := pullreq.NewFilePRIndex(gitHubMock, cacheStore, 2)

			tc.init(t, cacheStore)

			langIndex, err := filePRIndex.LangIndex(langCode)
			if !errors.Is(err, tc.expectedErr) {
				t.Fatalf("unexpected error\nexpected: %v\nactual  : %v", tc.expectedErr, err)
			}

			if !reflect.DeepEqual(tc.expectedLangIndex, langIndex) {
				t.Errorf("unexpected result\nexpected: %v\nactual  : %v", tc.expectedLangIndex, langIndex)
			}
		})
	}
}

func TestFilePRIndex_RefreshIndex(t *testing.T) {
	ctx := t.Context()
	langCode := "pl"

	for _, tc := range []struct {
		name string
		init func(
			t *testing.T,
			gitHubMock *mocks.MockGitHub,
			cacheStore *mocks.MockCacheStorage,
		)
	}{
		{
			name: "first run",
			init: func(
				t *testing.T,
				gitHubMock *mocks.MockGitHub,
				cacheStore *mocks.MockCacheStorage,
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
							Page:    1,
							PerPage: 2,
						},
					).
					Return(
						&github.PRSearchResult{
							Items: []github.PRItem{
								{Number: 12, UpdatedAt: "D001"},
								{Number: 14, UpdatedAt: "D003"},
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
							Page:    1,
							PerPage: 2,
						},
					).
					Return(
						&github.PRSearchResult{
							Items: []github.PRItem{
								{Number: 15, UpdatedAt: "D004"},
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
							Page:    1,
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
					Read(
						pullreq.PRCommitsCacheBucket(langCode),
						pullreq.PRCommitsCacheKey(12),
						gomock.Any(),
					).
					DoAndReturn(storetests.MockReadNotFound())

				gitHubMock.EXPECT().
					GetPRCommits(ctx, 12).
					Return([]string{"C1"}, nil)

				cacheStore.EXPECT().
					Write(
						pullreq.PRCommitsCacheBucket(langCode),
						pullreq.PRCommitsCacheKey(12),
						cachetypes.PRCommits{
							UpdatedAt: "D001",
							CommitIDs: []string{"C1"},
						},
					).
					Return(nil)

				cacheStore.EXPECT().
					Read(
						pullreq.CommitFilesCacheBucket(langCode),
						pullreq.CommitFilesCacheKey("C1"),
						gomock.Any(),
					).
					DoAndReturn(storetests.MockReadNotFound())

				gitHubMock.EXPECT().
					GetCommitFiles(ctx, "C1").
					Return(&github.CommitFiles{Files: []string{"content/pl/F1"}}, nil)

				cacheStore.EXPECT().
					Write(
						pullreq.CommitFilesCacheBucket(langCode),
						pullreq.CommitFilesCacheKey("C1"),
						gomock.Any(),
					).
					Return(nil)

				cacheStore.EXPECT().
					Read(
						pullreq.PRCommitsCacheBucket(langCode),
						pullreq.PRCommitsCacheKey(14),
						gomock.Any(),
					).
					DoAndReturn(storetests.MockReadNotFound())

				gitHubMock.EXPECT().
					GetPRCommits(ctx, 14).
					Return([]string{"C2", "C3"}, nil)

				cacheStore.EXPECT().
					Write(
						pullreq.PRCommitsCacheBucket(langCode),
						pullreq.PRCommitsCacheKey(14),
						cachetypes.PRCommits{
							UpdatedAt: "D003",
							CommitIDs: []string{"C2", "C3"},
						},
					).
					Return(nil)

				cacheStore.EXPECT().
					Read(
						pullreq.CommitFilesCacheBucket(langCode),
						pullreq.CommitFilesCacheKey("C2"),
						gomock.Any(),
					).
					DoAndReturn(storetests.MockReadNotFound())

				gitHubMock.EXPECT().
					GetCommitFiles(ctx, "C2").
					Return(&github.CommitFiles{Files: []string{"content/pl/F2", "content/pl/F3"}}, nil)

				cacheStore.EXPECT().
					Write(
						pullreq.CommitFilesCacheBucket(langCode),
						pullreq.CommitFilesCacheKey("C2"),
						gomock.Any(),
					).
					Return(nil)

				cacheStore.EXPECT().
					Read(
						pullreq.CommitFilesCacheBucket(langCode),
						pullreq.CommitFilesCacheKey("C3"),
						gomock.Any(),
					).
					DoAndReturn(storetests.MockReadNotFound())

				gitHubMock.EXPECT().
					GetCommitFiles(ctx, "C3").
					Return(&github.CommitFiles{Files: []string{"content/pl/F1", "content/pl/F4"}}, nil)

				cacheStore.EXPECT().
					Write(
						pullreq.CommitFilesCacheBucket(langCode),
						pullreq.CommitFilesCacheKey("C3"),
						gomock.Any(),
					).
					Return(nil)

				cacheStore.EXPECT().
					Read(
						pullreq.PRCommitsCacheBucket(langCode),
						pullreq.PRCommitsCacheKey(15),
						gomock.Any(),
					).
					DoAndReturn(storetests.MockReadNotFound())

				gitHubMock.EXPECT().
					GetPRCommits(ctx, 15).
					Return([]string{"C4"}, nil)

				cacheStore.EXPECT().
					Write(
						pullreq.PRCommitsCacheBucket(langCode),
						pullreq.PRCommitsCacheKey(15),
						cachetypes.PRCommits{
							UpdatedAt: "D004",
							CommitIDs: []string{"C4"},
						},
					).
					Return(nil)

				cacheStore.EXPECT().
					Read(
						pullreq.CommitFilesCacheBucket(langCode),
						pullreq.CommitFilesCacheKey("C4"),
						gomock.Any(),
					).
					DoAndReturn(storetests.MockReadNotFound())

				gitHubMock.EXPECT().
					GetCommitFiles(ctx, "C4").
					Return(&github.CommitFiles{Files: []string{"content/pl/F5"}}, nil)

				cacheStore.EXPECT().
					Write(
						pullreq.CommitFilesCacheBucket(langCode),
						pullreq.CommitFilesCacheKey("C4"),
						gomock.Any(),
					).
					Return(nil)

				cacheStore.EXPECT().
					Write(
						pullreq.FilePRsIndexCacheBucket(langCode),
						pullreq.FilePRsIndexCacheKey(langCode),
						pullreq.FilePRIndexData{
							"content/pl/F1": {14, 12},
							"content/pl/F2": {14},
							"content/pl/F3": {14},
							"content/pl/F4": {14},
							"content/pl/F5": {15},
						},
					).
					Return(nil)
			},
		},
		{
			name: "there is no update",
			init: func(
				t *testing.T,
				gitHubMock *mocks.MockGitHub,
				cacheStore *mocks.MockCacheStorage,
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
							Page:    1,
							PerPage: 2,
						},
					).
					Return(
						&github.PRSearchResult{
							Items: []github.PRItem{
								{Number: 12, UpdatedAt: "D001"},
								{Number: 14, UpdatedAt: "D003"},
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
							Page:    1,
							PerPage: 2,
						},
					).
					Return(
						&github.PRSearchResult{
							Items: []github.PRItem{
								{Number: 15, UpdatedAt: "D004"},
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
							Page:    1,
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
					Read(
						pullreq.PRCommitsCacheBucket(langCode),
						pullreq.PRCommitsCacheKey(12),
						gomock.Any(),
					).
					DoAndReturn(storetests.MockReadReturn(
						true,
						cachetypes.PRCommits{
							UpdatedAt: "D001",
							CommitIDs: []string{"C1"},
						},
						nil,
					))

				cacheStore.EXPECT().
					Read(
						pullreq.CommitFilesCacheBucket(langCode),
						pullreq.CommitFilesCacheKey("C1"),
						gomock.Any(),
					).
					DoAndReturn(storetests.MockReadReturn(
						true,
						&github.CommitFiles{
							CommitID: "C1",
							Files:    []string{"content/pl/F1"},
						},
						nil,
					))

				cacheStore.EXPECT().
					Read(
						pullreq.PRCommitsCacheBucket(langCode),
						pullreq.PRCommitsCacheKey(14),
						gomock.Any(),
					).
					DoAndReturn(storetests.MockReadReturn(
						true,
						cachetypes.PRCommits{
							UpdatedAt: "D003",
							CommitIDs: []string{"C2", "C3"},
						},
						nil,
					))

				cacheStore.EXPECT().
					Read(
						pullreq.CommitFilesCacheBucket(langCode),
						pullreq.CommitFilesCacheKey("C2"),
						gomock.Any(),
					).
					DoAndReturn(storetests.MockReadReturn(
						true,
						&github.CommitFiles{
							CommitID: "C2",
							Files:    []string{"content/pl/F2", "content/pl/F3"},
						},
						nil,
					))

				cacheStore.EXPECT().
					Read(
						pullreq.CommitFilesCacheBucket(langCode),
						pullreq.CommitFilesCacheKey("C3"),
						gomock.Any(),
					).
					DoAndReturn(storetests.MockReadReturn(
						true,
						&github.CommitFiles{
							CommitID: "C3",
							Files:    []string{"content/pl/F1", "content/pl/F4"},
						},
						nil,
					))

				cacheStore.EXPECT().
					Read(
						pullreq.PRCommitsCacheBucket(langCode),
						pullreq.PRCommitsCacheKey(15),
						gomock.Any(),
					).
					DoAndReturn(storetests.MockReadReturn(
						true,
						cachetypes.PRCommits{
							UpdatedAt: "D004",
							CommitIDs: []string{"C4"},
						},
						nil,
					))

				cacheStore.EXPECT().
					Read(
						pullreq.CommitFilesCacheBucket(langCode),
						pullreq.CommitFilesCacheKey("C4"),
						gomock.Any(),
					).
					DoAndReturn(storetests.MockReadReturn(
						true,
						&github.CommitFiles{
							CommitID: "C4",
							Files:    []string{"content/pl/F5"},
						},
						nil,
					))

				cacheStore.EXPECT().
					Write(
						pullreq.FilePRsIndexCacheBucket(langCode),
						pullreq.FilePRsIndexCacheKey(langCode),
						pullreq.FilePRIndexData{
							"content/pl/F1": {14, 12},
							"content/pl/F2": {14},
							"content/pl/F3": {14},
							"content/pl/F4": {14},
							"content/pl/F5": {15},
						},
					).
					Return(nil)
			},
		},
		{
			name: "handle updates where a new file needs to be added to the index",
			init: func(
				t *testing.T,
				gitHubMock *mocks.MockGitHub,
				cacheStore *mocks.MockCacheStorage,
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
							Page:    1,
							PerPage: 2,
						},
					).
					Return(
						&github.PRSearchResult{
							Items: []github.PRItem{
								{Number: 12, UpdatedAt: "D001"},
								{Number: 14, UpdatedAt: "D003"},
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
							Page:    1,
							PerPage: 2,
						},
					).
					Return(
						&github.PRSearchResult{
							Items: []github.PRItem{
								{Number: 15, UpdatedAt: "D005"},
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
							Page:    1,
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
					Read(
						pullreq.PRCommitsCacheBucket(langCode),
						pullreq.PRCommitsCacheKey(12),
						gomock.Any(),
					).
					DoAndReturn(storetests.MockReadReturn(
						true,
						cachetypes.PRCommits{
							UpdatedAt: "D001",
							CommitIDs: []string{"C1"},
						},
						nil,
					))

				cacheStore.EXPECT().
					Read(
						pullreq.CommitFilesCacheBucket(langCode),
						pullreq.CommitFilesCacheKey("C1"),
						gomock.Any(),
					).
					DoAndReturn(storetests.MockReadReturn(
						true,
						&github.CommitFiles{
							CommitID: "C1",
							Files:    []string{"content/pl/F1"},
						},
						nil,
					))

				cacheStore.EXPECT().
					Read(
						pullreq.PRCommitsCacheBucket(langCode),
						pullreq.PRCommitsCacheKey(14),
						gomock.Any(),
					).
					DoAndReturn(storetests.MockReadReturn(
						true,
						cachetypes.PRCommits{
							UpdatedAt: "D003",
							CommitIDs: []string{"C2", "C3"},
						},
						nil,
					))

				cacheStore.EXPECT().
					Read(
						pullreq.CommitFilesCacheBucket(langCode),
						pullreq.CommitFilesCacheKey("C2"),
						gomock.Any(),
					).
					DoAndReturn(storetests.MockReadReturn(
						true,
						&github.CommitFiles{
							CommitID: "C2",
							Files:    []string{"content/pl/F2", "content/pl/F3"},
						},
						nil,
					))

				cacheStore.EXPECT().
					Read(
						pullreq.CommitFilesCacheBucket(langCode),
						pullreq.CommitFilesCacheKey("C3"),
						gomock.Any(),
					).
					DoAndReturn(storetests.MockReadReturn(
						true,
						&github.CommitFiles{
							CommitID: "C3",
							Files:    []string{"content/pl/F1", "content/pl/F4"},
						},
						nil,
					))

				cacheStore.EXPECT().
					Read(
						pullreq.PRCommitsCacheBucket(langCode),
						pullreq.PRCommitsCacheKey(15),
						gomock.Any(),
					).
					DoAndReturn(storetests.MockReadReturn(
						true,
						cachetypes.PRCommits{
							UpdatedAt: "D004",
							CommitIDs: []string{"C4"},
						},
						nil,
					))

				gitHubMock.EXPECT().
					GetPRCommits(ctx, 15).
					Return([]string{"C5"}, nil)

				cacheStore.EXPECT().
					Write(
						pullreq.PRCommitsCacheBucket(langCode),
						pullreq.PRCommitsCacheKey(15),
						cachetypes.PRCommits{
							UpdatedAt: "D005",
							CommitIDs: []string{"C5"},
						},
					).
					Return(nil)

				cacheStore.EXPECT().
					Read(
						pullreq.CommitFilesCacheBucket(langCode),
						pullreq.CommitFilesCacheKey("C5"),
						gomock.Any(),
					).
					DoAndReturn(storetests.MockReadNotFound())

				gitHubMock.EXPECT().
					GetCommitFiles(ctx, "C5").
					Return(&github.CommitFiles{Files: []string{"content/pl/F5", "content/pl/F6"}}, nil)

				cacheStore.EXPECT().
					Write(
						pullreq.CommitFilesCacheBucket(langCode),
						pullreq.CommitFilesCacheKey("C5"),
						gomock.Any(),
					).
					Return(nil)

				cacheStore.EXPECT().
					Write(
						pullreq.FilePRsIndexCacheBucket(langCode),
						pullreq.FilePRsIndexCacheKey(langCode),
						pullreq.FilePRIndexData{
							"content/pl/F1": {14, 12},
							"content/pl/F2": {14},
							"content/pl/F3": {14},
							"content/pl/F4": {14},
							"content/pl/F5": {15},
							"content/pl/F6": {15},
						},
					).
					Return(nil)
			},
		},
		{
			name: "filters out files outside content/<langCode>",
			init: func(
				t *testing.T,
				gitHubMock *mocks.MockGitHub,
				cacheStore *mocks.MockCacheStorage,
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
							Page:    1,
							PerPage: 2,
						},
					).
					Return(
						&github.PRSearchResult{
							Items: []github.PRItem{
								{Number: 12, UpdatedAt: "D001"},
							},
							TotalCount: 1,
						},
						nil,
					).Times(1)

				gitHubMock.EXPECT().
					PRSearch(
						ctx,
						github.PRSearchFilter{
							LangCode:    langCode,
							UpdatedFrom: "D001",
							OnlyOpen:    true,
						},
						github.PageRequest{
							Sort:    "updated",
							Order:   "asc",
							Page:    1,
							PerPage: 2,
						},
					).
					Return(
						&github.PRSearchResult{
							Items:      []github.PRItem{},
							TotalCount: 1,
						},
						nil,
					).Times(1)

				cacheStore.EXPECT().
					Read(
						pullreq.PRCommitsCacheBucket(langCode),
						pullreq.PRCommitsCacheKey(12),
						gomock.Any(),
					).
					DoAndReturn(storetests.MockReadNotFound())

				gitHubMock.EXPECT().
					GetPRCommits(ctx, 12).
					Return([]string{"C1"}, nil)

				cacheStore.EXPECT().
					Write(
						pullreq.PRCommitsCacheBucket(langCode),
						pullreq.PRCommitsCacheKey(12),
						cachetypes.PRCommits{
							UpdatedAt: "D001",
							CommitIDs: []string{"C1"},
						},
					).
					Return(nil)

				cacheStore.EXPECT().
					Read(
						pullreq.CommitFilesCacheBucket(langCode),
						pullreq.CommitFilesCacheKey("C1"),
						gomock.Any(),
					).
					DoAndReturn(storetests.MockReadNotFound())

				gitHubMock.EXPECT().
					GetCommitFiles(ctx, "C1").
					Return(
						&github.CommitFiles{
							Files: []string{
								"content/pl/OK.md",
								"content/en/SHOULD_IGNORE.md",
								"README.md",
								"content/pl/sub/file.txt",
							},
						},
						nil,
					)

				cacheStore.EXPECT().
					Write(
						pullreq.CommitFilesCacheBucket(langCode),
						pullreq.CommitFilesCacheKey("C1"),
						gomock.Any(),
					).
					Return(nil)

				cacheStore.EXPECT().
					Write(
						pullreq.FilePRsIndexCacheBucket(langCode),
						pullreq.FilePRsIndexCacheKey(langCode),
						pullreq.FilePRIndexData{
							"content/pl/OK.md":        {12},
							"content/pl/sub/file.txt": {12},
						},
					).
					Return(nil)
			},
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			gitHub := mocks.NewMockGitHub(ctrl)
			cacheStore := mocks.NewMockCacheStorage(ctrl)

			tc.init(t, gitHub, cacheStore)

			filePRIndex := pullreq.NewFilePRIndex(gitHub, cacheStore, 2)

			err := filePRIndex.RefreshIndex(ctx, langCode)
			if err != nil {
				t.Fatal(err)
			}
		})
	}
}
