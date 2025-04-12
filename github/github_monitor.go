package github

//go:generate mockgen -typed -source=github_monitor.go -destination=../mocks/mock_github_monitor.go -package=mocks

import (
	"context"
	"errors"
	"fmt"
	"log"
	"time"

	"go-kweb-lang/proxycache"
)

type Monitor struct {
	gh           GitHub
	langProvider LangProvider
	storage      MonitorStorage
}

type LangProvider interface {
	LangCodes() ([]string, error)
}

type MonitorStorage interface {
	ReadLastRepoUpdatedAt() (string, error)
	WriteLastRepoUpdatedAt(value string) error
	ReadLastPRUpdatedAt(langCode string) (string, error)
	WriteLastPRUpdatedAt(langCode, value string) error
}

type MonitorTask interface {
	OnUpdate(ctx context.Context, repoUpdated bool, prUpdatedLangCodes []string) error
}

const (
	categoryLastRepoUpdatedAt = "github-monitor-repo-last-updated-at"
	categoryLastPRUpdatedAt   = "github-monitor-pr-last-updated-at"
)

type MonitorFileStorage struct {
	cacheDir string
}

func NewMonitorFileStorage(cacheDir string) *MonitorFileStorage {
	return &MonitorFileStorage{
		cacheDir: cacheDir,
	}
}

func (s *MonitorFileStorage) ReadLastRepoUpdatedAt() (string, error) {
	return proxycache.Get(
		context.Background(),
		s.cacheDir,
		categoryLastRepoUpdatedAt,
		"",
		nil,
		func(ctx context.Context) (string, error) {
			return "", nil
		},
	)
}

func (s *MonitorFileStorage) WriteLastRepoUpdatedAt(value string) error {
	return proxycache.Put(
		s.cacheDir,
		categoryLastRepoUpdatedAt,
		"",
		value,
	)
}

func (s *MonitorFileStorage) ReadLastPRUpdatedAt(langCode string) (string, error) {
	return proxycache.Get(
		context.Background(),
		s.cacheDir,
		categoryLastPRUpdatedAt,
		langCode,
		nil,
		func(ctx context.Context) (string, error) {
			return "", nil
		},
	)
}

func (s *MonitorFileStorage) WriteLastPRUpdatedAt(langCode, value string) error {
	return proxycache.Put(
		s.cacheDir,
		categoryLastPRUpdatedAt,
		langCode,
		value,
	)
}

func NewMonitor(gh GitHub, langProvider LangProvider, storage MonitorStorage) *Monitor {
	return &Monitor{
		gh:           gh,
		langProvider: langProvider,
		storage:      storage,
	}
}

func (mon *Monitor) StartIntervalCheck(
	ctx context.Context,
	delay time.Duration,
	task MonitorTask,
) error {
	type retryableError interface {
		IsRetryable() bool
	}

	for {
		err := mon.Check(ctx, task)
		if err != nil {
			var retErr retryableError

			isRetryable := errors.As(err, &retErr) && retErr.IsRetryable()
			if !isRetryable {
				return fmt.Errorf("error while checking github for updates: %w", err)
			}

			log.Printf("error while checking for github updates: %v", err)
		}

		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-time.After(delay):
		}
	}
}

func (mon *Monitor) Check(
	ctx context.Context,
	task MonitorTask,
) error {
	repoUpdated, err := mon.isRepoUpdated(ctx)
	if err != nil {
		return fmt.Errorf("error while checking if git repo has been updated: %w", err)
	}

	prUpdatedLangCodes, err := mon.prUpdatedLangCodes(ctx)
	if err != nil {
		return fmt.Errorf("error while checking if pull requests have been updated: %w", err)
	}

	if repoUpdated || len(prUpdatedLangCodes) > 0 {
		if err := task.OnUpdate(ctx, repoUpdated, prUpdatedLangCodes); err != nil {
			return fmt.Errorf("error while performing on-update task: %w", err)
		}
	}

	return nil
}

func (mon *Monitor) isRepoUpdated(ctx context.Context) (bool, error) {
	lastUpdatedAt, err := mon.storage.ReadLastRepoUpdatedAt()
	if err != nil {
		return false, err
	}

	currentUpdatedAt, err := mon.getCurrentLastRepoUpdatedAt(ctx)
	if err != nil {
		return false, err
	}

	isUpdated := len(lastUpdatedAt) == 0 || lastUpdatedAt != currentUpdatedAt

	if isUpdated {
		if err := mon.storage.WriteLastRepoUpdatedAt(currentUpdatedAt); err != nil {
			return false, err
		}
	}

	return isUpdated, nil
}

func (mon *Monitor) getCurrentLastRepoUpdatedAt(ctx context.Context) (string, error) {
	commitInfo, err := mon.gh.GetLatestCommit(ctx)
	if err != nil {
		return "", fmt.Errorf("GitHub get latest commit error: %w", err)
	}

	return commitInfo.DateTime, nil
}

func (mon *Monitor) isPRUpdated(ctx context.Context, langCode string) (bool, error) {
	lastUpdatedAt, err := mon.storage.ReadLastPRUpdatedAt(langCode)
	if err != nil {
		return false, err
	}

	currentUpdatedAt, err := mon.getCurrentLastPRUpdatedAt(ctx, langCode)
	if err != nil {
		return false, err
	}

	isUpdated := len(lastUpdatedAt) == 0 || lastUpdatedAt != currentUpdatedAt

	if isUpdated {
		if err := mon.storage.WriteLastPRUpdatedAt(langCode, currentUpdatedAt); err != nil {
			return false, err
		}
	}

	return isUpdated, nil
}

func (mon *Monitor) prUpdatedLangCodes(ctx context.Context) ([]string, error) {
	log.Printf("checking for PR changes")

	langCodes, err := mon.langProvider.LangCodes()
	if err != nil {
		return nil, fmt.Errorf("error while getting available languages: %w", err)
	}

	var updatedLangCodes []string

	for _, langCode := range langCodes {
		updated, err := mon.isPRUpdated(ctx, langCode)
		if err != nil {
			return nil, fmt.Errorf("error while checking github for changes: %w", err)
		}

		if updated {
			updatedLangCodes = append(updatedLangCodes, langCode)
		}
	}

	return updatedLangCodes, nil
}

func (mon *Monitor) getCurrentLastPRUpdatedAt(ctx context.Context, langCode string) (string, error) {
	result, err := mon.gh.PRSearch(
		ctx,
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
		return "", errors.New("no PRs found")
	}

	return result.Items[0].UpdatedAt, nil
}
