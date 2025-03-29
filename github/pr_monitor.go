package github

import (
	"context"
	"fmt"
	"go-kweb-lang/langcnt"
	"go-kweb-lang/proxycache"
	"log"
)

type PRMonitor struct {
	gh          GitHub
	cacheDir    string
	langContent *langcnt.Content
	tasks       []OnPRUpdateTask
}

type OnPRUpdateTask interface {
	Run(ctx context.Context, langCode string) error
}

func NewPRMonitor(
	gh GitHub,
	cacheDir string,
	langContent *langcnt.Content,
	tasks []OnPRUpdateTask,
) *PRMonitor {
	return &PRMonitor{
		gh:          gh,
		cacheDir:    cacheDir,
		langContent: langContent,
		tasks:       tasks,
	}
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

func (mon *PRMonitor) lastMaxUpdatedAt(langCode string) (string, error) {
	return proxycache.Get(
		context.Background(),
		mon.cacheDir,
		"github-last-lang-updated-at",
		langCode,
		nil,
		func(ctx context.Context) (string, error) {
			return "", nil
		},
	)
}

func (mon *PRMonitor) setLastMaxUpdatedAt(maxUpdatedAt, langCode string) error {
	return proxycache.Put(
		mon.cacheDir,
		"github-lang-last-updated-at",
		langCode,
		maxUpdatedAt,
	)
}

func (mon *PRMonitor) CheckLang(ctx context.Context, langCode string) (bool, error) {
	lastMaxUpdatedAt, err := mon.lastMaxUpdatedAt(langCode)
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
		if err := task.Run(ctx, langCode); err != nil {
			log.Printf("task error: %v", err)
		}
	}

	if err := mon.setLastMaxUpdatedAt(maxUpdatedAt, langCode); err != nil {
		return true, fmt.Errorf("error while saving the last PRs: %w", err)
	}

	return true, nil
}

func (mon *PRMonitor) Check(ctx context.Context) error {
	log.Printf("checking for PR changes")

	langs, err := mon.langContent.Langs()
	if err != nil {
		return fmt.Errorf("error while getting available languages: %w", err)
	}

	for _, lang := range langs {
		if _, err := mon.CheckLang(ctx, lang); err != nil {
			return fmt.Errorf("error while checking github for changes: %v", err)
		}

		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
		}
	}

	return nil
}
