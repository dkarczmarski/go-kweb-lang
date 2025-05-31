package github_test

import (
	"context"
	_ "embed"
	"net/http"
	"net/http/httptest"
	"net/url"
	"reflect"
	"testing"

	"go-kweb-lang/github"
)

//go:embed testdata/TestGitHub_GetCommitFiles.txt
var GetCommitFiles []byte

func TestGitHub_GetCommitFiles_Integration(t *testing.T) {
	for _, tc := range []struct {
		name           string
		commitID       string
		response       []byte
		expectedURL    string
		expectedResult *github.CommitFiles
	}{
		{
			commitID:    "f9ef60a9cf2ce7fdc4e242c292d8ed728deab912",
			response:    GetCommitFiles,
			expectedURL: "/repos/kubernetes/website/commits/f9ef60a9cf2ce7fdc4e242c292d8ed728deab912",
			expectedResult: &github.CommitFiles{
				CommitID: "f9ef60a9cf2ce7fdc4e242c292d8ed728deab912",
				Files: []string{
					"layouts/index.html",
					"layouts/shortcodes/site-searchbar.html",
				},
			},
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			ctx := context.Background()

			mockServer := newMockServer(t, tc.expectedURL, url.Values{}, tc.response)
			defer mockServer.Close()

			gh := &github.GitHub{
				HTTPClient: mockServer.Client(),
				BaseURL:    mockServer.URL,
			}

			actualResult, err := gh.GetCommitFiles(ctx, tc.commitID)
			if err != nil {
				t.Fatal(err)
			}

			if !reflect.DeepEqual(tc.expectedResult, actualResult) {
				t.Errorf("result error\nexpected : %+v\nactual   : %+v", tc.expectedResult, actualResult)
			}
		})
	}
}

//go:embed testdata/TestGitHub_GetPRCommits.txt
var GetPRCommits []byte

func TestGitHub_GetPRCommits_Integration(t *testing.T) {
	for _, tc := range []struct {
		name           string
		prNumber       int
		response       []byte
		expectedURL    string
		expectedResult []string
	}{
		{
			prNumber:       42,
			response:       GetPRCommits,
			expectedURL:    "/repos/kubernetes/website/pulls/42/commits",
			expectedResult: []string{"5bac466fc45325e2f5cfa63d06b9f2032ecba712"},
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			ctx := context.Background()

			mockServer := newMockServer(t, tc.expectedURL, url.Values{}, tc.response)
			defer mockServer.Close()

			gh := &github.GitHub{
				HTTPClient: mockServer.Client(),
				BaseURL:    mockServer.URL,
			}

			actualResult, err := gh.GetPRCommits(ctx, tc.prNumber)
			if err != nil {
				t.Fatal(err)
			}

			if !reflect.DeepEqual(tc.expectedResult, actualResult) {
				t.Errorf("result error\nexpected : %+v\nactual   : %+v", tc.expectedResult, actualResult)
			}
		})
	}
}

//go:embed testdata/TestGitHub_PRSearch.txt
var PRSearch []byte

func TestGitHub_PRSearch_Integration(t *testing.T) {
	for _, tc := range []struct {
		name                string
		filter              github.PRSearchFilter
		page                github.PageRequest
		response            []byte
		expectedURL         string
		expectedQueryParams url.Values
		expectedResult      *github.PRSearchResult
	}{
		{
			filter: github.PRSearchFilter{
				LangCode:    "pl",
				UpdatedFrom: "",
				OnlyOpen:    true,
			},
			page: github.PageRequest{
				Sort:    "updated",
				Order:   "asc",
				PerPage: 4,
			},
			response:    PRSearch,
			expectedURL: "/search/issues",
			expectedQueryParams: url.Values{
				"q":        []string{"repo:kubernetes/website is:pr state:open label:language/pl"},
				"page":     []string{"1"},
				"per_page": []string{"4"},
				"order":    []string{"asc"},
				"sort":     []string{"updated"},
			},
			expectedResult: &github.PRSearchResult{
				Items: []github.PRItem{
					{49640, "2025-02-04T14:53:37Z"},
					{49669, "2025-02-07T07:18:42Z"},
					{49639, "2025-02-10T07:40:28Z"},
					{49633, "2025-02-10T07:47:50Z"},
				},
				TotalCount: 49,
			},
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			ctx := context.Background()

			mockServer := newMockServer(t, tc.expectedURL, tc.expectedQueryParams, tc.response)
			defer mockServer.Close()

			gh := &github.GitHub{
				HTTPClient: mockServer.Client(),
				BaseURL:    mockServer.URL,
			}

			actualResult, err := gh.PRSearch(ctx, tc.filter, tc.page)
			if err != nil {
				t.Fatal(err)
			}

			if !reflect.DeepEqual(tc.expectedResult, actualResult) {
				t.Errorf("result error\nexpected : %+v\nactual   : %+v", tc.expectedResult, actualResult)
			}
		})
	}
}

func newMockServer(t *testing.T, expectedURL string, expectedQueryParams url.Values, response []byte) *httptest.Server {
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != expectedURL {
			t.Errorf("result URL\nexpected : %+v\nactual   : %+v", expectedURL, r.URL.Path)
		}

		expectedQuery := expectedQueryParams
		actualQuery := r.URL.Query()

		for key, expectedValues := range expectedQuery {
			actualValues, ok := actualQuery[key]
			if !ok {
				t.Errorf("missing query param: %s", key)
				continue
			}

			if expectedValues[0] != actualValues[0] {
				t.Errorf("query param %s\nexpected: %s\nactual   : %s", key, expectedValues[0], actualValues[0])
			}
		}

		for key := range actualQuery {
			_, ok := expectedQuery[key]
			if !ok {
				t.Errorf("unexptected query param: %s", key)
				continue
			}
		}

		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write(response)
	}))
}
