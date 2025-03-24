package github

import (
	"context"
	"fmt"
	"log"
	"os"
	"path/filepath"
)

type PRMonitor struct {
	gh       *GitHub
	cacheDir string
	tasks    []OnPRUpdateTask
}

type OnPRUpdateTask interface {
	Run(ctx context.Context) error
}

func (mon *PRMonitor) maxUpdatedAt(langCode string) (string, error) {
	result, err := mon.gh.PRSearch(
		PRSearchFilter{
			LangCode: langCode,
		},
		PageRequest{
			Sort:    "updated",
			Order:   "desc",
			PerPage: 1,
		},
	)
	if err != nil {
		return "", err
	}

	if len(result.Items) == 0 {
		return "", fmt.Errorf("no PRs found")
	}

	return result.Items[0].UpdatedAt, nil
}

func (mon *PRMonitor) lastUpdatedAtFile() string {
	return filepath.Join(mon.cacheDir, "github", "last-updated-at.txt")
}

func (mon *PRMonitor) lastMaxUpdatedAt() (string, error) {
	path := mon.lastUpdatedAtFile()

	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return "", nil
		}

		return "", fmt.Errorf("error while reading file %s: %w", path, err)
	}

	return string(data), nil
}

func (mon *PRMonitor) setLastMaxUpdatedAt(maxUpdatedAt string) error {
	path := mon.lastUpdatedAtFile()

	if err := os.WriteFile(path, []byte(maxUpdatedAt), 0644); err != nil {
		return fmt.Errorf("error while writing to file %s: %w", path, err)
	}

	return nil
}

func (mon *PRMonitor) Check(ctx context.Context, langCode string) (bool, error) {
	lastMaxUpdatedAt, err := mon.lastMaxUpdatedAt()
	if err != nil {
		return false, fmt.Errorf("error while getting the last updatedAt value: %w", err)
	}

	maxUpdatedAt, err := mon.maxUpdatedAt(langCode)
	if err != nil {
		return false, fmt.Errorf("error while getting the maximal updatedAt value: %w", err)
	}

	if lastMaxUpdatedAt == maxUpdatedAt {
		return false, nil
	}

	for _, task := range mon.tasks {
		if err := task.Run(ctx); err != nil {
			log.Printf("task error: %v", err)
		}
	}

	if err := mon.setLastMaxUpdatedAt(maxUpdatedAt); err != nil {
		return true, fmt.Errorf("error while saving the last PRs: %w", err)
	}

	return true, nil
}
