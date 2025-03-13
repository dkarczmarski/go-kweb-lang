package git

import (
	"bytes"
	"fmt"
	"os/exec"
)

func runCommand(cwd string, command string, args ...string) (string, error) {
	cmd := exec.Command(command, args...)
	cmd.Dir = cwd

	var out bytes.Buffer
	var stderr bytes.Buffer
	cmd.Stdout = &out
	cmd.Stderr = &stderr

	err := cmd.Run()
	if err != nil {
		return "", fmt.Errorf("%s: %w", stderr.String(), err)
	}

	return out.String(), nil
}
