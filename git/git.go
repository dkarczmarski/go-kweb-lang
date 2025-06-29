// Package git provides an interface to work with a git repository.
package git

import (
	"context"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"
	"time"

	"go-kweb-lang/git/internal"
)

// CommitInfo represents commit details.
type CommitInfo struct {
	// CommitID is a commit hash
	CommitID string
	// DateTime is a commit timestamp
	DateTime string
	// Comment is a commit comment
	Comment string
}

type NewRepoConfig struct {
	Runner CommandRunner
}

func NewRepo(path string, opts ...func(config *NewRepoConfig)) *Git {
	config := NewRepoConfig{
		Runner: &internal.StdCommandRunner{},
	}
	for _, opt := range opts {
		opt(&config)
	}

	return &Git{
		path:   path,
		runner: config.Runner,
	}
}

type CommandRunner interface {
	Exec(ctx context.Context, workingDir string, cmd string, args ...string) (string, error)
}

type Git struct {
	path   string
	runner CommandRunner
}

// Create performs a git clone using the given url.
func (g *Git) Create(ctx context.Context, url string) error {
	return execToErr(g.exec(ctx, g.path,
		"git",
		"clone",
		url,
		".",
	))
}

// Checkout checks out the revision specified by the commitID parameter.
func (g *Git) Checkout(ctx context.Context, commitID string) error {
	return execToErr(g.exec(ctx, g.path,
		"git",
		"checkout",
		commitID,
	))
}

// ListMainBranchCommits lists all commits in the main branch.
func (g *Git) ListMainBranchCommits(ctx context.Context) ([]CommitInfo, error) {
	return execToCommitInfoSlice(g.exec(ctx, g.path,
		"git",
		"--no-pager",
		"log",
		"main",
		"--pretty=format:%H %cd %s",
		"--date=iso-strict",
		"--first-parent",
	))
}

// FileExists checks whether the file exists in a repository.
func (g *Git) FileExists(path string) (bool, error) {
	_, err := os.Stat(g.path + "/" + path)
	if os.IsNotExist(err) {
		return false, nil
	}
	if err != nil {
		return false, fmt.Errorf("failed to stat file: %w", err)
	}

	return true, nil
}

// ListFiles lists all files under the given path (and all subdirectories).
func (g *Git) ListFiles(path string) ([]string, error) {
	var files []string

	fullPath := filepath.Join(g.path, path)
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

// FindFileLastCommit provides information about the last commit for the given file.
func (g *Git) FindFileLastCommit(ctx context.Context, path string) (CommitInfo, error) {
	return execToCommitInfo(g.exec(ctx, g.path,
		"git",
		"log",
		"-1",
		"--format=%H %cd %s",
		"--date=iso-strict",
		"--",
		path,
	))
}

// FindFileCommitsAfter lists all commits that contain the given file
// and that are newer than the commitIDFrom parameter.
func (g *Git) FindFileCommitsAfter(ctx context.Context, path string, commitIDFrom string) ([]CommitInfo, error) {
	return execToCommitInfoSlice(g.exec(ctx, g.path,
		"git",
		"log",
		"--pretty=format:%H %cd %s",
		"--date=iso-strict",
		commitIDFrom+"..",
		"--",
		path,
	))
}

// ListMergePoints lists all commits that are merge points starting from the commitID parameter.
func (g *Git) ListMergePoints(ctx context.Context, commitID string) ([]CommitInfo, error) {
	return execToCommitInfoSlice(g.exec(ctx, g.path,
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

// Fetch runs git fetch.
func (g *Git) Fetch(ctx context.Context) error {
	return execToErr(g.exec(ctx, g.path,
		"git",
		"fetch",
	))
}

// ListFreshCommits lists all commits in the main branch that are at origin/main and are not merged yet.
func (g *Git) ListFreshCommits(ctx context.Context) ([]CommitInfo, error) {
	return execToCommitInfoSlice(g.exec(ctx, g.path,
		"git",
		"--no-pager",
		"log",
		"--pretty=format:%H %cd %s",
		"--date=iso-strict",
		"main..origin/main",
	))
}

// Pull runs git pull.
func (g *Git) Pull(ctx context.Context) error {
	return execToErr(g.exec(ctx, g.path,
		"git",
		"pull",
	))
}

// ListFilesInCommit lists all files that are in the commit with the commitID parameter.
func (g *Git) ListFilesInCommit(ctx context.Context, commitID string) ([]string, error) {
	return execToLines(g.exec(ctx, g.path,
		"git",
		"diff-tree",
		"--no-commit-id",
		"--name-only",
		"-r",
		commitID,
	))
}

// ListAncestorCommits lists ancestor commits starting from the commitID parameter.
func (g *Git) ListAncestorCommits(ctx context.Context, commitID string) ([]CommitInfo, error) {
	return execToCommitInfoSlice(g.exec(ctx, g.path,
		"git",
		"log",
		"--first-parent",
		"--pretty=format:%H %cd %s",
		"--date=iso-strict",
		commitID,
	))
}

// ListCommitParents lists the parent commits of the given merge commit ID.
func (g *Git) ListCommitParents(ctx context.Context, commitID string) ([]string, error) {
	return execToSeparatedLines(g.exec(ctx, g.path,
		"git",
		"log",
		"-1",
		"--pretty=%P",
		commitID,
	))
}

// ListFilesBetweenCommits lists all files present in the branch between the fork commit and the last branch commit.
func (g *Git) ListFilesBetweenCommits(ctx context.Context, forkCommitID, branchLastCommitID string) ([]string, error) {
	return execToLines(g.exec(ctx, g.path,
		"git",
		"diff",
		"--name-only",
		forkCommitID,
		branchLastCommitID,
	))
}

func (g *Git) exec(ctx context.Context, workingDir string, cmd string, args ...string) (string, error) {
	out, err := g.runner.Exec(ctx, workingDir, cmd, args...)
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

	return lineToCommitInfo(out)
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
		commit, err := lineToCommitInfo(line)
		if err != nil {
			return nil, err
		}
		commits = append(commits, commit)
	}

	return commits, nil
}

func execToSeparatedLines(out string, err error) ([]string, error) {
	lines, err := execToLines(out, err)
	if err != nil {
		return lines, err
	}

	var separatedLines []string

	for _, line := range lines {
		segments := strings.Split(line, " ")
		separatedLines = append(separatedLines, segments...)
	}

	return separatedLines, nil
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

func lineToCommitInfo(line string) (CommitInfo, error) {
	if len(strings.TrimSpace(line)) == 0 {
		return CommitInfo{}, nil
	}

	segs := strings.SplitN(line, " ", 3)
	if len(segs) != 3 {
		log.Fatalf("line syntax error: %s", line)
	}

	normalizedDateTimeStr, err := normalizeDateTimeStr(segs[1])
	if err != nil {
		return CommitInfo{}, fmt.Errorf("failed to normalize date time: %w", err)
	}

	return CommitInfo{
		CommitID: segs[0],
		DateTime: normalizedDateTimeStr,
		Comment:  strings.TrimRight(segs[2], "\n"),
	}, nil
}

func normalizeDateTimeStr(dateTimeStr string) (string, error) {
	parsedTime, err := time.Parse(time.RFC3339, dateTimeStr)
	if err != nil {
		return "", fmt.Errorf("failed to parse date time string %s: %w", dateTimeStr, err)
	}

	formatted := parsedTime.Format("2006-01-02T15:04:05+00:00")

	return formatted, nil
}
