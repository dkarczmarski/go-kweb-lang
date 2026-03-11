//nolint:paralleltest
package gitseek_test

import (
	"testing"

	"github.com/dkarczmarski/go-kweb-lang/git"
	"github.com/dkarczmarski/go-kweb-lang/gitseek"
)

func TestGitSeek_CheckLang_ComputesAndWritesCache_Integration(t *testing.T) {
	ctx := t.Context()
	env := newIntegrationEnv(t, "abc_en_update")

	result, err := env.gitSeeker.CheckLang(ctx, "pl", env.pair)
	if err != nil {
		t.Fatalf("CheckLang returned error: %v", err)
	}

	expected := expectedFileInfoABC()

	assertEqualFileInfo(t, expected, result)
	assertCachedFileInfo(t, env.cache, env.pair, expected)
}

func TestGitSeek_CheckLang_ReturnsCachedValueWithoutInvalidation_Integration(t *testing.T) {
	ctx := t.Context()
	env := newIntegrationEnv(t, "abc_en_update")

	firstResult, err := env.gitSeeker.CheckLang(ctx, "pl", env.pair)
	if err != nil {
		t.Fatalf("first CheckLang returned error: %v", err)
	}

	expectedBeforeRepoChange := expectedFileInfoABC()
	assertEqualFileInfo(t, expectedBeforeRepoChange, firstResult)

	runScenarioScript(t, env.tmpDir, env.scenarioDir, "step_add_d_en_update.sh")

	secondResult, err := env.gitSeeker.CheckLang(ctx, "pl", env.pair)
	if err != nil {
		t.Fatalf("second CheckLang returned error: %v", err)
	}

	assertEqualFileInfo(t, expectedBeforeRepoChange, secondResult)
	assertCachedFileInfo(t, env.cache, env.pair, expectedBeforeRepoChange)
}

func TestGitSeek_InvalidateFile_ForcesRecompute_Integration(t *testing.T) {
	ctx := t.Context()
	env := newIntegrationEnv(t, "abc_en_update")

	_, err := env.gitSeeker.CheckLang(ctx, "pl", env.pair)
	if err != nil {
		t.Fatalf("initial CheckLang returned error: %v", err)
	}

	runScenarioScript(t, env.tmpDir, env.scenarioDir, "step_add_d_en_update.sh")

	if err := env.gitSeeker.InvalidateFile("pl", env.pair.LangPath); err != nil {
		t.Fatalf("InvalidateFile returned error: %v", err)
	}

	result, err := env.gitSeeker.CheckLang(ctx, "pl", env.pair)
	if err != nil {
		t.Fatalf("CheckLang after invalidation returned error: %v", err)
	}

	expected := expectedFileInfoABCD()
	assertEqualFileInfo(t, expected, result)
	assertCachedFileInfo(t, env.cache, env.pair, expected)
}

func expectedFileInfoABC() gitseek.FileInfo {
	return gitseek.FileInfo{
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
}

func expectedFileInfoABCD() gitseek.FileInfo {
	return gitseek.FileInfo{
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
}
