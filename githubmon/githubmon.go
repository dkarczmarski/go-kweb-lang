package githubmon

//go:generate mockgen -typed -source=githubmon.go -destination=./internal/mocks/mocks.go -package=mocks

import (
	"context"
	"errors"
	"fmt"
	"log"
	"path/filepath"
	"time"

	"github.com/dkarczmarski/go-kweb-lang/github"
)

const (
	firstPage = 1
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
	ReadLastPRUpdatedAt() (string, error)
	WriteLastPRUpdatedAt(value string) error
	ReadLastLangPRUpdatedAt(langCode string) (string, error)
	WriteLastLangPRUpdatedAt(langCode, value string) error
}

type OnUpdateTask interface {
	OnUpdate(ctx context.Context, repoUpdated bool, changedLangCodesInPR []string) error
}

const (
	bucketLastRepoUpdatedAt = "githubmon-repo-updated-at"
	bucketLastPRUpdatedAt   = "githubmon-pr-updated-at"
	singleKey               = ""

	defaultRetryDelaySeconds         = 15
	nonRetryableFallbackDelayMinutes = 5
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

// LastRepoUpdatedAtCacheBucket returns the cache bucket used for the repository
// last-updated timestamp.
func LastRepoUpdatedAtCacheBucket() string {
	return bucketLastRepoUpdatedAt
}

// LastRepoUpdatedAtCacheKey returns the cache key used for the repository
// last-updated timestamp.
func LastRepoUpdatedAtCacheKey() string {
	return singleKey
}

// LastPRUpdatedAtCacheBucket returns the cache bucket used for the global pull
// request last-updated timestamp.
func LastPRUpdatedAtCacheBucket() string {
	return bucketLastPRUpdatedAt
}

// LastPRUpdatedAtCacheKey returns the cache key used for the global pull
// request last-updated timestamp.
func LastPRUpdatedAtCacheKey() string {
	return singleKey
}

// LastLangPRUpdatedAtCacheBucket returns the cache bucket used for the given
// language pull request last-updated timestamp.
func LastLangPRUpdatedAtCacheBucket(langCode string) string {
	return filepath.Join("lang", langCode, bucketLastPRUpdatedAt)
}

// LastLangPRUpdatedAtCacheKey returns the cache key used for the language pull
// request last-updated timestamp.
func LastLangPRUpdatedAtCacheKey() string {
	return singleKey
}

func (s *MonitorFileStorage) ReadLastRepoUpdatedAt() (string, error) {
	var value string

	exists, err := s.cacheStore.Read(
		LastRepoUpdatedAtCacheBucket(),
		LastRepoUpdatedAtCacheKey(),
		&value,
	)
	if err != nil {
		return "", fmt.Errorf("read repo update timestamp cache: %w", err)
	}

	if !exists {
		return "", nil
	}

	return value, nil
}

func (s *MonitorFileStorage) WriteLastRepoUpdatedAt(value string) error {
	if err := s.cacheStore.Write(
		LastRepoUpdatedAtCacheBucket(),
		LastRepoUpdatedAtCacheKey(),
		value,
	); err != nil {
		return fmt.Errorf("write repo update timestamp cache: %w", err)
	}

	return nil
}

func (s *MonitorFileStorage) ReadLastPRUpdatedAt() (string, error) {
	var value string

	exists, err := s.cacheStore.Read(
		LastPRUpdatedAtCacheBucket(),
		LastPRUpdatedAtCacheKey(),
		&value,
	)
	if err != nil {
		return "", fmt.Errorf("read PR update timestamp cache: %w", err)
	}

	if !exists {
		return "", nil
	}

	return value, nil
}

func (s *MonitorFileStorage) WriteLastPRUpdatedAt(value string) error {
	if err := s.cacheStore.Write(
		LastPRUpdatedAtCacheBucket(),
		LastPRUpdatedAtCacheKey(),
		value,
	); err != nil {
		return fmt.Errorf("write PR update timestamp cache: %w", err)
	}

	return nil
}

func (s *MonitorFileStorage) ReadLastLangPRUpdatedAt(langCode string) (string, error) {
	var value string

	exists, err := s.cacheStore.Read(
		LastLangPRUpdatedAtCacheBucket(langCode),
		LastLangPRUpdatedAtCacheKey(),
		&value,
	)
	if err != nil {
		return "", fmt.Errorf("read PR update timestamp cache for %s: %w", langCode, err)
	}

	if !exists {
		return "", nil
	}

	return value, nil
}

func (s *MonitorFileStorage) WriteLastLangPRUpdatedAt(langCode, value string) error {
	if err := s.cacheStore.Write(
		LastLangPRUpdatedAtCacheBucket(langCode),
		LastLangPRUpdatedAtCacheKey(),
		value,
	); err != nil {
		return fmt.Errorf("write PR update timestamp cache for %s: %w", langCode, err)
	}

	return nil
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
	retryDelay := defaultRetryDelaySeconds * time.Second

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

			return fmt.Errorf("interval check context done: %w", ctx.Err())
		case <-timer.C:
		}
	}
}

func (mon *Monitor) RetryCheck(
	ctx context.Context,
	retryDelay time.Duration,
	onUpdateTask OnUpdateTask,
) error {
	for {
		err := mon.Check(ctx, onUpdateTask)
		if err != nil {
			delay := nextRetryDelay(err, retryDelay)

			log.Printf("[githubmon] check failed: %v", err)
			log.Printf("[githubmon] retrying in %s", delay)

			timer := time.NewTimer(delay)
			select {
			case <-ctx.Done():
				if !timer.Stop() {
					<-timer.C
				}

				return fmt.Errorf("retry check context done: %w", ctx.Err())
			case <-timer.C:
			}

			continue
		}

		return nil
	}
}

func nextRetryDelay(err error, baseDelay time.Duration) time.Duration {
	type retryableError interface {
		IsRetryable() bool
	}

	var retErr retryableError
	if errors.As(err, &retErr) && retErr.IsRetryable() {
		return baseDelay
	}

	// fallback delay for non-retryable errors
	return nonRetryableFallbackDelayMinutes * time.Minute
}

func (mon *Monitor) Check(
	ctx context.Context,
	onUpdateTask OnUpdateTask,
) error {
	repoUpdate, err := mon.checkRepoUpdates(ctx)
	if err != nil {
		return err
	}

	lastPRUpdated, lastPRUpdatedAt, prLangChanges, err := mon.checkPRUpdates(ctx)
	if err != nil {
		return err
	}

	if err := mon.runUpdateTask(ctx, onUpdateTask, repoUpdate.Updated, prLangChanges); err != nil {
		return err
	}

	if repoUpdate.Updated {
		if err := mon.storage.WriteLastRepoUpdatedAt(repoUpdate.UpdatedAt); err != nil {
			return fmt.Errorf("write repo update timestamp: %w", err)
		}
	}

	if lastPRUpdated {
		if err := mon.storage.WriteLastPRUpdatedAt(lastPRUpdatedAt); err != nil {
			return fmt.Errorf("write PR update timestamp: %w", err)
		}
	}

	if len(prLangChanges) > 0 {
		if err := mon.writeChangedLangCodesInPR(prLangChanges); err != nil {
			return fmt.Errorf("write language PR update timestamps: %w", err)
		}
	}

	return nil
}

//nolint:exhaustruct
func (mon *Monitor) checkRepoUpdates(ctx context.Context) (repoChange, error) {
	if mon.skipGitChecking {
		return repoChange{}, nil
	}

	repoUpdate, err := mon.isRepoUpdated(ctx)
	if err != nil {
		return repoChange{}, fmt.Errorf("check repo updates: %w", err)
	}

	return repoUpdate, nil
}

func (mon *Monitor) checkPRUpdates(ctx context.Context) (bool, string, []langChange, error) {
	if mon.skipPRChecking {
		return false, "", nil, nil
	}

	// pre-check
	lastPRUpdated, lastPRUpdatedAt, err := mon.changedInPR(ctx)
	if err != nil {
		return false, "", nil, fmt.Errorf("check PR updates: %w", err)
	}

	var prLangChanges []langChange

	if lastPRUpdated {
		log.Println("[githubmon] PR changes found in pre-check")

		prLangChanges, err = mon.changedLangCodesInPR(ctx)
		if err != nil {
			return false, "", nil, fmt.Errorf("check PR language updates: %w", err)
		}
	} else {
		log.Printf("[githubmon] no PR changes since %s", lastPRUpdatedAt)
	}

	return lastPRUpdated, lastPRUpdatedAt, prLangChanges, nil
}

func (mon *Monitor) runUpdateTask(
	ctx context.Context,
	onUpdateTask OnUpdateTask,
	repoUpdated bool,
	prLangChanges []langChange,
) error {
	if !repoUpdated && len(prLangChanges) == 0 {
		return nil
	}

	changedLangCodes := make([]string, 0, len(prLangChanges))
	for _, change := range prLangChanges {
		changedLangCodes = append(changedLangCodes, change.Code)
	}

	if err := onUpdateTask.OnUpdate(ctx, repoUpdated, changedLangCodes); err != nil {
		return fmt.Errorf("run on-update task: %w", err)
	}

	return nil
}

func (mon *Monitor) writeChangedLangCodesInPR(changes []langChange) error {
	for _, change := range changes {
		if err := mon.storage.WriteLastLangPRUpdatedAt(change.Code, change.UpdatedAt); err != nil {
			return fmt.Errorf("write PR update timestamp for %s: %w", change.Code, err)
		}
	}

	return nil
}

type repoChange struct {
	UpdatedAt string
	Updated   bool
}

func (mon *Monitor) isRepoUpdated(ctx context.Context) (repoChange, error) {
	log.Printf("[githubmon] checking repo updates")

	lastUpdatedAt, err := mon.storage.ReadLastRepoUpdatedAt()
	if err != nil {
		return repoChange{}, fmt.Errorf("read repo update timestamp: %w", err)
	}

	currentUpdatedAt, err := mon.getCurrentLastRepoUpdatedAt(ctx)
	if err != nil {
		return repoChange{}, err
	}

	isUpdated := len(lastUpdatedAt) == 0 || lastUpdatedAt != currentUpdatedAt

	if isUpdated {
		log.Printf("[githubmon] repo updated: %s -> %s", lastUpdatedAt, currentUpdatedAt)
	} else {
		log.Printf("[githubmon] repo unchanged since %s", lastUpdatedAt)
	}

	return repoChange{
		UpdatedAt: currentUpdatedAt,
		Updated:   isUpdated,
	}, nil
}

func (mon *Monitor) getCurrentLastRepoUpdatedAt(ctx context.Context) (string, error) {
	commitInfo, err := mon.gitHub.GetLatestCommit(ctx)
	if err != nil {
		return "", fmt.Errorf("get latest commit: %w", err)
	}

	return commitInfo.DateTime, nil
}

type langChange struct {
	Code      string
	UpdatedAt string
	Updated   bool
}

func (mon *Monitor) checkLangChange(ctx context.Context, langCode string) (langChange, error) {
	lastUpdatedAt, err := mon.storage.ReadLastLangPRUpdatedAt(langCode)
	if err != nil {
		return langChange{}, fmt.Errorf("read PR update timestamp for %s: %w", langCode, err)
	}

	currentUpdatedAt, err := mon.getLastLangPRUpdatedAt(ctx, langCode)
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

func (mon *Monitor) changedInPR(ctx context.Context) (bool, string, error) {
	lastUpdatedAt, err := mon.storage.ReadLastPRUpdatedAt()
	if err != nil {
		return false, "", fmt.Errorf("read PR update timestamp: %w", err)
	}

	currentUpdatedAt, err := mon.getLastPRUpdatedAt(ctx)
	if err != nil {
		return false, "", err
	}

	isUpdated := len(currentUpdatedAt) > 0 && lastUpdatedAt != currentUpdatedAt

	return isUpdated, currentUpdatedAt, nil
}

func (mon *Monitor) changedLangCodesInPR(ctx context.Context) ([]langChange, error) {
	log.Printf("[githubmon] checking PR updates for languages")

	langCodes, err := mon.langProvider.LangCodes()
	if err != nil {
		return nil, fmt.Errorf("get language codes: %w", err)
	}

	var updatedLangCodes []langChange

	for _, langCode := range langCodes {
		change, err := mon.checkLangChange(ctx, langCode)
		if err != nil {
			return nil, fmt.Errorf("check PR updates for %s: %w", langCode, err)
		}

		if change.Updated {
			log.Printf("[githubmon][%s] PR updated at %s", langCode, change.UpdatedAt)
			updatedLangCodes = append(updatedLangCodes, change)
		} else {
			log.Printf("[githubmon][%s] no PR updates since %s", langCode, change.UpdatedAt)
		}
	}

	if len(updatedLangCodes) > 0 {
		log.Printf("[githubmon] language PR changes found: %v", updatedLangCodes)
	} else {
		log.Printf("[githubmon] no language PR changes")
	}

	return updatedLangCodes, nil
}

func (mon *Monitor) getLastPRUpdatedAt(ctx context.Context) (string, error) {
	result, err := mon.gitHub.PRSearch(
		ctx,
		//nolint:exhaustruct
		github.PRSearchFilter{},
		github.PageRequest{
			Sort:    "updated",
			Order:   "desc",
			Page:    firstPage,
			PerPage: 1,
		},
	)
	if err != nil {
		return "", fmt.Errorf("search PRs: %w", err)
	}

	if len(result.Items) == 0 {
		return "", nil
	}

	return result.Items[0].UpdatedAt, nil
}

func (mon *Monitor) getLastLangPRUpdatedAt(ctx context.Context, langCode string) (string, error) {
	result, err := mon.gitHub.PRSearch(
		ctx,
		//nolint:exhaustruct
		github.PRSearchFilter{
			LangCode: langCode,
		},
		github.PageRequest{
			Sort:    "updated",
			Order:   "desc",
			Page:    firstPage,
			PerPage: 1,
		},
	)
	if err != nil {
		return "", fmt.Errorf("search PRs for %s: %w", langCode, err)
	}

	if len(result.Items) == 0 {
		return "", nil
	}

	return result.Items[0].UpdatedAt, nil
}
