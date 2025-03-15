package github

import (
	"context"
	"fmt"
	"log"
)

type Monitor struct {
	gh    *GitHub
	tasks []OnUpdateTask

	// todo: sync
	lastCommitID string
}

func NewMonitor(gh *GitHub, tasks []OnUpdateTask) *Monitor {
	return &Monitor{
		gh:    gh,
		tasks: tasks,
	}
}

type OnUpdateTask interface {
	Run() error
}

func (mon *Monitor) Check() error {
	commitInfo, err := mon.gh.GetLatestCommit(context.TODO())
	if err != nil {
		return fmt.Errorf("GitHub get latest commit error: %w", err)
	}

	if len(mon.lastCommitID) == 0 || commitInfo.CommitID != mon.lastCommitID {
		mon.lastCommitID = commitInfo.CommitID

		for _, task := range mon.tasks {
			if err := task.Run(); err != nil {
				log.Printf("task error: %v", err)
			}
		}
	}

	return nil
}
