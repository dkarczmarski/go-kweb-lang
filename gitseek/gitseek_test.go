package gitseek_test

import (
	"context"
	"reflect"
	"testing"

	"go-kweb-lang/git"

	"go-kweb-lang/gitseek"
	"go-kweb-lang/mocks"

	"go.uber.org/mock/gomock"
)

func TestGitSeek_CheckFiles(t *testing.T) {
	for _, tc := range []struct {
		name     string
		initMock func(mock *mocks.MockRepo, ctx context.Context)
		expected []gitseek.FileInfo
	}{
		{
			name: "origin file not exists",
			initMock: func(mock *mocks.MockRepo, ctx context.Context) {
				mock.EXPECT().ListMainBranchCommits(ctx).Return([]git.CommitInfo{
					// simplified commit history
					{
						CommitID: "CID1",
						DateTime: "DT1",
						Comment:  "Comment1",
					},
				}, nil)
				mock.EXPECT().FindFileLastCommit(ctx, "content/pl/path1").Return(
					git.CommitInfo{
						CommitID: "CID1",
						DateTime: "DT1",
						Comment:  "Comment1",
					}, nil)
				mock.EXPECT().FileExists("content/en/path1").Return(false, nil)
				mock.EXPECT().FindFileCommitsAfter(ctx, "content/en/path1", "CID1").
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
					LangForkCommit:   nil,
					OriginFileStatus: "NOT_EXIST",
					OriginUpdates:    nil,
				},
			},
		},
		{
			name: "origin file found with no updates",
			initMock: func(mock *mocks.MockRepo, ctx context.Context) {
				mock.EXPECT().ListMainBranchCommits(ctx).Return([]git.CommitInfo{
					// simplified commit history
					{
						CommitID: "CID1",
						DateTime: "DT1",
						Comment:  "Comment1",
					},
				}, nil)
				mock.EXPECT().FindFileLastCommit(ctx, "content/pl/path1").Return(
					git.CommitInfo{
						CommitID: "CID1",
						DateTime: "DT1",
						Comment:  "Comment1",
					}, nil)
				mock.EXPECT().FileExists("content/en/path1").Return(true, nil)
				mock.EXPECT().FindFileCommitsAfter(ctx, "content/en/path1", "CID1").
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
					LangForkCommit:   nil,
					OriginFileStatus: "",
					OriginUpdates:    nil,
				},
			},
		},
		{
			name: "origin file found with updates",
			initMock: func(mock *mocks.MockRepo, ctx context.Context) {
				mock.EXPECT().ListMainBranchCommits(ctx).Return([]git.CommitInfo{
					// simplified commit history
					{
						CommitID: "CID1",
						DateTime: "DT1",
						Comment:  "Comment1",
					},
					{
						CommitID: "CID4",
						DateTime: "DT4",
						Comment:  "Comment4",
					},
				}, nil)
				mock.EXPECT().FindFileLastCommit(ctx, "content/pl/path1").Return(
					git.CommitInfo{
						CommitID: "CID1",
						DateTime: "DT1",
						Comment:  "Comment1",
					}, nil)
				mock.EXPECT().FileExists("content/en/path1").Return(true, nil)
				mock.EXPECT().FindFileCommitsAfter(ctx, "content/en/path1", "CID1").
					Return([]git.CommitInfo{
						{
							CommitID: "CID2",
							DateTime: "DT2",
							Comment:  "Comment2",
						},
					}, nil)
				mock.EXPECT().ListMergePoints(ctx, "CID2").Return(
					[]git.CommitInfo{
						{
							CommitID: "CID3",
							DateTime: "DT3",
							Comment:  "Comment3",
						},
						{
							CommitID: "CID4",
							DateTime: "DT4",
							Comment:  "Comment4",
						},
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
					OriginFileStatus: "MODIFIED",
					OriginUpdates: []gitseek.OriginUpdate{
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
			name: "origin file found with updates and origin commit made direct to the main branch",
			initMock: func(mock *mocks.MockRepo, ctx context.Context) {
				mock.EXPECT().ListMainBranchCommits(ctx).Return([]git.CommitInfo{
					// simplified commit history
					{
						CommitID: "CID1",
						DateTime: "DT1",
						Comment:  "Comment1",
					},
					{
						CommitID: "CID2",
						DateTime: "DT2",
						Comment:  "Comment2",
					},
				}, nil)
				mock.EXPECT().FindFileLastCommit(ctx, "content/pl/path1").Return(
					git.CommitInfo{
						CommitID: "CID1",
						DateTime: "DT1",
						Comment:  "Comment1",
					}, nil)
				mock.EXPECT().FileExists("content/en/path1").Return(true, nil)
				mock.EXPECT().FindFileCommitsAfter(ctx, "content/en/path1", "CID1").
					Return([]git.CommitInfo{
						{
							CommitID: "CID2",
							DateTime: "DT2",
							Comment:  "Comment2",
						},
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
					OriginFileStatus: "MODIFIED",
					OriginUpdates: []gitseek.OriginUpdate{
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
			initMock: func(mock *mocks.MockRepo, ctx context.Context) {
				mock.EXPECT().ListMainBranchCommits(ctx).Return([]git.CommitInfo{
					// simplified commit history
					{
						CommitID: "CID2",
						DateTime: "DT2",
						Comment:  "Comment2",
					},
					{
						CommitID: "CID0",
						DateTime: "DT0",
						Comment:  "Comment0",
					},
				}, nil)
				mock.EXPECT().FindFileLastCommit(ctx, "content/pl/path1").Return(
					git.CommitInfo{
						CommitID: "CID1",
						DateTime: "DT1",
						Comment:  "Comment1",
					}, nil)
				mock.EXPECT().ListAncestorCommits(ctx, "CID1").Return([]git.CommitInfo{
					// simplified commit history
					{
						CommitID: "CID1",
						DateTime: "DT1",
						Comment:  "Comment1",
					},
					{
						CommitID: "CID0",
						DateTime: "DT0",
						Comment:  "Comment0",
					},
				}, nil)
				mock.EXPECT().FileExists("content/en/path1").Return(true, nil)
				mock.EXPECT().FindFileCommitsAfter(ctx, "content/en/path1", "CID0").
					Return([]git.CommitInfo{
						{
							CommitID: "CID2",
							DateTime: "DT2",
							Comment:  "Comment2",
						},
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
					LangForkCommit: &git.CommitInfo{
						CommitID: "CID0",
						DateTime: "DT0",
						Comment:  "Comment0",
					},
					OriginFileStatus: "MODIFIED",
					OriginUpdates: []gitseek.OriginUpdate{
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
			ctx := context.Background()

			ctrl := gomock.NewController(t)
			mock := mocks.NewMockRepo(ctrl)

			tc.initMock(mock, ctx)

			gitSeek := gitseek.New(mock)

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
