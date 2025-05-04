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
	return execToErr(lr.exec(ctx, lr.path,
		"git",
		"clone",
		url,
		".",
	))
}

func (lr *localRepo) Checkout(ctx context.Context, commitID string) error {
	return execToErr(lr.exec(ctx, lr.path,
		"git",
		"checkout",
		commitID,
	))
}

func (lr *localRepo) ListMainBranchCommits(ctx context.Context) ([]CommitInfo, error) {
	return execToCommitInfoSlice(lr.exec(ctx, lr.path,
		"git",
		"--no-pager",
		"log",
		"main",
		"--pretty=format:%H %cd %s",
		"--date=iso-strict",
		"--first-parent",
	))
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
	return execToCommitInfo(lr.exec(ctx, lr.path,
		"git",
		"log",
		"-1",
		"--format=%H %cd %s",
		"--date=iso-strict",
		"--",
		path,
	))
}

func (lr *localRepo) FindFileCommitsAfter(ctx context.Context, path string, commitIDFrom string) ([]CommitInfo, error) {
	return execToCommitInfoSlice(lr.exec(ctx, lr.path,
		"git",
		"log",
		"--pretty=format:%H %cd %s",
		"--date=iso-strict",
		commitIDFrom+"..",
		"--",
		path,
	))
}

func (lr *localRepo) ListMergePoints(ctx context.Context, commitID string) ([]CommitInfo, error) {
	return execToCommitInfoSlice(lr.exec(ctx, lr.path,
		"git",
		"--no-pager",
		"log",
		"--ancestry-path",
		"--merges",
		"--pretty=format:%H %cd %s",
		"--date=iso-strict",
		"--reverse",
		commitID+"..main",
	))
}

func (lr *localRepo) Fetch(ctx context.Context) error {
	return execToErr(lr.exec(ctx, lr.path,
		"git",
		"fetch",
	))
}

func (lr *localRepo) ListFreshCommits(ctx context.Context) ([]CommitInfo, error) {
	return execToCommitInfoSlice(lr.exec(ctx, lr.path,
		"git",
		"--no-pager",
		"log",
		"--pretty=format:%H %cd %s",
		"--date=iso-strict",
		"main..origin/main",
	))
}

func (lr *localRepo) Pull(ctx context.Context) error {
	return execToErr(lr.exec(ctx, lr.path,
		"git",
		"pull",
	))
}

func (lr *localRepo) ListFilesInCommit(ctx context.Context, commitID string) ([]string, error) {
	return execToLines(lr.exec(ctx, lr.path,
		"git",
		"diff-tree",
		"--no-commit-id",
		"--name-only",
		"-r",
		commitID,
	))
}

func (lr *localRepo) ListAncestorCommits(ctx context.Context, commitID string) ([]CommitInfo, error) {
	return execToCommitInfoSlice(lr.exec(ctx, lr.path,
		"git",
		"log",
		"--first-parent",
		"--pretty=format:%H %cd %s",
		"--date=iso-strict",
		commitID,
	))
}

func (lr *localRepo) exec(ctx context.Context, workingDir string, cmd string, args ...string) (string, error) {
	out, err := lr.runner.Exec(ctx, workingDir, cmd, args...)
	if err != nil {
		cmdStr := cmd + " " + strings.Join(args, " ")

		return "", fmt.Errorf("git command ( %v ) at working dir %v failed: %w", cmdStr, workingDir, err)
	}

	return out, nil
}

func execToErr(_ string, err error) error {
	return err
}

func execToCommitInfo(out string, err error) (CommitInfo, error) {
	if err != nil {
		return CommitInfo{}, err
	}

	return lineToCommitInfo(out), nil
}

func execToCommitInfoSlice(out string, err error) ([]CommitInfo, error) {
	if err != nil {
		return nil, err
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

func execToLines(out string, err error) ([]string, error) {
	if err != nil {
		return nil, err
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
	if len(strings.TrimSpace(line)) == 0 {
		return CommitInfo{}
	}

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
