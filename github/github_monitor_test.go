package github_test

import (
	"context"
	"testing"

	"go-kweb-lang/github"
	"go-kweb-lang/mocks"

	"go.uber.org/mock/gomock"
)

func TestMonitor_Check(t *testing.T) {
	for _, tc := range []struct {
		name      string
		initMocks func(ctx context.Context,
			githubMock *mocks.MockGitHub,
			langMock *mocks.MockLangProvider,
			storageMock *mocks.MockMonitorStorage,
			task *mocks.MockMonitorTask,
		)
		checkErr func(err error) bool
	}{
		{
			name: "first run (cache is empty)",
			initMocks: func(ctx context.Context,
				githubMock *mocks.MockGitHub,
				langMock *mocks.MockLangProvider,
				storageMock *mocks.MockMonitorStorage,
				task *mocks.MockMonitorTask,
			) {
				storageMock.EXPECT().ReadLastRepoUpdatedAt().Return("", nil)
				githubMock.EXPECT().GetLatestCommit(ctx).
					Return(&github.CommitInfo{CommitID: "C1", DateTime: "DT1"}, nil)
				storageMock.EXPECT().WriteLastRepoUpdatedAt("DT1").Return(nil)

				langMock.EXPECT().LangCodes().Return([]string{"pl"}, nil)

				storageMock.EXPECT().ReadLastPRUpdatedAt("pl").Return("", nil)
				githubMock.EXPECT().PRSearch(ctx, github.PRSearchFilter{LangCode: "pl"}, gomock.Any()).
					Return(&github.PRSearchResult{Items: []github.PRItem{{Number: 1, UpdatedAt: "PL-U1"}}}, nil)
				storageMock.EXPECT().WriteLastPRUpdatedAt("pl", "PL-U1")

				task.EXPECT().OnUpdate(ctx, true, []string{"pl"}).Return(nil)
			},
			checkErr: func(err error) bool {
				return err == nil
			},
		},
		{
			name: "when cache contains values but there are no updates",
			initMocks: func(ctx context.Context,
				githubMock *mocks.MockGitHub,
				langMock *mocks.MockLangProvider,
				storageMock *mocks.MockMonitorStorage,
				task *mocks.MockMonitorTask,
			) {
				storageMock.EXPECT().ReadLastRepoUpdatedAt().Return("DT0", nil)
				githubMock.EXPECT().GetLatestCommit(ctx).
					Return(&github.CommitInfo{CommitID: "C1", DateTime: "DT0"}, nil)

				langMock.EXPECT().LangCodes().Return([]string{"pl"}, nil)

				storageMock.EXPECT().ReadLastPRUpdatedAt("pl").Return("PL-U1", nil)
				githubMock.EXPECT().PRSearch(ctx, github.PRSearchFilter{LangCode: "pl"}, gomock.Any()).
					Return(&github.PRSearchResult{Items: []github.PRItem{{Number: 1, UpdatedAt: "PL-U1"}}}, nil)
			},
			checkErr: func(err error) bool {
				return err == nil
			},
		},
		{
			name: "when cache contains values and there is repo update",
			initMocks: func(ctx context.Context,
				githubMock *mocks.MockGitHub,
				langMock *mocks.MockLangProvider,
				storageMock *mocks.MockMonitorStorage,
				task *mocks.MockMonitorTask,
			) {
				storageMock.EXPECT().ReadLastRepoUpdatedAt().Return("DT0", nil)
				githubMock.EXPECT().GetLatestCommit(ctx).
					Return(&github.CommitInfo{CommitID: "C1", DateTime: "DT1"}, nil)
				storageMock.EXPECT().WriteLastRepoUpdatedAt("DT1").Return(nil)

				langMock.EXPECT().LangCodes().Return([]string{"pl"}, nil)

				storageMock.EXPECT().ReadLastPRUpdatedAt("pl").Return("PL-U1", nil)
				githubMock.EXPECT().PRSearch(ctx, github.PRSearchFilter{LangCode: "pl"}, gomock.Any()).
					Return(&github.PRSearchResult{Items: []github.PRItem{{Number: 1, UpdatedAt: "PL-U1"}}}, nil)

				task.EXPECT().OnUpdate(ctx, true, nil).Return(nil)
			},
			checkErr: func(err error) bool {
				return err == nil
			},
		},
		{
			name: "when cache contains values and there is PR update",
			initMocks: func(ctx context.Context,
				githubMock *mocks.MockGitHub,
				langMock *mocks.MockLangProvider,
				storageMock *mocks.MockMonitorStorage,
				task *mocks.MockMonitorTask,
			) {
				storageMock.EXPECT().ReadLastRepoUpdatedAt().Return("DT0", nil)
				githubMock.EXPECT().GetLatestCommit(ctx).
					Return(&github.CommitInfo{CommitID: "C1", DateTime: "DT0"}, nil)

				langMock.EXPECT().LangCodes().Return([]string{"pl"}, nil)

				storageMock.EXPECT().ReadLastPRUpdatedAt("pl").Return("PL-U1", nil)
				githubMock.EXPECT().PRSearch(ctx, github.PRSearchFilter{LangCode: "pl"}, gomock.Any()).
					Return(&github.PRSearchResult{Items: []github.PRItem{{Number: 1, UpdatedAt: "PL-U2"}}}, nil)
				storageMock.EXPECT().WriteLastPRUpdatedAt("pl", "PL-U2")

				task.EXPECT().OnUpdate(ctx, false, []string{"pl"}).Return(nil)
			},
			checkErr: func(err error) bool {
				return err == nil
			},
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			ctx := context.Background()

			ctrl := gomock.NewController(t)

			githubMock := mocks.NewMockGitHub(ctrl)
			langMock := mocks.NewMockLangProvider(ctrl)
			storageMock := mocks.NewMockMonitorStorage(ctrl)
			task := mocks.NewMockMonitorTask(ctrl)

			tc.initMocks(ctx, githubMock, langMock, storageMock, task)

			mon := github.NewMonitor(githubMock, langMock, storageMock)

			err := mon.Check(ctx, task)

			if !tc.checkErr(err) {
				t.Errorf("unexpected error result: %v", err)
			}
		})
	}
}
