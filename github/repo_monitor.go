package github

import (
	"context"
	"fmt"
	"log"
	"time"
)

type RepoMonitor struct {
	gh    GitHub
	tasks []OnUpdateTask

	// todo: sync
	lastCommitID string
}

func NewRepoMonitor(gh GitHub, tasks []OnUpdateTask) *RepoMonitor {
	return &RepoMonitor{
		gh:    gh,
		tasks: tasks,
	}
}

type OnUpdateTask interface {
	Run(ctx context.Context) error
}

func (mon *RepoMonitor) CheckRepo(ctx context.Context) error {
	commitInfo, err := mon.gh.GetLatestCommit(ctx)
	if err != nil {
		return fmt.Errorf("GitHub get latest commit error: %w", err)
	}

	if len(mon.lastCommitID) == 0 || commitInfo.CommitID != mon.lastCommitID {
		mon.lastCommitID = commitInfo.CommitID

		for _, task := range mon.tasks {
			if err := task.Run(ctx); err != nil {
				log.Printf("task error: %v", err)
			}
		}
	}

	return nil
}

func (mon *RepoMonitor) RepeatCheckRepo(ctx context.Context, delay time.Duration) {
	for {
		if err := mon.CheckRepo(ctx); err != nil {
			log.Printf("error while checking github for changes: %v", err)
		}

		select {
		case <-ctx.Done():
			return
		case <-time.After(delay):
		}
	}
}
