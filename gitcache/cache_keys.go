package gitcache

import (
	"path/filepath"
)

func FileLastCommitDir(parentDir string) string {
	return filepath.Join(parentDir, "git", "file-last-commit")
}

func FileUpdatesDir(parentDir string) string {
	return filepath.Join(parentDir, "git", "file-updates")
}

func MergePointsDir(parentDir string) string {
	return filepath.Join(parentDir, "git", "merge-points")
}
