package git_test

import (
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"testing"

	"github.com/dkarczmarski/go-kweb-lang/git"
)

type integrationEnv struct {
	tmpDir      string
	scenarioDir string
	gitRepo     *git.Git
}

func newIntegrationEnv(t *testing.T, scenarioName string) integrationEnv {
	t.Helper()

	tmpDir := t.TempDir()
	scenarioDir := scenarioPath(t, scenarioName)

	runScenarioScript(t, tmpDir, scenarioDir, "init.sh")

	repoPath := filepath.Join(tmpDir, "repo")
	gitRepo := git.NewRepo(repoPath)

	return integrationEnv{
		tmpDir:      tmpDir,
		scenarioDir: scenarioDir,
		gitRepo:     gitRepo,
	}
}

//nolint:gochecknoglobals
var integrationScriptDebugOutput = true

func scenarioPath(t *testing.T, scenarioName string) string {
	t.Helper()

	_, thisFile, _, ok := runtime.Caller(0)
	if !ok {
		t.Fatal("failed to determine current test file path")
	}

	baseDir := filepath.Dir(thisFile)

	return filepath.Join(baseDir, "testdata", "scenarios", scenarioName)
}

func runScenarioScript(t *testing.T, workDir, scenarioDir, scriptName string) {
	t.Helper()

	scriptPath := filepath.Join(scenarioDir, scriptName)

	cmd := exec.Command("bash", scriptPath)
	cmd.Dir = workDir

	if integrationScriptDebugOutput {
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
	}

	if err := cmd.Run(); err != nil {
		t.Fatalf("script execution failed (%s): %v", scriptPath, err)
	}
}
