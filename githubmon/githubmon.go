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
	gitHub          GitHub
	langProvider    LangProvider
	storage         MonitorStorage
	skipGitChecking bool
	skipPRChecking  bool
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
	OnUpdate(ctx context.Context, repoUpdated bool, changedLangCodesInPR []string) error
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

func NewMonitor(
	gitHub GitHub,
	langProvider LangProvider,
	storage MonitorStorage,
	skipGitChecking bool,
	skipPRChecking bool,
) *Monitor {
	return &Monitor{
		gitHub:          gitHub,
		langProvider:    langProvider,
		storage:         storage,
		skipGitChecking: skipGitChecking,
		skipPRChecking:  skipPRChecking,
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

			log.Printf("[githubmon] failed to check for GitHub updates: %v", err)
			log.Printf("[githubmon] retrying in %s...", retryDelay)

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
	var repoUpdate repoChange

	if !mon.skipGitChecking {
		var err error
		repoUpdate, err = mon.isRepoUpdated(ctx)
		if err != nil {
			return fmt.Errorf("failed to check if git repo has been updated: %w", err)
		}
	}

	var prChanges []langChange

	if !mon.skipPRChecking {
		var err error
		prChanges, err = mon.changedLangCodesInPR(ctx)
		if err != nil {
			return fmt.Errorf("failed to check if pull requests have been updated: %w", err)
		}
	}

	prChangesExists := len(prChanges) > 0

	if repoUpdate.Updated || prChangesExists {
		changedLangCodes := make([]string, 0, len(prChanges))
		for _, change := range prChanges {
			changedLangCodes = append(changedLangCodes, change.Code)
		}

		if err := onUpdateTask.OnUpdate(ctx, repoUpdate.Updated, changedLangCodes); err != nil {
			return fmt.Errorf("failed to perform on-update task: %w", err)
		}

		if prChangesExists {
			if err := mon.writeChangedLangCodesInPR(prChanges); err != nil {
				return fmt.Errorf("failed to write language change timestamp: %w", err)
			}
		}

		if repoUpdate.Updated {
			if err := mon.storage.WriteLastRepoUpdatedAt(repoUpdate.UpdatedAt); err != nil {
				return fmt.Errorf("failed to repo change timestamp: %w", err)
			}
		}
	}

	return nil
}

func (mon *Monitor) writeChangedLangCodesInPR(changes []langChange) error {
	for _, change := range changes {
		if err := mon.storage.WriteLastPRUpdatedAt(change.Code, change.UpdatedAt); err != nil {
			return err
		}
	}

	return nil
}

type repoChange struct {
	UpdatedAt string
	Updated   bool
}

func (mon *Monitor) isRepoUpdated(ctx context.Context) (repoChange, error) {
	log.Printf("[githubmon] checking for git repo changes")

	lastUpdatedAt, err := mon.storage.ReadLastRepoUpdatedAt()
	if err != nil {
		return repoChange{}, err
	}

	currentUpdatedAt, err := mon.getCurrentLastRepoUpdatedAt(ctx)
	if err != nil {
		return repoChange{}, err
	}

	isUpdated := len(lastUpdatedAt) == 0 || lastUpdatedAt != currentUpdatedAt

	if isUpdated {
		log.Printf("[githubmon] repo updated (lastUpdatedAt=%s -> currentUpdatedAt=%s)", lastUpdatedAt, currentUpdatedAt)
	} else {
		log.Printf("[githubmon] repo not updated since %s", lastUpdatedAt)
	}

	return repoChange{
		UpdatedAt: currentUpdatedAt,
		Updated:   isUpdated,
	}, nil
}

func (mon *Monitor) getCurrentLastRepoUpdatedAt(ctx context.Context) (string, error) {
	commitInfo, err := mon.gitHub.GetLatestCommit(ctx)
	if err != nil {
		return "", fmt.Errorf("GitHub get latest commit error: %w", err)
	}

	return commitInfo.DateTime, nil
}

type langChange struct {
	Code      string
	UpdatedAt string
	Updated   bool
}

func (mon *Monitor) checkLangChange(ctx context.Context, langCode string) (langChange, error) {
	lastUpdatedAt, err := mon.storage.ReadLastPRUpdatedAt(langCode)
	if err != nil {
		return langChange{}, err
	}

	currentUpdatedAt, err := mon.getCurrentLastPRUpdatedAt(ctx, langCode)
	if err != nil {
		return langChange{}, err
	}

	isUpdated := len(currentUpdatedAt) > 0 && lastUpdatedAt != currentUpdatedAt

	return langChange{
		Code:      langCode,
		UpdatedAt: currentUpdatedAt,
		Updated:   isUpdated,
	}, nil
}

func (mon *Monitor) changedLangCodesInPR(ctx context.Context) ([]langChange, error) {
	log.Printf("[githubmon] checking for PR changes")

	langCodes, err := mon.langProvider.LangCodes()
	if err != nil {
		return nil, fmt.Errorf("failed to get available languages: %w", err)
	}

	var updatedLangCodes []langChange

	for _, langCode := range langCodes {
		change, err := mon.checkLangChange(ctx, langCode)
		if err != nil {
			return nil, fmt.Errorf("failed to check github for changes: %w", err)
		}

		if change.Updated {
			log.Printf("[githubmon][%s] PR updated at %s", langCode, change.UpdatedAt)
			updatedLangCodes = append(updatedLangCodes, change)
		} else {
			log.Printf("[githubmon][%s] no PR updates since %s", langCode, change.UpdatedAt)
		}
	}

	if len(updatedLangCodes) > 0 {
		log.Printf("[githubmon] finished checking for PR changes. changes: %v", updatedLangCodes)
	} else {
		log.Printf("[githubmon] finished checking for PR changes. no changes")
	}

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
