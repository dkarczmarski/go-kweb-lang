package pullreq_test

import (
	"context"
	"reflect"
	"testing"

	"go-kweb-lang/github"
	"go-kweb-lang/mocks"
	"go-kweb-lang/pullreq"

	"go.uber.org/mock/gomock"
)

func TestFilePRFinder_ListPRs(t *testing.T) {
}

func TestFilePRFinder_Update(t *testing.T) {
	langCode := "pl"

	ctrl := gomock.NewController(t)
	mock := mocks.NewMockGitHub(ctrl)

	mock.EXPECT().
		PRSearch(
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

	mock.EXPECT().
		PRSearch(
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

	mock.EXPECT().
		PRSearch(
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

	mock.EXPECT().GetPRCommits(12).Return([]string{"C1"}, nil)
	mock.EXPECT().GetPRCommits(14).Return([]string{"C2", "C3"}, nil)
	mock.EXPECT().GetPRCommits(15).Return([]string{"C4"}, nil)

	mock.EXPECT().GetCommitFiles("C1").Return(&github.CommitFiles{Files: []string{"F1"}}, nil)
	mock.EXPECT().GetCommitFiles("C2").Return(&github.CommitFiles{Files: []string{"F2", "F3"}}, nil)
	mock.EXPECT().GetCommitFiles("C3").Return(&github.CommitFiles{Files: []string{"F1", "F4"}}, nil)
	mock.EXPECT().GetCommitFiles("C4").Return(&github.CommitFiles{Files: []string{"F5"}}, nil)

	filePRFinder := pullreq.NewFilePRFinder(
		mock,
		t.TempDir(),
		func(config *pullreq.FilePRFinderConfig) {
			config.PerPage = 2
		},
	)

	err := filePRFinder.Update(context.Background(), langCode)
	if err != nil {
		t.Fatal(err)
	}

	if numbers, err := filePRFinder.ListPR("F1"); err != nil || !reflect.DeepEqual(numbers, []int{14, 12}) {
		t.Error(numbers, err)
	}
	if numbers, err := filePRFinder.ListPR("F2"); err != nil || !reflect.DeepEqual(numbers, []int{14}) {
		t.Error(numbers, err)
	}
	if numbers, err := filePRFinder.ListPR("F3"); err != nil || !reflect.DeepEqual(numbers, []int{14}) {
		t.Error(numbers, err)
	}
	if numbers, err := filePRFinder.ListPR("F4"); err != nil || !reflect.DeepEqual(numbers, []int{14}) {
		t.Error(numbers, err)
	}
	if numbers, err := filePRFinder.ListPR("F5"); err != nil || !reflect.DeepEqual(numbers, []int{15}) {
		t.Error(numbers, err)
	}
}
