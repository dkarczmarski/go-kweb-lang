package githubmon

//go:generate mockgen -typed -source=githubmon.go -destination=./internal/mocks/mocks.go -package=mocks

import (
	"context"
	"errors"
	"fmt"
	"log"
	"time"

	"go-kweb-lang/github"
	"go-kweb-lang/proxycache"
)

type Monitor struct {
	gitHub       GitHub
	langProvider LangProvider
	storage      MonitorStorage
}

type GitHub interface {
	PRSearch(ctx context.Context, filter github.PRSearchFilter, page github.PageRequest) (*github.PRSearchResult, error)

	GetLatestCommit(ctx context.Context) (*github.CommitInfo, error)
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

type OnUpdateTask interface {
	OnUpdate(ctx context.Context, repoUpdated bool, prUpdatedLangCodes []string) error
}

const (
	bucketLastRepoUpdatedAt = "github-monitor-repo-last-updated-at"
	bucketLastPRUpdatedAt   = "github-monitor-pr-last-updated-at"
	singleKey               = ""
)

type CacheStore interface {
	Read(bucket, key string, buff any) (bool, error)
	Write(bucket, key string, data any) error
}

type MonitorFileStorage struct {
	cacheStore CacheStore
}

func NewMonitorFileStorage(cacheStore CacheStore) *MonitorFileStorage {
	return &MonitorFileStorage{
		cacheStore: cacheStore,
	}
}

func (s *MonitorFileStorage) ReadLastRepoUpdatedAt() (string, error) {
	return proxycache.Get(
		context.Background(),
		s.cacheStore,
		bucketLastRepoUpdatedAt,
		"",
		nil,
		func(ctx context.Context) (string, error) {
			return "", nil
		},
	)
}

func (s *MonitorFileStorage) WriteLastRepoUpdatedAt(value string) error {
	return s.cacheStore.Write(
		bucketLastRepoUpdatedAt,
		"",
		value,
	)
}

func (s *MonitorFileStorage) ReadLastPRUpdatedAt(langCode string) (string, error) {
	return proxycache.Get(
		context.Background(),
		s.cacheStore,
		fmt.Sprintf("lang/%s/%s", langCode, bucketLastPRUpdatedAt),
		singleKey,
		nil,
		func(ctx context.Context) (string, error) {
			return "", nil
		},
	)
}

func (s *MonitorFileStorage) WriteLastPRUpdatedAt(langCode, value string) error {
	return s.cacheStore.Write(
		fmt.Sprintf("lang/%s/%s", langCode, bucketLastPRUpdatedAt),
		singleKey,
		value,
	)
}

func NewMonitor(gitHub GitHub, langProvider LangProvider, storage MonitorStorage) *Monitor {
	return &Monitor{
		gitHub:       gitHub,
		langProvider: langProvider,
		storage:      storage,
	}
}

func (mon *Monitor) IntervalCheck(
	ctx context.Context,
	intervalDelay time.Duration,
	onUpdateTask OnUpdateTask,
) error {
	retryDelay := time.Second * 15 // magic number

	for {
		err := mon.RetryCheck(ctx, retryDelay, onUpdateTask)
		if err != nil {
			return err
		}

		timer := time.NewTimer(intervalDelay)
		select {
		case <-ctx.Done():
			if !timer.Stop() {
				<-timer.C
			}
			return ctx.Err()
		case <-timer.C:
		}
	}
}

func (mon *Monitor) RetryCheck(
	ctx context.Context,
	retryDelay time.Duration,
	onUpdateTask OnUpdateTask,
) error {
	type retryableError interface {
		IsRetryable() bool
	}

	for {
		err := mon.Check(ctx, onUpdateTask)
		if err != nil {
			var retErr retryableError

			isRetryable := errors.As(err, &retErr) && retErr.IsRetryable()
			if !isRetryable {
				return fmt.Errorf("failed to check GitHub for updates: %w", err)
			}

			log.Printf("failed to check for GitHub updates: %v", err)
			log.Printf("retrying in %s...", retryDelay)

			timer := time.NewTimer(retryDelay)
			select {
			case <-ctx.Done():
				if !timer.Stop() {
					<-timer.C
				}
				return ctx.Err()
			case <-timer.C:
			}

			continue
		}

		return nil
	}
}

func (mon *Monitor) Check(
	ctx context.Context,
	onUpdateTask OnUpdateTask,
) error {
	repoUpdated, err := mon.isRepoUpdated(ctx)
	if err != nil {
		return fmt.Errorf("error while checking if git repo has been updated: %w", err)
	}

	prUpdatedLangCodes, err := mon.prUpdatedLangCodes(ctx)
	if err != nil {
		return fmt.Errorf("failed to check if pull requests have been updated: %w", err)
	}

	if repoUpdated || len(prUpdatedLangCodes) > 0 {
		if err := onUpdateTask.OnUpdate(ctx, repoUpdated, prUpdatedLangCodes); err != nil {
			return fmt.Errorf("failed to perform on-update task: %w", err)
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
	commitInfo, err := mon.gitHub.GetLatestCommit(ctx)
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

	isUpdated := len(currentUpdatedAt) > 0 && (len(lastUpdatedAt) == 0 || lastUpdatedAt != currentUpdatedAt)

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

	log.Printf("finished checking for PR changes")

	return updatedLangCodes, nil
}

func (mon *Monitor) getCurrentLastPRUpdatedAt(ctx context.Context, langCode string) (string, error) {
	result, err := mon.gitHub.PRSearch(
		ctx,
		github.PRSearchFilter{
			LangCode: langCode,
		},
		github.PageRequest{
			Sort:    "updated",
			Order:   "desc",
			PerPage: 1,
		},
	)
	if err != nil {
		return "", err
	}

	if len(result.Items) == 0 {
		return "", nil
	}

	return result.Items[0].UpdatedAt, nil
}
