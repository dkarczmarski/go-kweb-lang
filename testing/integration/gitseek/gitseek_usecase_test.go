//nolint:paralleltest
package gitseek_test

import (
	"testing"

	"github.com/dkarczmarski/go-kweb-lang/git"
	"github.com/dkarczmarski/go-kweb-lang/gitseek"
)

func TestGitSeek_CheckLang_UseCase_EnUnchangedAfterLang(t *testing.T) {
	ctx := t.Context()
	env := newIntegrationEnv(t, "en_unchanged_after_lang")

	result, err := env.gitSeeker.CheckLang(ctx, "pl", env.pair)
	if err != nil {
		t.Fatalf("CheckLang returned error: %v", err)
	}

	expected := gitseek.FileInfo{
		LangPath: "content/pl/docs/test.md",
		LangLastCommit: git.CommitInfo{
			CommitID: "4c0887027d3bbcf5b31519ea9cdb8da01c1c9c31",
			DateTime: "2020-01-03T00:00:00+00:00",
			Comment:  "B: add content/pl/docs/test.md",
		},
		LangMergeCommit: nil,
		LangForkCommit:  nil,
		FileStatus:      gitseek.StatusLangFileUpToDate,
		EnUpdates:       nil,
	}

	assertEqualFileInfo(t, expected, result)
}

func TestGitSeek_CheckLang_UseCase_EnUpdatedOnMainAfterLang(t *testing.T) {
	ctx := t.Context()
	env := newIntegrationEnv(t, "en_updated_on_main_after_lang")

	result, err := env.gitSeeker.CheckLang(ctx, "pl", env.pair)
	if err != nil {
		t.Fatalf("CheckLang returned error: %v", err)
	}

	expected := gitseek.FileInfo{
		LangPath: "content/pl/docs/test.md",
		LangLastCommit: git.CommitInfo{
			CommitID: "4c0887027d3bbcf5b31519ea9cdb8da01c1c9c31",
			DateTime: "2020-01-03T00:00:00+00:00",
			Comment:  "B: add content/pl/docs/test.md",
		},
		LangMergeCommit: nil,
		LangForkCommit:  nil,
		FileStatus:      gitseek.StatusEnFileUpdated,
		EnUpdates: []gitseek.EnUpdate{
			{
				Commit: git.CommitInfo{
					CommitID: "2326e9caadb1bd8437bdfc09934fe2a73168c5f4",
					DateTime: "2020-01-04T00:00:00+00:00",
					Comment:  "C: update content/en/docs/test.md",
				},
				MergePoint: nil,
			},
		},
	}

	assertEqualFileInfo(t, expected, result)
}

func TestGitSeek_CheckLang_UseCase_MultipleEnUpdatesAfterLang(t *testing.T) {
	ctx := t.Context()
	env := newIntegrationEnv(t, "multiple_en_updates_after_lang")

	result, err := env.gitSeeker.CheckLang(ctx, "pl", env.pair)
	if err != nil {
		t.Fatalf("CheckLang returned error: %v", err)
	}

	expected := gitseek.FileInfo{
		LangPath: "content/pl/docs/test.md",
		LangLastCommit: git.CommitInfo{
			CommitID: "4c0887027d3bbcf5b31519ea9cdb8da01c1c9c31",
			DateTime: "2020-01-03T00:00:00+00:00",
			Comment:  "B: add content/pl/docs/test.md",
		},
		LangMergeCommit: nil,
		LangForkCommit:  nil,
		FileStatus:      gitseek.StatusEnFileUpdated,
		EnUpdates: []gitseek.EnUpdate{
			{
				Commit: git.CommitInfo{
					CommitID: "0307495485146d0bcd428e30c6fc07354380ebe5",
					DateTime: "2020-01-05T00:00:00+00:00",
					Comment:  "D: update content/en/docs/test.md again",
				},
				MergePoint: nil,
			},
			{
				Commit: git.CommitInfo{
					CommitID: "2326e9caadb1bd8437bdfc09934fe2a73168c5f4",
					DateTime: "2020-01-04T00:00:00+00:00",
					Comment:  "C: update content/en/docs/test.md",
				},
				MergePoint: nil,
			},
		},
	}

	assertEqualFileInfo(t, expected, result)
}

func TestGitSeek_CheckLang_UseCase_EnFileNoLongerExists(t *testing.T) {
	ctx := t.Context()
	env := newIntegrationEnv(t, "en_file_no_longer_exists")

	result, err := env.gitSeeker.CheckLang(ctx, "pl", env.pair)
	if err != nil {
		t.Fatalf("CheckLang returned error: %v", err)
	}

	expected := gitseek.FileInfo{
		LangPath: "content/pl/docs/test.md",
		LangLastCommit: git.CommitInfo{
			CommitID: "4c0887027d3bbcf5b31519ea9cdb8da01c1c9c31",
			DateTime: "2020-01-03T00:00:00+00:00",
			Comment:  "B: add content/pl/docs/test.md",
		},
		LangMergeCommit: nil,
		LangForkCommit:  nil,
		FileStatus:      gitseek.StatusEnFileNoLongerExists,
		EnUpdates: []gitseek.EnUpdate{
			{
				Commit: git.CommitInfo{
					CommitID: "8844dc55f5270680629cf178338f240db8d47752",
					DateTime: "2020-01-04T00:00:00+00:00",
					Comment:  "C: delete content/en/docs/test.md",
				},
				MergePoint: nil,
			},
		},
	}

	assertEqualFileInfo(t, expected, result)
}

func TestGitSeek_CheckLang_UseCase_EnFileDoesNotExist(t *testing.T) {
	ctx := t.Context()
	env := newIntegrationEnv(t, "en_file_does_not_exist")

	result, err := env.gitSeeker.CheckLang(ctx, "pl", env.pair)
	if err != nil {
		t.Fatalf("CheckLang returned error: %v", err)
	}

	expected := gitseek.FileInfo{
		LangPath: "content/pl/docs/test.md",
		LangLastCommit: git.CommitInfo{
			CommitID: "3c036010fc8368e056a4b6a199473ee07f6f8298",
			DateTime: "2020-01-02T00:00:00+00:00",
			Comment:  "B: add content/pl/docs/test.md",
		},
		LangMergeCommit: nil,
		LangForkCommit:  nil,
		FileStatus:      gitseek.StatusEnFileDoesNotExist,
		EnUpdates:       nil,
	}

	assertEqualFileInfo(t, expected, result)
}

func TestGitSeek_CheckLang_UseCase_ForkCommitExistsWithoutEnUpdatesAfterFork(t *testing.T) {
	ctx := t.Context()
	env := newIntegrationEnv(t, "fork_commit_exists_without_en_updates_after_fork")

	result, err := env.gitSeeker.CheckLang(ctx, "pl", env.pair)
	if err != nil {
		t.Fatalf("CheckLang returned error: %v", err)
	}

	expected := gitseek.FileInfo{
		LangPath: "content/pl/docs/test.md",
		LangLastCommit: git.CommitInfo{
			CommitID: "eb34068d5e46cc966188f04a42164f717d947670",
			DateTime: "2020-01-03T00:00:00+00:00",
			Comment:  "B: add content/pl/docs/test.md on pl-branch",
		},
		LangMergeCommit: &git.CommitInfo{
			CommitID: "b63a81e84e97825a8d7a6c52c27f79baf76781ff",
			DateTime: "2020-01-04T00:00:00+00:00",
			Comment:  "Merge branch 'pl-branch'",
		},
		LangForkCommit: &git.CommitInfo{
			CommitID: "a0817c4b1ebc34b35b7a726d63532ef3e835b1b6",
			DateTime: "2020-01-02T00:00:00+00:00",
			Comment:  "A: add content/en/docs/test.md",
		},
		FileStatus: gitseek.StatusLangFileUpToDate,
		EnUpdates:  nil,
	}

	assertEqualFileInfo(t, expected, result)
}

func TestGitSeek_CheckLang_UseCase_ForkCommitIsUsedAsStartPoint(t *testing.T) {
	ctx := t.Context()
	env := newIntegrationEnv(t, "fork_commit_is_used_as_start_point")

	result, err := env.gitSeeker.CheckLang(ctx, "pl", env.pair)
	if err != nil {
		t.Fatalf("CheckLang returned error: %v", err)
	}

	expected := gitseek.FileInfo{
		LangPath: "content/pl/docs/test.md",
		LangLastCommit: git.CommitInfo{
			CommitID: "eb34068d5e46cc966188f04a42164f717d947670",
			DateTime: "2020-01-03T00:00:00+00:00",
			Comment:  "B: add content/pl/docs/test.md on pl-branch",
		},
		LangMergeCommit: &git.CommitInfo{
			CommitID: "b63a81e84e97825a8d7a6c52c27f79baf76781ff",
			DateTime: "2020-01-04T00:00:00+00:00",
			Comment:  "Merge branch 'pl-branch'",
		},
		LangForkCommit: &git.CommitInfo{
			CommitID: "a0817c4b1ebc34b35b7a726d63532ef3e835b1b6",
			DateTime: "2020-01-02T00:00:00+00:00",
			Comment:  "A: add content/en/docs/test.md",
		},
		FileStatus: gitseek.StatusEnFileUpdated,
		EnUpdates: []gitseek.EnUpdate{
			{
				Commit: git.CommitInfo{
					CommitID: "59d51399b60442230bfbd9d7d7f65bac0d430c1f",
					DateTime: "2020-01-05T00:00:00+00:00",
					Comment:  "C: update content/en/docs/test.md on main after merge",
				},
				MergePoint: nil,
			},
		},
	}

	assertEqualFileInfo(t, expected, result)
}

func TestGitSeek_CheckLang_UseCase_LangLastCommitNewerByDateButForkStillDeterminesMissingEnUpdates(t *testing.T) {
	ctx := t.Context()
	env := newIntegrationEnv(t, "lang_last_commit_newer_by_date_but_fork_still_determines_missing_en_updates")

	result, err := env.gitSeeker.CheckLang(ctx, "pl", env.pair)
	if err != nil {
		t.Fatalf("CheckLang returned error: %v", err)
	}

	expected := gitseek.FileInfo{
		LangPath: "content/pl/docs/test.md",
		LangLastCommit: git.CommitInfo{
			CommitID: "7f213b0c70984e41ee4ac6f4bd621dec8a69225e",
			DateTime: "2020-01-04T00:00:00+00:00",
			Comment:  "B: add content/pl/docs/test.md after EN update date",
		},
		LangMergeCommit: &git.CommitInfo{
			CommitID: "53f6d0102083d9a325217ffd4cae9b46c409fac7",
			DateTime: "2020-01-05T00:00:00+00:00",
			Comment:  "Merge branch 'pl-branch'",
		},
		LangForkCommit: &git.CommitInfo{
			CommitID: "a0817c4b1ebc34b35b7a726d63532ef3e835b1b6",
			DateTime: "2020-01-02T00:00:00+00:00",
			Comment:  "A: add content/en/docs/test.md",
		},
		FileStatus: gitseek.StatusEnFileUpdated,
		EnUpdates: []gitseek.EnUpdate{
			{
				Commit: git.CommitInfo{
					CommitID: "1b3bdcb8f9385005b507ed6272a6aed3de121a02",
					DateTime: "2020-01-03T00:00:00+00:00",
					Comment:  "C: update content/en/docs/test.md on main",
				},
				MergePoint: nil,
			},
		},
	}

	assertEqualFileInfo(t, expected, result)
}

func TestGitSeek_CheckLang_UseCase_EnUpdatedOnMergedBranchAfterLang(t *testing.T) {
	ctx := t.Context()
	env := newIntegrationEnv(t, "en_updated_on_merged_branch_after_lang")

	result, err := env.gitSeeker.CheckLang(ctx, "pl", env.pair)
	if err != nil {
		t.Fatalf("CheckLang returned error: %v", err)
	}

	mergePoint := &git.CommitInfo{
		CommitID: "6114b9a4ac13112fbfc3a3faa05fe844cc46ac00",
		DateTime: "2020-01-05T00:00:00+00:00",
		Comment:  "Merge branch 'en-branch'",
	}

	expected := gitseek.FileInfo{
		LangPath: "content/pl/docs/test.md",
		LangLastCommit: git.CommitInfo{
			CommitID: "4c0887027d3bbcf5b31519ea9cdb8da01c1c9c31",
			DateTime: "2020-01-03T00:00:00+00:00",
			Comment:  "B: add content/pl/docs/test.md",
		},
		LangMergeCommit: nil,
		LangForkCommit:  nil,
		FileStatus:      gitseek.StatusEnFileUpdated,
		EnUpdates: []gitseek.EnUpdate{
			{
				Commit: git.CommitInfo{
					CommitID: "373c79996574fabb115a481e8e5140fc5e47dc4a",
					DateTime: "2020-01-04T00:00:00+00:00",
					Comment:  "C: update content/en/docs/test.md on en-branch",
				},
				MergePoint: mergePoint,
			},
		},
	}

	assertEqualFileInfo(t, expected, result)
}

func TestGitSeek_CheckLang_UseCase_MultipleEnUpdatesOnMergedBranchAfterLang(t *testing.T) {
	ctx := t.Context()
	env := newIntegrationEnv(t, "multiple_en_updates_on_merged_branch_after_lang")

	result, err := env.gitSeeker.CheckLang(ctx, "pl", env.pair)
	if err != nil {
		t.Fatalf("CheckLang returned error: %v", err)
	}

	mergePoint := &git.CommitInfo{
		CommitID: "b02619665337dc779030b10137004c21deb18d0b",
		DateTime: "2020-01-06T00:00:00+00:00",
		Comment:  "Merge branch 'en-branch'",
	}

	expected := gitseek.FileInfo{
		LangPath: "content/pl/docs/test.md",
		LangLastCommit: git.CommitInfo{
			CommitID: "4c0887027d3bbcf5b31519ea9cdb8da01c1c9c31",
			DateTime: "2020-01-03T00:00:00+00:00",
			Comment:  "B: add content/pl/docs/test.md",
		},
		LangMergeCommit: nil,
		LangForkCommit:  nil,
		FileStatus:      gitseek.StatusEnFileUpdated,
		EnUpdates: []gitseek.EnUpdate{
			{
				Commit: git.CommitInfo{
					CommitID: "04fe35fa9a99536282d30b1c624566681fc1b751",
					DateTime: "2020-01-05T00:00:00+00:00",
					Comment:  "D: update content/en/docs/test.md on en-branch again",
				},
				MergePoint: mergePoint,
			},
			{
				Commit: git.CommitInfo{
					CommitID: "373c79996574fabb115a481e8e5140fc5e47dc4a",
					DateTime: "2020-01-04T00:00:00+00:00",
					Comment:  "C: update content/en/docs/test.md on en-branch",
				},
				MergePoint: mergePoint,
			},
		},
	}

	assertEqualFileInfo(t, expected, result)
}

func TestGitSeek_CheckLang_UseCase_MergeCommitBringsEnUpdatesWithoutDirectEnChangeInMergeCommit(t *testing.T) {
	ctx := t.Context()
	env := newIntegrationEnv(t, "merge_commit_brings_en_updates_without_direct_en_change_in_merge_commit")

	result, err := env.gitSeeker.CheckLang(ctx, "pl", env.pair)
	if err != nil {
		t.Fatalf("CheckLang returned error: %v", err)
	}

	mergePoint := &git.CommitInfo{
		CommitID: "0e256652053295650cc3091f5ec864c818f9adc6",
		DateTime: "2020-01-06T00:00:00+00:00",
		Comment:  "Merge branch 'en-branch'",
	}

	expected := gitseek.FileInfo{
		LangPath: "content/pl/docs/test.md",
		LangLastCommit: git.CommitInfo{
			CommitID: "4c0887027d3bbcf5b31519ea9cdb8da01c1c9c31",
			DateTime: "2020-01-03T00:00:00+00:00",
			Comment:  "B: add content/pl/docs/test.md",
		},
		LangMergeCommit: nil,
		LangForkCommit:  nil,
		FileStatus:      gitseek.StatusEnFileUpdated,
		EnUpdates: []gitseek.EnUpdate{
			{
				Commit: git.CommitInfo{
					CommitID: "373c79996574fabb115a481e8e5140fc5e47dc4a",
					DateTime: "2020-01-04T00:00:00+00:00",
					Comment:  "C: update content/en/docs/test.md on en-branch",
				},
				MergePoint: mergePoint,
			},
		},
	}

	assertEqualFileInfo(t, expected, result)
}
