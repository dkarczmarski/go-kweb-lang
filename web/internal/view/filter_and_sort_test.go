package view_test

import (
	"encoding/json"
	"testing"

	"go-kweb-lang/git"
	"go-kweb-lang/web/internal/view"
)

func TestFilterAndSort(t *testing.T) {
	// sort SortByStatus
	// sort SortByEnUpdate

	for _, tc := range []struct {
		name      string
		files     []view.FileModel
		itemsType int
		fileName  string
		filePath  string
		sort      int
		sortOrder int
		expected  []view.FileModel
	}{
		{
			name: "when empty",
		},
		{
			name: "filter with filename",
			files: []view.FileModel{
				{
					LangRelPath: view.LinkModel{Text: "a1"},
				},
				{
					LangRelPath: view.LinkModel{Text: "a2"},
				},
				{
					LangRelPath: view.LinkModel{Text: "a3"},
				},
			},
			fileName: "a2",
			expected: []view.FileModel{
				{
					LangRelPath: view.LinkModel{Text: "a2"},
				},
			},
		},
		{
			name: "filter with filepath",
			files: []view.FileModel{
				{
					LangRelPath: view.LinkModel{Text: "a/b/c1"},
				},
				{
					LangRelPath: view.LinkModel{Text: "a/x"},
				},
				{
					LangRelPath: view.LinkModel{Text: "a/b/c2"},
				},
			},
			filePath: "/b",
			expected: []view.FileModel{
				{
					LangRelPath: view.LinkModel{Text: "a/b/c1"},
				},
				{
					LangRelPath: view.LinkModel{Text: "a/b/c2"},
				},
			},
		},
		{
			name: "filter when items type is ItemsAll",
			files: []view.FileModel{
				{
					LangRelPath: view.LinkModel{Text: "a1"},
				},
				{
					LangRelPath: view.LinkModel{Text: "a2"},
				},
				{
					LangRelPath: view.LinkModel{Text: "a3"},
				},
			},
			itemsType: view.ItemsAll,
			expected: []view.FileModel{
				{
					LangRelPath: view.LinkModel{Text: "a1"},
				},
				{
					LangRelPath: view.LinkModel{Text: "a2"},
				},
				{
					LangRelPath: view.LinkModel{Text: "a3"},
				},
			},
		},
		{
			name: "filter when items type is ItemsWithUpdate",
			files: []view.FileModel{
				{
					LangRelPath: view.LinkModel{Text: "a1"},
				},
				{
					LangRelPath: view.LinkModel{Text: "a2"},
					ENUpdates: view.ENUpdateGroups{
						AfterLastCommit: []view.ENUpdate{
							{
								Commit: view.CommitLinkModel{
									CommitInfo: git.CommitInfo{
										DateTime: "DT10",
									},
								},
							},
						},
					},
				},
				{
					LangRelPath: view.LinkModel{Text: "a3"},
					PRs: []view.LinkModel{
						{Text: "1001"},
					},
				},
				{
					LangRelPath: view.LinkModel{Text: "a4"},
					ENUpdates: view.ENUpdateGroups{
						AfterLastCommit: []view.ENUpdate{
							{
								Commit: view.CommitLinkModel{
									CommitInfo: git.CommitInfo{
										DateTime: "DT10",
									},
								},
							},
						},
					},
					PRs: []view.LinkModel{
						{Text: "1009"},
					},
				},
			},
			itemsType: view.ItemsWithUpdate,
			expected: []view.FileModel{
				{
					LangRelPath: view.LinkModel{Text: "a2"},
					ENUpdates: view.ENUpdateGroups{
						AfterLastCommit: []view.ENUpdate{
							{
								Commit: view.CommitLinkModel{
									CommitInfo: git.CommitInfo{
										DateTime: "DT10",
									},
								},
							},
						},
					},
				},
				{
					LangRelPath: view.LinkModel{Text: "a4"},
					ENUpdates: view.ENUpdateGroups{
						AfterLastCommit: []view.ENUpdate{
							{
								Commit: view.CommitLinkModel{
									CommitInfo: git.CommitInfo{
										DateTime: "DT10",
									},
								},
							},
						},
					},
					PRs: []view.LinkModel{
						{Text: "1009"},
					},
				},
			},
		},
		{
			name: "filter when items type is ItemsWithUpdateOrPR",
			files: []view.FileModel{
				{
					LangRelPath: view.LinkModel{Text: "a1"},
				},
				{
					LangRelPath: view.LinkModel{Text: "a2"},
					ENUpdates: view.ENUpdateGroups{
						AfterLastCommit: []view.ENUpdate{
							{
								Commit: view.CommitLinkModel{
									CommitInfo: git.CommitInfo{
										DateTime: "DT10",
									},
								},
							},
						},
					},
				},
				{
					LangRelPath: view.LinkModel{Text: "a3"},
					PRs: []view.LinkModel{
						{Text: "1001"},
					},
				},
				{
					LangRelPath: view.LinkModel{Text: "a4"},
					ENUpdates: view.ENUpdateGroups{
						AfterLastCommit: []view.ENUpdate{
							{
								Commit: view.CommitLinkModel{
									CommitInfo: git.CommitInfo{
										DateTime: "DT10",
									},
								},
							},
						},
					},
					PRs: []view.LinkModel{
						{Text: "1009"},
					},
				},
			},
			itemsType: view.ItemsWithUpdateOrPR,
			expected: []view.FileModel{
				{
					LangRelPath: view.LinkModel{Text: "a2"},
					ENUpdates: view.ENUpdateGroups{
						AfterLastCommit: []view.ENUpdate{
							{
								Commit: view.CommitLinkModel{
									CommitInfo: git.CommitInfo{
										DateTime: "DT10",
									},
								},
							},
						},
					},
				},
				{
					LangRelPath: view.LinkModel{Text: "a3"},
					PRs: []view.LinkModel{
						{Text: "1001"},
					},
				},
				{
					LangRelPath: view.LinkModel{Text: "a4"},
					ENUpdates: view.ENUpdateGroups{
						AfterLastCommit: []view.ENUpdate{
							{
								Commit: view.CommitLinkModel{
									CommitInfo: git.CommitInfo{
										DateTime: "DT10",
									},
								},
							},
						},
					},
					PRs: []view.LinkModel{
						{Text: "1009"},
					},
				},
			},
		},
		{
			name: "filter when items type is ItemsWithPR",
			files: []view.FileModel{
				{
					LangRelPath: view.LinkModel{Text: "a1"},
				},
				{
					LangRelPath: view.LinkModel{Text: "a2"},
					ENUpdates: view.ENUpdateGroups{
						AfterLastCommit: []view.ENUpdate{
							{
								Commit: view.CommitLinkModel{
									CommitInfo: git.CommitInfo{
										DateTime: "DT10",
									},
								},
							},
						},
					},
				},
				{
					LangRelPath: view.LinkModel{Text: "a3"},
					PRs: []view.LinkModel{
						{Text: "1001"},
					},
				},
				{
					LangRelPath: view.LinkModel{Text: "a4"},
					ENUpdates: view.ENUpdateGroups{
						AfterLastCommit: []view.ENUpdate{
							{
								Commit: view.CommitLinkModel{
									CommitInfo: git.CommitInfo{
										DateTime: "DT10",
									},
								},
							},
						},
					},
					PRs: []view.LinkModel{
						{Text: "1009"},
					},
				},
			},
			itemsType: view.ItemsWithPR,
			expected: []view.FileModel{
				{
					LangRelPath: view.LinkModel{Text: "a3"},
					PRs: []view.LinkModel{
						{Text: "1001"},
					},
				},
				{
					LangRelPath: view.LinkModel{Text: "a4"},
					ENUpdates: view.ENUpdateGroups{
						AfterLastCommit: []view.ENUpdate{
							{
								Commit: view.CommitLinkModel{
									CommitInfo: git.CommitInfo{
										DateTime: "DT10",
									},
								},
							},
						},
					},
					PRs: []view.LinkModel{
						{Text: "1009"},
					},
				},
			},
		},
		{
			name: "filter when sort is SortByFileName",
			files: []view.FileModel{
				{
					LangRelPath: view.LinkModel{Text: "a2"},
				},
				{
					LangRelPath: view.LinkModel{Text: "a3"},
				},
				{
					LangRelPath: view.LinkModel{Text: "a1"},
				},
			},
			sort: view.SortByFileName,
			expected: []view.FileModel{
				{
					LangRelPath: view.LinkModel{Text: "a1"},
				},
				{
					LangRelPath: view.LinkModel{Text: "a2"},
				},
				{
					LangRelPath: view.LinkModel{Text: "a3"},
				},
			},
		},
		{
			name: "filter when sort is SortByFileName and SortOrderDesc",
			files: []view.FileModel{
				{
					LangRelPath: view.LinkModel{Text: "a2"},
				},
				{
					LangRelPath: view.LinkModel{Text: "a3"},
				},
				{
					LangRelPath: view.LinkModel{Text: "a1"},
				},
			},
			sort:      view.SortByFileName,
			sortOrder: view.SortOrderDesc,
			expected: []view.FileModel{
				{
					LangRelPath: view.LinkModel{Text: "a3"},
				},
				{
					LangRelPath: view.LinkModel{Text: "a2"},
				},
				{
					LangRelPath: view.LinkModel{Text: "a1"},
				},
			},
		},
		{
			name: "filter when sort is SortByStatus",
			files: []view.FileModel{
				{
					LangRelPath: view.LinkModel{Text: "a1"},
					ENStatus:    "C",
				},
				{
					LangRelPath: view.LinkModel{Text: "a2"},
					ENStatus:    "A",
				},
				{
					LangRelPath: view.LinkModel{Text: "a3"},
					ENStatus:    "B",
				},
			},
			sort: view.SortByStatus,
			expected: []view.FileModel{
				{
					LangRelPath: view.LinkModel{Text: "a2"},
					ENStatus:    "A",
				},
				{
					LangRelPath: view.LinkModel{Text: "a3"},
					ENStatus:    "B",
				},
				{
					LangRelPath: view.LinkModel{Text: "a1"},
					ENStatus:    "C",
				},
			},
		},
		{
			name: "filter when sort is SortByEnUpdate",
			files: []view.FileModel{
				{
					LangRelPath: view.LinkModel{Text: "a1"},
					ENUpdates: view.ENUpdateGroups{
						LastCommit: git.CommitInfo{DateTime: "DT-30"},
					},
				},
				{
					LangRelPath: view.LinkModel{Text: "a2"},
					ENUpdates: view.ENUpdateGroups{
						LastCommit: git.CommitInfo{DateTime: "DT-10"},
					},
				},
				{
					LangRelPath: view.LinkModel{Text: "a3"},
					ENUpdates: view.ENUpdateGroups{
						LastCommit: git.CommitInfo{DateTime: "DT-20"},
					},
				},
			},
			sort: view.SortByEnUpdate,
			expected: []view.FileModel{
				{
					LangRelPath: view.LinkModel{Text: "a2"},
					ENUpdates: view.ENUpdateGroups{
						LastCommit: git.CommitInfo{DateTime: "DT-10"},
					},
				},
				{
					LangRelPath: view.LinkModel{Text: "a3"},
					ENUpdates: view.ENUpdateGroups{
						LastCommit: git.CommitInfo{DateTime: "DT-20"},
					},
				},
				{
					LangRelPath: view.LinkModel{Text: "a1"},
					ENUpdates: view.ENUpdateGroups{
						LastCommit: git.CommitInfo{DateTime: "DT-30"},
					},
				},
			},
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			actual := view.FilterAndSort(tc.files, tc.itemsType, tc.fileName, tc.filePath, tc.sort, tc.sortOrder)

			actualJSON := toJSONString(t, actual)
			expectedJSON := toJSONString(t, tc.expected)

			if actualJSON != expectedJSON {
				t.Errorf("actual:\n%s\nexpected:%s", actualJSON, expectedJSON)
			}
		})
	}
}

func toJSONString(t testing.TB, v any) string {
	t.Helper()

	b, err := json.MarshalIndent(v, "", "\t")
	if err != nil {
		t.Fatal(err)
	}

	return string(b)
}
