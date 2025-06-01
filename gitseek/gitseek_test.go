package gitseek_test

import (
	"context"
	"reflect"
	"testing"

	"go-kweb-lang/git"
	"go-kweb-lang/gitseek"
	"go-kweb-lang/gitseek/internal/mocks"
	"go-kweb-lang/testing/storetests"

	"go.uber.org/mock/gomock"
)

func TestGitSeek_CheckFiles(t *testing.T) {
	ctx := context.Background()

	for _, tc := range []struct {
		name     string
		initMock func(gitRepo *mocks.MockGitRepo, gitRepoHist *mocks.MockGitRepoHist)
		expected []gitseek.FileInfo
	}{
		{
			name: "EN file not exists",
			initMock: func(gitRepo *mocks.MockGitRepo, gitRepoHist *mocks.MockGitRepoHist) {
				gitRepo.EXPECT().FindFileLastCommit(ctx, "content/pl/path1").Return(
					git.CommitInfo{
						CommitID: "CID1",
						DateTime: "DT1",
						Comment:  "Comment1",
					}, nil)
				gitRepoHist.EXPECT().FindForkCommit(ctx, "CID1").Return(nil, nil)
				gitRepo.EXPECT().FileExists("content/en/path1").Return(false, nil)
				gitRepo.EXPECT().FindFileCommitsAfter(ctx, "content/en/path1", "CID1").
					Return([]git.CommitInfo{}, nil)
			},
			expected: []gitseek.FileInfo{
				{
					LangRelPath: "path1",
					LangLastCommit: git.CommitInfo{
						CommitID: "CID1",
						DateTime: "DT1",
						Comment:  "Comment1",
					},
					LangForkCommit: nil,
					ENFileStatus:   "NOT_EXIST",
					ENUpdates:      nil,
				},
			},
		},
		{
			name: "EN file found with no updates",
			initMock: func(gitRepo *mocks.MockGitRepo, gitRepoHist *mocks.MockGitRepoHist) {
				gitRepo.EXPECT().FindFileLastCommit(ctx, "content/pl/path1").Return(
					git.CommitInfo{
						CommitID: "CID1",
						DateTime: "DT1",
						Comment:  "Comment1",
					}, nil)
				gitRepoHist.EXPECT().FindForkCommit(ctx, "CID1").Return(nil, nil)
				gitRepo.EXPECT().FileExists("content/en/path1").Return(true, nil)
				gitRepo.EXPECT().FindFileCommitsAfter(ctx, "content/en/path1", "CID1").
					Return([]git.CommitInfo{}, nil)
			},
			expected: []gitseek.FileInfo{
				{
					LangRelPath: "path1",
					LangLastCommit: git.CommitInfo{
						CommitID: "CID1",
						DateTime: "DT1",
						Comment:  "Comment1",
					},
					LangForkCommit: nil,
					ENFileStatus:   "",
					ENUpdates:      nil,
				},
			},
		},
		{
			name: "EN file found with updates",
			initMock: func(gitRepo *mocks.MockGitRepo, gitRepoHist *mocks.MockGitRepoHist) {
				gitRepo.EXPECT().FindFileLastCommit(ctx, "content/pl/path1").Return(
					git.CommitInfo{
						CommitID: "CID1",
						DateTime: "DT1",
						Comment:  "Comment1",
					}, nil)
				gitRepoHist.EXPECT().FindForkCommit(ctx, "CID1").Return(nil, nil)
				gitRepo.EXPECT().FileExists("content/en/path1").Return(true, nil)
				gitRepo.EXPECT().FindFileCommitsAfter(ctx, "content/en/path1", "CID1").
					Return([]git.CommitInfo{
						{
							CommitID: "CID2",
							DateTime: "DT2",
							Comment:  "Comment2",
						},
					}, nil)
				gitRepoHist.EXPECT().FindMergeCommit(ctx, "CID2").Return(&git.CommitInfo{
					CommitID: "CID4",
					DateTime: "DT4",
					Comment:  "Comment4",
				}, nil)
			},
			expected: []gitseek.FileInfo{
				{
					LangRelPath: "path1",
					LangLastCommit: git.CommitInfo{
						CommitID: "CID1",
						DateTime: "DT1",
						Comment:  "Comment1",
					},
					LangForkCommit: nil,
					ENFileStatus:   "MODIFIED",
					ENUpdates: []gitseek.ENUpdate{
						{
							Commit: git.CommitInfo{
								CommitID: "CID2",
								DateTime: "DT2",
								Comment:  "Comment2",
							},
							MergePoint: &git.CommitInfo{
								CommitID: "CID4",
								DateTime: "DT4",
								Comment:  "Comment4",
							},
						},
					},
				},
			},
		},
		{
			name: "EN file found with updates and EN commit made direct to the main branch",
			initMock: func(gitRepo *mocks.MockGitRepo, gitRepoHist *mocks.MockGitRepoHist) {
				gitRepo.EXPECT().FindFileLastCommit(ctx, "content/pl/path1").Return(
					git.CommitInfo{
						CommitID: "CID1",
						DateTime: "DT1",
						Comment:  "Comment1",
					}, nil)
				gitRepo.EXPECT().FileExists("content/en/path1").Return(true, nil)
				gitRepo.EXPECT().FindFileCommitsAfter(ctx, "content/en/path1", "CID1").
					Return([]git.CommitInfo{
						{
							CommitID: "CID2",
							DateTime: "DT2",
							Comment:  "Comment2",
						},
					}, nil)
				gitRepoHist.EXPECT().FindForkCommit(ctx, "CID1").Return(nil, nil)
				gitRepoHist.EXPECT().FindMergeCommit(ctx, "CID2").Return(nil, nil)
			},
			expected: []gitseek.FileInfo{
				{
					LangRelPath: "path1",
					LangLastCommit: git.CommitInfo{
						CommitID: "CID1",
						DateTime: "DT1",
						Comment:  "Comment1",
					},
					LangForkCommit: nil,
					ENFileStatus:   "MODIFIED",
					ENUpdates: []gitseek.ENUpdate{
						{
							Commit: git.CommitInfo{
								CommitID: "CID2",
								DateTime: "DT2",
								Comment:  "Comment2",
							},
							MergePoint: nil,
						},
					},
				},
			},
		},
		{
			name: "lang commit is in a separate branch",
			initMock: func(gitRepo *mocks.MockGitRepo, gitRepoHist *mocks.MockGitRepoHist) {
				gitRepo.EXPECT().FindFileLastCommit(ctx, "content/pl/path1").Return(
					git.CommitInfo{
						CommitID: "CID1",
						DateTime: "DT1",
						Comment:  "Comment1",
					}, nil)

				gitRepo.EXPECT().FileExists("content/en/path1").Return(true, nil)
				gitRepo.EXPECT().FindFileCommitsAfter(ctx, "content/en/path1", "CID0").
					Return([]git.CommitInfo{
						{
							CommitID: "CID2",
							DateTime: "DT2",
							Comment:  "Comment2",
						},
					}, nil)
				gitRepoHist.EXPECT().FindForkCommit(ctx, "CID1").Return(&git.CommitInfo{
					CommitID: "CID0",
					DateTime: "DT0",
					Comment:  "Comment0",
				}, nil)
				gitRepoHist.EXPECT().FindMergeCommit(ctx, "CID2").Return(nil, nil)
			},
			expected: []gitseek.FileInfo{
				{
					LangRelPath: "path1",
					LangLastCommit: git.CommitInfo{
						CommitID: "CID1",
						DateTime: "DT1",
						Comment:  "Comment1",
					},
					LangForkCommit: &git.CommitInfo{
						CommitID: "CID0",
						DateTime: "DT0",
						Comment:  "Comment0",
					},
					ENFileStatus: "MODIFIED",
					ENUpdates: []gitseek.ENUpdate{
						{
							Commit: git.CommitInfo{
								CommitID: "CID2",
								DateTime: "DT2",
								Comment:  "Comment2",
							},
							MergePoint: nil,
						},
					},
				},
			},
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			gitRepo := mocks.NewMockGitRepo(ctrl)
			gitRepoHist := mocks.NewMockGitRepoHist(ctrl)
			cacheStore := mocks.NewMockCacheStore(ctrl)

			cacheStore.EXPECT().
				Read(gomock.Any(), gomock.Any(), gomock.Any()).
				DoAndReturn(storetests.MockReadNotFound())

			cacheStore.EXPECT().
				Write(gomock.Any(), gomock.Any(), gomock.Any()).Return(nil)

			tc.initMock(gitRepo, gitRepoHist)

			gitSeek := gitseek.New(gitRepo, gitRepoHist, cacheStore)

			fileInfos, err := gitSeek.CheckFiles(ctx, []string{"path1"}, "pl")
			if err != nil {
				t.Errorf("unexpected error: %v", err)
			}

			if !reflect.DeepEqual(tc.expected, fileInfos) {
				t.Errorf("unexpected result\nexpected: %+v\nactual  : %+v", tc.expected, fileInfos)
			}
		})
	}
}

func TestGitSeek_InvalidateFile(t *testing.T) {
	for _, tc := range []struct {
		name     string
		file     string
		initMock func(gitRepo *mocks.MockGitRepo, gitRepoHist *mocks.MockGitRepoHist, cacheStore *mocks.MockCacheStore)
	}{
		{
			name: "invalidate file outside the 'content' dir",
			file: "dir/file1",
			initMock: func(gitRepo *mocks.MockGitRepo, gitRepoHist *mocks.MockGitRepoHist, cacheStore *mocks.MockCacheStore) {
			},
		},
		{
			name: "invalidate lang file",
			file: "content/pl/file1",
			initMock: func(gitRepo *mocks.MockGitRepo, gitRepoHist *mocks.MockGitRepoHist, cacheStore *mocks.MockCacheStore) {
				cacheStore.EXPECT().Delete("lang/pl/git-file-info", "file1").Return(nil)
			},
		},
		{
			name: "invalidate EN file",
			file: "content/en/file1",
			initMock: func(gitRepo *mocks.MockGitRepo, gitRepoHist *mocks.MockGitRepoHist, cacheStore *mocks.MockCacheStore) {
				cacheStore.EXPECT().ListBuckets("lang").Return([]string{"pl", "fr"}, nil)

				cacheStore.EXPECT().Delete("lang/en/git-file-info", "file1").Return(nil)
				cacheStore.EXPECT().Delete("lang/pl/git-file-info", "file1").Return(nil)
				cacheStore.EXPECT().Delete("lang/fr/git-file-info", "file1").Return(nil)
			},
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			gitRepo := mocks.NewMockGitRepo(ctrl)
			gitRepoHist := mocks.NewMockGitRepoHist(ctrl)
			cacheStore := mocks.NewMockCacheStore(ctrl)

			tc.initMock(gitRepo, gitRepoHist, cacheStore)

			gitSeek := gitseek.New(gitRepo, gitRepoHist, cacheStore)

			if err := gitSeek.InvalidateFile(tc.file); err != nil {
				t.Fatal(err)
			}
		})
	}
}
