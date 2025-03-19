package internal

import (
	"bytes"
	"context"
	"fmt"
	"os/exec"
)

type StdCommandRunner struct {
}

func (cr *StdCommandRunner) Exec(ctx context.Context, workingDir string, command string, args ...string) (string, error) {
	var out bytes.Buffer
	var stderr bytes.Buffer

	cmd := exec.CommandContext(ctx, command, args...)
	cmd.Dir = workingDir
	cmd.Stdout = &out
	cmd.Stderr = &stderr

	err := cmd.Run()
	if err != nil {
		return "", fmt.Errorf("%s: %w", stderr.String(), err)
	}

	return out.String(), nil
}
