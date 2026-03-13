//nolint:paralleltest,goconst
package githist_test

import (
	"errors"
	"reflect"
	"testing"

	"github.com/dkarczmarski/go-kweb-lang/githist"
)

// TestGitHist_FindForkCommit_ReturnsErrCommitOnMainBranch_Integration verifies that
// FindForkCommit returns ErrCommitOnMainBranch when the given commit already
// belongs to the main branch.
func TestGitHist_FindForkCommit_ReturnsErrCommitOnMainBranch_Integration(t *testing.T) {
	ctx := t.Context()
	env := newGitHistIntegrationEnv(t, "find_fork_commit_on_main")

	commitID := mustGitRevParseBySubject(t, env.repoDir, "B: main second commit")

	forkCommit, err := env.gitHist.FindForkCommit(ctx, commitID)
	if !errors.Is(err, githist.ErrCommitOnMainBranch) {
		t.Fatalf("expected ErrCommitOnMainBranch, got: %v", err)
	}

	if forkCommit != nil {
		t.Fatalf("expected nil fork commit, got %+v", forkCommit)
	}
}

// TestGitHist_FindForkCommit_ReturnsForkPointForBranchCommit_Integration verifies that
// FindForkCommit returns the correct fork point when the commit comes from a branch
// created from the main branch.
func TestGitHist_FindForkCommit_ReturnsForkPointForBranchCommit_Integration(t *testing.T) {
	ctx := t.Context()
	env := newGitHistIntegrationEnv(t, "find_fork_commit_simple_branch")

	branchCommitID := mustGitRevParseBySubject(t, env.repoDir, "B: feature commit")
	expectedForkCommitID := mustGitRevParseBySubject(t, env.repoDir, "A: main base commit")

	forkCommit, err := env.gitHist.FindForkCommit(ctx, branchCommitID)
	if err != nil {
		t.Fatalf("FindForkCommit returned error: %v", err)
	}

	if forkCommit == nil {
		t.Fatal("expected non-nil fork commit, got nil")
	}

	if forkCommit.CommitID != expectedForkCommitID {
		t.Fatalf("unexpected fork commit id: got %s, want %s", forkCommit.CommitID, expectedForkCommitID)
	}

	if forkCommit.Comment != "A: main base commit" {
		t.Fatalf("unexpected fork commit comment: got %q", forkCommit.Comment)
	}
}

// TestGitHist_FindForkCommit_ReturnsNearestCommonMainCommit_Integration verifies that
// FindForkCommit returns the nearest common commit on the main branch for a commit
// located deeper in a feature branch.
func TestGitHist_FindForkCommit_ReturnsNearestCommonMainCommit_Integration(t *testing.T) {
	ctx := t.Context()
	env := newGitHistIntegrationEnv(t, "find_fork_commit_deeper_branch")

	branchCommitID := mustGitRevParseBySubject(t, env.repoDir, "D: feature second commit")
	expectedForkCommitID := mustGitRevParseBySubject(t, env.repoDir, "B: main second commit")

	forkCommit, err := env.gitHist.FindForkCommit(ctx, branchCommitID)
	if err != nil {
		t.Fatalf("FindForkCommit returned error: %v", err)
	}

	if forkCommit == nil {
		t.Fatal("expected non-nil fork commit, got nil")
	}

	if forkCommit.CommitID != expectedForkCommitID {
		t.Fatalf("unexpected fork commit id: got %s, want %s", forkCommit.CommitID, expectedForkCommitID)
	}

	if forkCommit.Comment != "B: main second commit" {
		t.Fatalf("unexpected fork commit comment: got %q", forkCommit.Comment)
	}
}

// TestGitHist_FindMergeCommit_ReturnsErrCommitOnMainBranch_Integration verifies that
// FindMergeCommit returns ErrCommitOnMainBranch when the given commit already
// belongs to the main branch.
func TestGitHist_FindMergeCommit_ReturnsErrCommitOnMainBranch_Integration(t *testing.T) {
	ctx := t.Context()
	env := newGitHistIntegrationEnv(t, "find_merge_commit_on_main")

	commitID := mustGitRevParseBySubject(t, env.repoDir, "B: main second commit")

	mergeCommit, err := env.gitHist.FindMergeCommit(ctx, commitID)
	if !errors.Is(err, githist.ErrCommitOnMainBranch) {
		t.Fatalf("expected ErrCommitOnMainBranch, got: %v", err)
	}

	if mergeCommit != nil {
		t.Fatalf("expected nil merge commit, got %+v", mergeCommit)
	}
}

// TestGitHist_FindMergeCommit_ReturnsMergeCommitForBranchCommit_Integration verifies that
// FindMergeCommit returns the merge commit when the given commit comes from a branch
// that was merged into the main branch.
func TestGitHist_FindMergeCommit_ReturnsMergeCommitForBranchCommit_Integration(t *testing.T) {
	ctx := t.Context()
	env := newGitHistIntegrationEnv(t, "find_merge_commit_simple_branch")

	branchCommitID := mustGitRevParseBySubject(t, env.repoDir, "B: feature commit")
	expectedMergeCommitID := mustGitRevParseBySubject(t, env.repoDir, "M: merge feature")

	mergeCommit, err := env.gitHist.FindMergeCommit(ctx, branchCommitID)
	if err != nil {
		t.Fatalf("FindMergeCommit returned error: %v", err)
	}

	if mergeCommit == nil {
		t.Fatal("expected non-nil merge commit, got nil")
	}

	if mergeCommit.CommitID != expectedMergeCommitID {
		t.Fatalf("unexpected merge commit id: got %s, want %s", mergeCommit.CommitID, expectedMergeCommitID)
	}

	if mergeCommit.Comment != "M: merge feature" {
		t.Fatalf("unexpected merge commit comment: got %q", mergeCommit.Comment)
	}
}

// TestGitHist_FindMergeCommit_ReturnsMergeCommitForDeepBranchCommit_Integration verifies that
// FindMergeCommit returns the correct merge commit for a commit located deeper in a feature branch.
func TestGitHist_FindMergeCommit_ReturnsMergeCommitForDeepBranchCommit_Integration(t *testing.T) {
	ctx := t.Context()
	env := newGitHistIntegrationEnv(t, "find_merge_commit_deeper_branch")

	branchCommitID := mustGitRevParseBySubject(t, env.repoDir, "D: feature second commit")
	expectedMergeCommitID := mustGitRevParseBySubject(t, env.repoDir, "M: merge feature")

	mergeCommit, err := env.gitHist.FindMergeCommit(ctx, branchCommitID)
	if err != nil {
		t.Fatalf("FindMergeCommit returned error: %v", err)
	}

	if mergeCommit == nil {
		t.Fatal("expected non-nil merge commit, got nil")
	}

	if mergeCommit.CommitID != expectedMergeCommitID {
		t.Fatalf("unexpected merge commit id: got %s, want %s", mergeCommit.CommitID, expectedMergeCommitID)
	}

	if mergeCommit.Comment != "M: merge feature" {
		t.Fatalf("unexpected merge commit comment: got %q", mergeCommit.Comment)
	}
}

// TestGitHist_GetLastMainBranchCommit_ReturnsLatestMainCommit_Integration verifies that
// GetLastMainBranchCommit returns the newest commit from the main branch history.
func TestGitHist_GetLastMainBranchCommit_ReturnsLatestMainCommit_Integration(t *testing.T) {
	ctx := t.Context()
	env := newGitHistIntegrationEnv(t, "get_last_main_branch_commit_basic")

	commit, err := env.gitHist.GetLastMainBranchCommit(ctx)
	if err != nil {
		t.Fatalf("GetLastMainBranchCommit returned error: %v", err)
	}

	expectedCommitID := mustGitRevParseBySubject(t, env.repoDir, "C: main third commit")

	if commit.CommitID != expectedCommitID {
		t.Fatalf("unexpected commit id: got %s, want %s", commit.CommitID, expectedCommitID)
	}

	if commit.Comment != "C: main third commit" {
		t.Fatalf("unexpected commit comment: got %q", commit.Comment)
	}
}

// TestGitHist_GetLastMainBranchCommit_ReturnsSameValueFromCache_Integration verifies that
// GetLastMainBranchCommit returns the same latest main commit when called again from cache.
func TestGitHist_GetLastMainBranchCommit_ReturnsSameValueFromCache_Integration(t *testing.T) {
	ctx := t.Context()
	env := newGitHistIntegrationEnv(t, "get_last_main_branch_commit_basic")

	firstCommit, err := env.gitHist.GetLastMainBranchCommit(ctx)
	if err != nil {
		t.Fatalf("first GetLastMainBranchCommit returned error: %v", err)
	}

	secondCommit, err := env.gitHist.GetLastMainBranchCommit(ctx)
	if err != nil {
		t.Fatalf("second GetLastMainBranchCommit returned error: %v", err)
	}

	if firstCommit.CommitID != secondCommit.CommitID {
		t.Fatalf("unexpected commit id mismatch: first=%s second=%s", firstCommit.CommitID, secondCommit.CommitID)
	}

	if firstCommit.Comment != secondCommit.Comment {
		t.Fatalf("unexpected commit comment mismatch: first=%q second=%q", firstCommit.Comment, secondCommit.Comment)
	}

	if firstCommit.DateTime != secondCommit.DateTime {
		t.Fatalf("unexpected commit datetime mismatch: first=%q second=%q", firstCommit.DateTime, secondCommit.DateTime)
	}
}

// TestGitHist_GetLastMainBranchCommit_ReturnsUpdatedValueAfterCacheInvalidation_Integration verifies that
// GetLastMainBranchCommit returns a newer main commit after repository update
// and explicit invalidation of the cached main branch commits.
func TestGitHist_GetLastMainBranchCommit_ReturnsUpdatedValueAfterCacheInvalidation_Integration(t *testing.T) {
	ctx := t.Context()
	env := newGitHistIntegrationEnv(t, "get_last_main_branch_commit_with_update")

	initialCommit, err := env.gitHist.GetLastMainBranchCommit(ctx)
	if err != nil {
		t.Fatalf("initial GetLastMainBranchCommit returned error: %v", err)
	}

	initialExpectedCommitID := mustGitRevParseBySubject(t, env.repoDir, "B: main second commit")

	if initialCommit.CommitID != initialExpectedCommitID {
		t.Fatalf(
			"unexpected initial commit id: got %s, want %s",
			initialCommit.CommitID,
			initialExpectedCommitID,
		)
	}

	runScenarioScript(t, env.tmpDir, env.scenarioDir, "step_add_main_commit.sh")

	if err := env.gitHist.InvalidateMainBranchCommits(); err != nil {
		t.Fatalf("InvalidateMainBranchCommits returned error: %v", err)
	}

	updatedCommit, err := env.gitHist.GetLastMainBranchCommit(ctx)
	if err != nil {
		t.Fatalf("updated GetLastMainBranchCommit returned error: %v", err)
	}

	updatedExpectedCommitID := mustGitRevParseBySubject(t, env.repoDir, "C: main third commit")

	if updatedCommit.CommitID != updatedExpectedCommitID {
		t.Fatalf(
			"unexpected updated commit id: got %s, want %s",
			updatedCommit.CommitID,
			updatedExpectedCommitID,
		)
	}

	if updatedCommit.Comment != "C: main third commit" {
		t.Fatalf(
			"unexpected updated commit comment: got %q",
			updatedCommit.Comment,
		)
	}
}

// TestGitHist_IsMainBranchCommit_ReturnsTrueForMainCommit_Integration verifies that
// IsMainBranchCommit returns true for a commit that belongs to the main branch.
func TestGitHist_IsMainBranchCommit_ReturnsTrueForMainCommit_Integration(t *testing.T) {
	ctx := t.Context()
	env := newGitHistIntegrationEnv(t, "is_main_branch_commit_basic")

	commitID := mustGitRevParseBySubject(t, env.repoDir, "B: main second commit")

	isMain, err := env.gitHist.IsMainBranchCommit(ctx, commitID)
	if err != nil {
		t.Fatalf("IsMainBranchCommit returned error: %v", err)
	}

	if !isMain {
		t.Fatalf("expected commit %s to be on main branch", commitID)
	}
}

// TestGitHist_IsMainBranchCommit_ReturnsFalseForBranchCommit_Integration verifies that
// IsMainBranchCommit returns false for a commit that exists only on a feature branch.
func TestGitHist_IsMainBranchCommit_ReturnsFalseForBranchCommit_Integration(t *testing.T) {
	ctx := t.Context()
	env := newGitHistIntegrationEnv(t, "is_main_branch_commit_branch_commit")

	commitID := mustGitRevParseBySubject(t, env.repoDir, "B: feature commit")

	isMain, err := env.gitHist.IsMainBranchCommit(ctx, commitID)
	if err != nil {
		t.Fatalf("IsMainBranchCommit returned error: %v", err)
	}

	if isMain {
		t.Fatalf("expected commit %s to not be on main branch", commitID)
	}
}

// TestGitHist_IsMainBranchCommit_ReturnsTrueForMergedCommit_Integration verifies that
// IsMainBranchCommit returns true for a merge commit that belongs to the main branch.
func TestGitHist_IsMainBranchCommit_ReturnsTrueForMergedCommit_Integration(t *testing.T) {
	ctx := t.Context()
	env := newGitHistIntegrationEnv(t, "is_main_branch_commit_merged_branch")

	commitID := mustGitRevParseBySubject(t, env.repoDir, "M: merge feature")

	isMain, err := env.gitHist.IsMainBranchCommit(ctx, commitID)
	if err != nil {
		t.Fatalf("IsMainBranchCommit returned error: %v", err)
	}

	if !isMain {
		t.Fatalf("expected merge commit %s to be on main branch", commitID)
	}
}

// TestGitHist_IsMainBranchCommit_ReturnsFalseForUnknownCommit_Integration verifies that
// IsMainBranchCommit returns false for a commit id that does not exist in the repository.
func TestGitHist_IsMainBranchCommit_ReturnsFalseForUnknownCommit_Integration(t *testing.T) {
	ctx := t.Context()
	env := newGitHistIntegrationEnv(t, "is_main_branch_commit_basic")

	isMain, err := env.gitHist.IsMainBranchCommit(ctx, "deadbeefdeadbeefdeadbeefdeadbeefdeadbeef")
	if err != nil {
		t.Fatalf("IsMainBranchCommit returned error: %v", err)
	}

	if isMain {
		t.Fatal("expected unknown commit to not be on main branch")
	}
}

// TestGitHist_MergeCommitFiles_ReturnsEmptyForNonMergeCommit_Integration verifies that
// MergeCommitFiles returns an empty list when the given commit is not a merge commit.
func TestGitHist_MergeCommitFiles_ReturnsEmptyForNonMergeCommit_Integration(t *testing.T) {
	ctx := t.Context()
	env := newGitHistIntegrationEnv(t, "merge_commit_files_non_merge")

	commitID := mustGitRevParseBySubject(t, env.repoDir, "B: main second commit")

	files, err := env.gitHist.MergeCommitFiles(ctx, commitID)
	if err != nil {
		t.Fatalf("MergeCommitFiles returned error: %v", err)
	}

	if len(files) != 0 {
		t.Fatalf("expected empty file list, got %v", files)
	}
}

// TestGitHist_MergeCommitFiles_ReturnsFilesFromMergedBranch_Integration verifies that
// MergeCommitFiles returns files that were introduced on the merged feature branch.
func TestGitHist_MergeCommitFiles_ReturnsFilesFromMergedBranch_Integration(t *testing.T) {
	ctx := t.Context()
	env := newGitHistIntegrationEnv(t, "merge_commit_files_simple_merge")

	mergeCommitID := mustGitRevParseBySubject(t, env.repoDir, "M: merge feature")

	files, err := env.gitHist.MergeCommitFiles(ctx, mergeCommitID)
	if err != nil {
		t.Fatalf("MergeCommitFiles returned error: %v", err)
	}

	expected := []string{
		"docs/feature.md",
	}

	if !reflect.DeepEqual(expected, files) {
		t.Fatalf("unexpected files: got %v, want %v", files, expected)
	}
}

// TestGitHist_MergeCommitFiles_ReturnsAllFilesFromMergedBranchHistory_Integration verifies that
// MergeCommitFiles returns all files changed on the merged branch between fork and branch head.
func TestGitHist_MergeCommitFiles_ReturnsAllFilesFromMergedBranchHistory_Integration(t *testing.T) {
	ctx := t.Context()
	env := newGitHistIntegrationEnv(t, "merge_commit_files_multi_commit_branch")

	mergeCommitID := mustGitRevParseBySubject(t, env.repoDir, "M: merge feature")

	files, err := env.gitHist.MergeCommitFiles(ctx, mergeCommitID)
	if err != nil {
		t.Fatalf("MergeCommitFiles returned error: %v", err)
	}

	expected := []string{
		"docs/feature-a.md",
		"docs/feature-b.md",
	}

	if !reflect.DeepEqual(expected, files) {
		t.Fatalf("unexpected files: got %v, want %v", files, expected)
	}
}

// TestGitHist_PullRefresh_ReturnsNoFilesWhenOriginHasNoNewCommits_Integration verifies that
// PullRefresh returns no changed files when origin/main has no new commits.
func TestGitHist_PullRefresh_ReturnsNoFilesWhenOriginHasNoNewCommits_Integration(t *testing.T) {
	ctx := t.Context()
	env := newGitHistRemoteIntegrationEnv(t, "pull_refresh_no_changes")

	files, err := env.gitHist.PullRefresh(ctx)
	if err != nil {
		t.Fatalf("PullRefresh returned error: %v", err)
	}

	if len(files) != 0 {
		t.Fatalf("expected no changed files, got %v", files)
	}

	headSubject := mustGitHeadSubject(t, env.repoDir)
	if headSubject != "B: main second commit" {
		t.Fatalf("unexpected HEAD after PullRefresh: got %q", headSubject)
	}
}

// TestGitHist_PullRefresh_ReturnsFilesForNewMainCommit_Integration verifies that
// PullRefresh returns files changed by a new commit pushed to origin/main.
func TestGitHist_PullRefresh_ReturnsFilesForNewMainCommit_Integration(t *testing.T) {
	ctx := t.Context()
	env := newGitHistRemoteIntegrationEnv(t, "pull_refresh_new_main_commit")

	runScenarioScript(t, env.tmpDir, env.scenarioDir, "step_add_main_commit.sh")

	files, err := env.gitHist.PullRefresh(ctx)
	if err != nil {
		t.Fatalf("PullRefresh returned error: %v", err)
	}

	expectedFiles := []string{
		"docs/main.md",
	}

	if !reflect.DeepEqual(expectedFiles, files) {
		t.Fatalf("unexpected changed files: got %v, want %v", files, expectedFiles)
	}

	headSubject := mustGitHeadSubject(t, env.repoDir)
	if headSubject != "C: main third commit" {
		t.Fatalf("unexpected HEAD after PullRefresh: got %q", headSubject)
	}
}

// TestGitHist_PullRefresh_ReturnsFilesForMergedBranch_Integration verifies that
// PullRefresh returns files changed by all fresh commits pushed to origin/main,
// including commits that later become part of a merge.
func TestGitHist_PullRefresh_ReturnsFilesForMergedBranch_Integration(t *testing.T) {
	ctx := t.Context()
	env := newGitHistRemoteIntegrationEnv(t, "pull_refresh_merge_commit")

	runScenarioScript(t, env.tmpDir, env.scenarioDir, "step_add_merge_commit.sh")

	files, err := env.gitHist.PullRefresh(ctx)
	if err != nil {
		t.Fatalf("PullRefresh returned error: %v", err)
	}

	expectedFiles := []string{
		"docs/feature.md",
		"docs/main.md",
	}

	if !reflect.DeepEqual(expectedFiles, files) {
		t.Fatalf("unexpected changed files: got %v, want %v", files, expectedFiles)
	}

	headSubject := mustGitHeadSubject(t, env.repoDir)
	if headSubject != "M: merge feature" {
		t.Fatalf("unexpected HEAD after PullRefresh: got %q", headSubject)
	}
}
