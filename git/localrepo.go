package git

import (
	"context"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"
)

type CommandRunner interface {
	Exec(ctx context.Context, workingDir string, cmd string, args ...string) (string, error)
}

type localRepo struct {
	path   string
	runner CommandRunner
}

func (lr *localRepo) Create(ctx context.Context, url string) error {
	_, err := lr.runner.Exec(ctx, lr.path,
		"git",
		"clone",
		url,
		".",
	)
	if err != nil {
		return fmt.Errorf("git command failed: %w", err)
	}

	return nil
}

func (lr *localRepo) Checkout(ctx context.Context, commitID string) error {
	_, err := lr.runner.Exec(ctx, lr.path,
		"git",
		"checkout",
		commitID,
	)
	if err != nil {
		return fmt.Errorf("git command ( %v ) failed: %w",
			fmt.Sprintf("git checkout %v", commitID), err)
	}

	return nil
}

func (lr *localRepo) MainBranchCommits(ctx context.Context) ([]CommitInfo, error) {
	out, err := lr.runner.Exec(ctx, lr.path,
		"git",
		"--no-pager",
		"log",
		"main",
		"--pretty=format:%H %cd %s",
		"--date=iso-strict",
		"--first-parent",
	)
	if err != nil {
		return nil, fmt.Errorf("git command ( %v ) failed: %w",
			"git --no-pager log main --pretty=format:\"%H %cd %s\" --date=iso-strict --first-parent", err)
	}

	if len(strings.TrimSpace(out)) == 0 {
		return nil, nil
	}

	lines := outputToLines(out)
	commits := make([]CommitInfo, 0, len(lines))

	for _, line := range lines {
		commits = append(commits, lineToCommitInfo(line))
	}

	return commits, nil
}

func (lr *localRepo) FileExists(path string) (bool, error) {
	_, err := os.Stat(lr.path + "/" + path)
	if os.IsNotExist(err) {
		return false, nil
	}
	if err != nil {
		return false, fmt.Errorf("failed to stat file: %w", err)
	}

	return true, nil
}

func (lr *localRepo) ListFiles(path string) ([]string, error) {
	var files []string

	fullPath := filepath.Join(lr.path, path)
	err := filepath.WalkDir(fullPath, func(path string, d os.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if !d.IsDir() {
			relPath, err := filepath.Rel(fullPath, path)
			if err != nil {
				return fmt.Errorf("failed to compute relative path: %w", err)
			}

			if strings.HasPrefix(relPath, ".git/") {
				return nil
			}

			files = append(files, relPath)
		}

		return nil
	})
	if err != nil {
		return nil, fmt.Errorf("failed to walk directory: %w", err)
	}

	return files, nil
}

func (lr *localRepo) FindFileLastCommit(ctx context.Context, path string) (CommitInfo, error) {
	out, err := lr.runner.Exec(ctx, lr.path,
		"git",
		"log",
		"-1",
		"--format=%H %cd %s",
		"--date=iso-strict",
		"--",
		path,
	)
	if err != nil {
		return CommitInfo{}, fmt.Errorf("git command failed: %w", err)
	}

	// todo: probably it should be some "not-found" error
	if len(strings.TrimSpace(out)) == 0 {
		return CommitInfo{}, nil
	}

	return lineToCommitInfo(out), nil
}

func (lr *localRepo) FindFileCommitsAfter(ctx context.Context, path string, commitIDFrom string) ([]CommitInfo, error) {
	out, err := lr.runner.Exec(ctx, lr.path,
		"git",
		"log",
		"--pretty=format:%H %cd %s",
		"--date=iso-strict",
		commitIDFrom+"..",
		"--",
		path,
	)
	if err != nil {
		return nil, fmt.Errorf("git command failed: %w", err)
	}

	if len(strings.TrimSpace(out)) == 0 {
		return nil, nil
	}

	lines := outputToLines(out)
	commits := make([]CommitInfo, 0, len(lines))

	for _, line := range lines {
		commits = append(commits, lineToCommitInfo(line))
	}

	return commits, nil
}

func (lr *localRepo) FindMergePoints(ctx context.Context, commitID string) ([]CommitInfo, error) {
	// todo: probably we can do it better, to list only necessary merging point to the main branch
	// todo: return result in reverse order?
	out, err := lr.runner.Exec(ctx, lr.path,
		"git",
		"--no-pager",
		"log",
		"--ancestry-path",
		"--merges",
		"--pretty=format:%H %cd %s",
		"--date=iso-strict",
		commitID+"..main",
	)
	if err != nil {
		return nil, fmt.Errorf("git command failed: %w", err)
	}
	if len(strings.TrimSpace(out)) == 0 {
		return nil, nil
	}

	lines := outputToLines(out)
	commits := make([]CommitInfo, 0, len(lines))

	for _, line := range lines {
		commits = append(commits, lineToCommitInfo(line))
	}

	return commits, nil
}

func (lr *localRepo) Fetch(ctx context.Context) error {
	_, err := lr.runner.Exec(ctx, lr.path,
		"git",
		"fetch",
	)
	if err != nil {
		return fmt.Errorf("git command failed: %w", err)
	}

	return nil
}

func (lr *localRepo) FreshCommits(ctx context.Context) ([]CommitInfo, error) {
	out, err := lr.runner.Exec(ctx, lr.path,
		"git",
		"--no-pager",
		"log",
		"--pretty=format:%H %cd %s",
		"--date=iso-strict",
		"main..origin/main",
	)
	if err != nil {
		return nil, fmt.Errorf("git command failed: %w", err)
	}

	if len(strings.TrimSpace(out)) == 0 {
		return nil, nil
	}

	lines := outputToLines(out)
	commits := make([]CommitInfo, 0, len(lines))

	for _, line := range lines {
		commits = append(commits, lineToCommitInfo(line))
	}

	return commits, nil
}

func (lr *localRepo) Pull(ctx context.Context) error {
	_, err := lr.runner.Exec(ctx, lr.path,
		"git",
		"pull",
	)
	if err != nil {
		return fmt.Errorf("git command failed: %w", err)
	}

	return nil
}

func (lr *localRepo) FilesInCommit(ctx context.Context, commitID string) ([]string, error) {
	out, err := lr.runner.Exec(ctx, lr.path,
		"git",
		"diff-tree",
		"--no-commit-id",
		"--name-only",
		"-r",
		commitID,
	)
	if err != nil {
		return nil, fmt.Errorf("git command failed: %w", err)
	}

	if len(strings.TrimSpace(out)) == 0 {
		return nil, nil
	}

	return outputToLines(out), nil
}

func outputToLines(out string) []string {
	return strings.Split(strings.TrimSuffix(out, "\n"), "\n")
}

func lineToCommitInfo(line string) CommitInfo {
	segs := strings.SplitN(line, " ", 3)
	if len(segs) != 3 {
		log.Fatalf("line syntax error: %s", line)
	}

	return CommitInfo{
		CommitID: segs[0],
		DateTime: segs[1],
		Comment:  strings.TrimRight(segs[2], "\n"),
	}
}
