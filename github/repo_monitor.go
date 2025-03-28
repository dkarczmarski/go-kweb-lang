package github

import (
	"context"
	"fmt"
	"log"
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
	log.Printf("checking for repo changes")

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
