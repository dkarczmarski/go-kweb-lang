package gitseek_test

import (
	"os"
	"os/exec"
	"path/filepath"
	"reflect"
	"runtime"
	"testing"

	"github.com/dkarczmarski/go-kweb-lang/git"
	"github.com/dkarczmarski/go-kweb-lang/githist"
	"github.com/dkarczmarski/go-kweb-lang/gitseek"
	"github.com/dkarczmarski/go-kweb-lang/store"
)

type integrationEnv struct {
	tmpDir      string
	scenarioDir string
	cache       *store.JSONFileStore
	gitRepoHist *githist.GitHist
	gitSeeker   *gitseek.GitSeek
	pair        gitseek.Pair
}

func newIntegrationEnv(t *testing.T, scenarioName string) integrationEnv {
	t.Helper()

	tmpDir := t.TempDir()
	scenarioDir := scenarioPath(t, scenarioName)

	runScenarioScript(t, tmpDir, scenarioDir, "init.sh")

	repoPath := filepath.Join(tmpDir, "repo")
	cacheDir := filepath.Join(tmpDir, "cache")

	gitRepo := git.NewRepo(repoPath)
	cache := store.NewFileStore(cacheDir)

	gitRepoHist := githist.New(gitRepo, cache)
	gitSeeker := gitseek.New(gitRepo, gitRepoHist, cache)

	return integrationEnv{
		tmpDir:      tmpDir,
		scenarioDir: scenarioDir,
		cache:       cache,
		gitRepoHist: gitRepoHist,
		gitSeeker:   gitSeeker,
		pair: gitseek.Pair{
			EnPath:   "content/en/docs/test.md",
			LangPath: "content/pl/docs/test.md",
		},
	}
}

func assertCachedFileInfo(t *testing.T, cache *store.JSONFileStore, pair gitseek.Pair, expected gitseek.FileInfo) {
	t.Helper()

	var cached gitseek.FileInfo

	exists, err := cache.Read(gitseek.FileInfoCacheBucket("pl"), pair.LangPath, &cached)
	if err != nil {
		t.Fatalf("cache.Read returned error: %v", err)
	}

	if !exists {
		t.Fatal("expected cache entry to exist")
	}

	if !reflect.DeepEqual(expected, cached) {
		t.Fatalf("cached value differs:\n got:  %#v\nwant: %#v", cached, expected)
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

func assertEqualFileInfo(t *testing.T, expected, actual gitseek.FileInfo) {
	t.Helper()

	if !reflect.DeepEqual(expected, actual) {
		t.Fatalf("unexpected FileInfo:\n got:  %#v\nwant: %#v", actual, expected)
	}
}
