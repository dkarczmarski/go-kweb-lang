//nolint:paralleltest,goconst
package githist_test

import "testing"

// TestGitHist_MainBranchCache_IsBuiltOnFirstRead_Integration verifies that
// the main branch cache is built from the repository on first read.
func TestGitHist_MainBranchCache_IsBuiltOnFirstRead_Integration(t *testing.T) {
	ctx := t.Context()
	env := newGitHistIntegrationEnv(t, "githist_cache_basic")

	commit, err := env.gitHist.GetLastMainBranchCommit(ctx)
	if err != nil {
		t.Fatalf("GetLastMainBranchCommit returned error: %v", err)
	}

	expectedCommitID := mustGitRevParseBySubject(t, env.repoDir, "C: main third commit")

	if commit.CommitID != expectedCommitID {
		t.Fatalf("unexpected commit id: got %s want %s", commit.CommitID, expectedCommitID)
	}

	if commit.Comment != "C: main third commit" {
		t.Fatalf("unexpected commit comment: %s", commit.Comment)
	}
}

// TestGitHist_MainBranchCache_IsStableAcrossRepeatedReads_Integration verifies that
// repeated reads return the same value when cache is not invalidated.
func TestGitHist_MainBranchCache_IsStableAcrossRepeatedReads_Integration(t *testing.T) {
	ctx := t.Context()
	env := newGitHistIntegrationEnv(t, "githist_cache_basic")

	first, err := env.gitHist.GetLastMainBranchCommit(ctx)
	if err != nil {
		t.Fatalf("first GetLastMainBranchCommit returned error: %v", err)
	}

	second, err := env.gitHist.GetLastMainBranchCommit(ctx)
	if err != nil {
		t.Fatalf("second GetLastMainBranchCommit returned error: %v", err)
	}

	if first.CommitID != second.CommitID {
		t.Fatalf("commit mismatch: %s vs %s", first.CommitID, second.CommitID)
	}

	if first.Comment != second.Comment {
		t.Fatalf("comment mismatch: %s vs %s", first.Comment, second.Comment)
	}

	if first.DateTime != second.DateTime {
		t.Fatalf("datetime mismatch: %s vs %s", first.DateTime, second.DateTime)
	}
}

// TestGitHist_MainBranchCache_IsRefreshedAfterInvalidation_Integration verifies that
// cache invalidation forces fresh read from the repository.
func TestGitHist_MainBranchCache_IsRefreshedAfterInvalidation_Integration(t *testing.T) {
	ctx := t.Context()
	env := newGitHistIntegrationEnv(t, "githist_cache_with_update")

	initial, err := env.gitHist.GetLastMainBranchCommit(ctx)
	if err != nil {
		t.Fatalf("initial GetLastMainBranchCommit error: %v", err)
	}

	initialID := mustGitRevParseBySubject(t, env.repoDir, "B: main second commit")

	if initial.CommitID != initialID {
		t.Fatalf("unexpected initial commit: %s", initial.CommitID)
	}

	runScenarioScript(t, env.tmpDir, env.scenarioDir, "step_add_main_commit.sh")

	if err := env.gitHist.InvalidateMainBranchCommits(); err != nil {
		t.Fatalf("InvalidateMainBranchCommits error: %v", err)
	}

	updated, err := env.gitHist.GetLastMainBranchCommit(ctx)
	if err != nil {
		t.Fatalf("updated GetLastMainBranchCommit error: %v", err)
	}

	updatedID := mustGitRevParseBySubject(t, env.repoDir, "C: main third commit")

	if updated.CommitID != updatedID {
		t.Fatalf("unexpected updated commit: %s", updated.CommitID)
	}
}

// TestGitHist_IsMainBranchCommit_IsStableAcrossCalls_Integration verifies that
// IsMainBranchCommit returns stable results across repeated calls.
func TestGitHist_IsMainBranchCommit_IsStableAcrossCalls_Integration(t *testing.T) {
	ctx := t.Context()
	env := newGitHistIntegrationEnv(t, "githist_cache_basic")

	mainCommit := mustGitRevParseBySubject(t, env.repoDir, "C: main third commit")

	ok1, err := env.gitHist.IsMainBranchCommit(ctx, mainCommit)
	if err != nil {
		t.Fatalf("IsMainBranchCommit error: %v", err)
	}

	ok2, err := env.gitHist.IsMainBranchCommit(ctx, mainCommit)
	if err != nil {
		t.Fatalf("IsMainBranchCommit error: %v", err)
	}

	if !ok1 || !ok2 {
		t.Fatal("expected commit to belong to main branch")
	}
}

// TestGitHist_FindForkCommit_IsStableAfterInvalidation_Integration verifies that
// FindForkCommit returns the same fork commit even after cache invalidation.
func TestGitHist_FindForkCommit_IsStableAfterInvalidation_Integration(t *testing.T) {
	ctx := t.Context()
	env := newGitHistIntegrationEnv(t, "find_fork_commit_deeper_branch")

	commitID := mustGitRevParseBySubject(t, env.repoDir, "D: feature second commit")

	first, err := env.gitHist.FindForkCommit(ctx, commitID)
	if err != nil {
		t.Fatalf("FindForkCommit error: %v", err)
	}

	if first == nil {
		t.Fatal("expected fork commit")
	}

	if err := env.gitHist.InvalidateMainBranchCommits(); err != nil {
		t.Fatalf("InvalidateMainBranchCommits error: %v", err)
	}

	second, err := env.gitHist.FindForkCommit(ctx, commitID)
	if err != nil {
		t.Fatalf("FindForkCommit error: %v", err)
	}

	if second == nil {
		t.Fatal("expected fork commit")
	}

	if first.CommitID != second.CommitID {
		t.Fatalf("fork mismatch %s vs %s", first.CommitID, second.CommitID)
	}
}

// TestGitHist_FindMergeCommit_IsStableAfterInvalidation_Integration verifies that
// FindMergeCommit returns the same merge commit after cache invalidation.
func TestGitHist_FindMergeCommit_IsStableAfterInvalidation_Integration(t *testing.T) {
	ctx := t.Context()
	env := newGitHistIntegrationEnv(t, "find_merge_commit_deeper_branch")

	commitID := mustGitRevParseBySubject(t, env.repoDir, "D: feature second commit")

	first, err := env.gitHist.FindMergeCommit(ctx, commitID)
	if err != nil {
		t.Fatalf("FindMergeCommit error: %v", err)
	}

	if first == nil {
		t.Fatal("expected merge commit")
	}

	if err := env.gitHist.InvalidateMainBranchCommits(); err != nil {
		t.Fatalf("InvalidateMainBranchCommits error: %v", err)
	}

	second, err := env.gitHist.FindMergeCommit(ctx, commitID)
	if err != nil {
		t.Fatalf("FindMergeCommit error: %v", err)
	}

	if second == nil {
		t.Fatal("expected merge commit")
	}

	if first.CommitID != second.CommitID {
		t.Fatalf("merge mismatch %s vs %s", first.CommitID, second.CommitID)
	}
}
