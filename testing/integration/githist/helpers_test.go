package githist_test

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	"github.com/dkarczmarski/go-kweb-lang/git"
	"github.com/dkarczmarski/go-kweb-lang/githist"
	"github.com/dkarczmarski/go-kweb-lang/store"
)

type gitHistIntegrationEnv struct {
	tmpDir      string
	repoDir     string
	cacheDir    string
	scenarioDir string
	gitRepo     *git.Git
	gitHist     *githist.GitHist
}

func newGitHistIntegrationEnv(t *testing.T, scenarioName string) gitHistIntegrationEnv {
	t.Helper()

	tmpDir := t.TempDir()

	cwd, err := os.Getwd()
	if err != nil {
		t.Fatalf("failed to get working directory: %v", err)
	}

	scenarioDir := filepath.Join(cwd, "testdata", scenarioName)
	repoDir := filepath.Join(tmpDir, "repo")
	cacheDir := filepath.Join(tmpDir, "cache")

	runScenarioScript(t, tmpDir, scenarioDir, "init.sh")

	gitRepo := git.NewRepo(repoDir)
	cache := store.NewFileStore(cacheDir)
	gitRepoHist := githist.New(gitRepo, cache)

	return gitHistIntegrationEnv{
		tmpDir:      tmpDir,
		repoDir:     repoDir,
		cacheDir:    cacheDir,
		scenarioDir: scenarioDir,
		gitRepo:     gitRepo,
		gitHist:     gitRepoHist,
	}
}

func runScenarioScript(t *testing.T, workDir, scenarioDir, scriptName string) {
	t.Helper()

	scriptPath, err := filepath.Abs(filepath.Join(scenarioDir, scriptName))
	if err != nil {
		t.Fatalf("failed to build absolute script path: %v", err)
	}

	cmd := exec.Command("bash", scriptPath)
	cmd.Dir = workDir
	cmd.Env = append(os.Environ(), "HOME="+workDir)

	output, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("failed to run scenario script %s: %v\noutput:\n%s", scriptPath, err, string(output))
	}
}

func mustGitRevParseBySubject(t *testing.T, repoDir, subject string) string {
	t.Helper()

	cmd := exec.Command("git", "log", "--all", "--format=%H %s")
	cmd.Dir = repoDir

	output, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("git log failed: %v\noutput:\n%s", err, string(output))
	}

	lines := strings.Split(strings.TrimSpace(string(output)), "\n")
	for _, line := range lines {
		if strings.HasSuffix(line, " "+subject) {
			parts := strings.SplitN(line, " ", 2)
			if len(parts) != 2 {
				t.Fatalf("unexpected git log line format: %q", line)
			}

			return parts[0]
		}
	}

	t.Fatalf("commit with subject %q not found", subject)

	return ""
}

type gitHistRemoteIntegrationEnv struct {
	tmpDir      string
	scenarioDir string
	repoDir     string
	updaterDir  string
	originDir   string
	cacheDir    string
	gitRepo     *git.Git
	gitHist     *githist.GitHist
}

func newGitHistRemoteIntegrationEnv(t *testing.T, scenarioName string) gitHistRemoteIntegrationEnv {
	t.Helper()

	tmpDir := t.TempDir()

	cwd, err := os.Getwd()
	if err != nil {
		t.Fatalf("failed to get working directory: %v", err)
	}

	scenarioDir := filepath.Join(cwd, "testdata", scenarioName)
	repoDir := filepath.Join(tmpDir, "repo")
	updaterDir := filepath.Join(tmpDir, "updater")
	originDir := filepath.Join(tmpDir, "origin.git")
	cacheDir := filepath.Join(tmpDir, "cache")

	runScenarioScript(t, tmpDir, scenarioDir, "init.sh")

	gitRepo := git.NewRepo(repoDir)
	cache := store.NewFileStore(cacheDir)
	gitRepoHist := githist.New(gitRepo, cache)

	return gitHistRemoteIntegrationEnv{
		tmpDir:      tmpDir,
		scenarioDir: scenarioDir,
		repoDir:     repoDir,
		updaterDir:  updaterDir,
		originDir:   originDir,
		cacheDir:    cacheDir,
		gitRepo:     gitRepo,
		gitHist:     gitRepoHist,
	}
}

func mustGitHeadSubject(t *testing.T, repoDir string) string {
	t.Helper()

	cmd := exec.Command("git", "log", "-1", "--format=%s")
	cmd.Dir = repoDir

	output, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("git log -1 failed: %v\noutput:\n%s", err, string(output))
	}

	return strings.TrimSpace(string(output))
}
