package internal_test

import (
	"context"
	"go-kweb-lang/git/internal"
	"strings"
	"testing"
)

func TestStdCommandRunner_Exec(t *testing.T) {
	runner := &internal.StdCommandRunner{}

	for _, tc := range []struct {
		name           string
		workingDir     string
		command        string
		args           []string
		expectedErr    func(err error) bool
		expectedOutput func(out string) bool
	}{
		{
			name:        "working directory /",
			workingDir:  "/",
			command:     "pwd",
			expectedErr: noError,
			expectedOutput: func(out string) bool {
				return strings.TrimSpace(out) == "/"
			},
		},
		{
			name:        "working directory /tmp",
			workingDir:  "/tmp",
			command:     "pwd",
			expectedErr: noError,
			expectedOutput: func(out string) bool {
				return strings.TrimSpace(out) == "/tmp"
			},
		},
		{
			name:        "echo 1 2 3",
			workingDir:  ".",
			command:     "echo",
			args:        []string{"1", "2", "3"},
			expectedErr: noError,
			expectedOutput: func(out string) bool {
				return strings.TrimSpace(out) == "1 2 3"
			},
		},
		{
			name:           "not existent command",
			workingDir:     ".",
			command:        "/fake-command",
			args:           []string{"1", "2", "3"},
			expectedErr:    withError,
			expectedOutput: emptyOutput,
		},
		{
			name:           "command fails",
			workingDir:     ".",
			command:        "false",
			expectedErr:    withError,
			expectedOutput: emptyOutput,
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			out, err := runner.Exec(context.Background(), tc.workingDir, tc.command, tc.args...)
			if !tc.expectedErr(err) {
				t.Errorf("unexpected error: %v", err)
			}
			if !tc.expectedOutput(out) {
				t.Errorf("unexpected output: %v", out)
			}
		})
	}
}

func noError(err error) bool {
	return err == nil
}

func withError(err error) bool {
	return err != nil
}

func emptyOutput(out string) bool {
	return len(strings.TrimSpace(out)) == 0
}
