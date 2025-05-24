package githist

import (
	"context"
	"errors"
	"fmt"

	"go-kweb-lang/git"
)

// MergeCommitFiles lists all files from the branch that was merged in the merge commit specified by mergeCommitID.
func MergeCommitFiles(ctx context.Context, gitRepoHist *GitHist, gitRepo git.Repo, mergeCommitID string) ([]string, error) {
	parents, err := gitRepo.ListCommitParents(ctx, mergeCommitID)
	if err != nil {
		return nil, fmt.Errorf("failed to list parents of merge commit %v: %w", mergeCommitID, err)
	}

	if len(parents) == 1 {
		// it is not a merge commit
		return []string{}, nil
	} else if len(parents) != 2 {
		// todo: check it. are there cases with a merge commit of 3 parents
		return nil, errors.New("merge commit should have two parents")
	}

	var branchParentCommitID string
	for i := 0; i < 2; i++ {
		parent := parents[i]
		isMain, err := gitRepoHist.IsMainBranchCommit(ctx, parent)
		if err != nil {
			return nil, fmt.Errorf("error while checking if the commit %v is on the main branch: %w",
				parent, err)
		}
		if !isMain {
			if len(branchParentCommitID) > 0 {
				return nil, errors.New("more than one parent is on the main branch. it should be impossible")
			}

			branchParentCommitID = parent
		}
	}
	if len(branchParentCommitID) == 0 {
		return nil, errors.New("no parent is on the main branch. it should be impossible")
	}

	forkCommit, err := gitRepoHist.FindForkCommit(ctx, branchParentCommitID)
	if err != nil {
		return nil, fmt.Errorf("failed to find fork commit for %v: %w",
			branchParentCommitID, err)
	}

	files, err := gitRepo.ListFilesBetweenCommits(ctx, forkCommit.CommitID, branchParentCommitID)
	if err != nil {
		return nil, fmt.Errorf("failed to list files between commits %v and %v: %w",
			forkCommit.CommitID, branchParentCommitID, err)
	}

	return files, nil
}
