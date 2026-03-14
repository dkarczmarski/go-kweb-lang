//nolint:gosec
package tasks_test

import (
	"io"
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"testing"

	"github.com/dkarczmarski/go-kweb-lang/dashboard"
	"github.com/dkarczmarski/go-kweb-lang/filepairs"
	"github.com/dkarczmarski/go-kweb-lang/git"
	"github.com/dkarczmarski/go-kweb-lang/githist"
	"github.com/dkarczmarski/go-kweb-lang/gitseek"
	"github.com/dkarczmarski/go-kweb-lang/langcnt"
	"github.com/dkarczmarski/go-kweb-lang/pullreq"
	"github.com/dkarczmarski/go-kweb-lang/store"
	"github.com/dkarczmarski/go-kweb-lang/tasks"
	"github.com/dkarczmarski/go-kweb-lang/web"
)

const renderTestHTMLEnvName = "RENDER_TEST_HTML"

type refreshDashboardRenderEnv struct {
	tmpDir         string
	scenarioDir    string
	dashboardStore *dashboard.Store
	task           *tasks.RefreshDashboardTask
}

func TestRefreshDashboardTask_Run_RenderHTML_Integration(t *testing.T) {
	t.Parallel()

	ctx := t.Context()

	env := newRefreshDashboardRenderEnv(
		t,
		"multiple_en_updates_on_merged_branch_after_lang",
		map[string]pullreq.FilePRIndexData{
			"pl": {
				"content/pl/docs/test.md":    {101, 102},
				"content/pl/docs/missing.md": {999},
			},
		},
	)

	if err := env.task.Run(ctx); err != nil {
		t.Fatalf("RefreshDashboardTask.Run returned error: %v", err)
	}

	langDashboard, err := env.dashboardStore.ReadDashboard("pl")
	if err != nil {
		t.Fatalf("ReadDashboard returned error: %v", err)
	}

	if langDashboard.LangCode != "pl" {
		t.Fatalf("expected LangCode pl, got %q", langDashboard.LangCode)
	}

	if len(langDashboard.Items) != 2 {
		t.Fatalf("expected 2 dashboard items, got %d", len(langDashboard.Items))
	}

	indexHTML := renderResponseBody(
		t,
		env.dashboardStore,
		http.MethodGet,
		"/",
		"",
	)
	assertContainsAll(
		t,
		string(indexHTML),
		"Lang",
		`href="/lang/pl"`,
	)
	writeRenderedHTMLIfEnabled(t, "dashboard-index.html", indexHTML)

	pageHTML := renderResponseBody(
		t,
		env.dashboardStore,
		http.MethodGet,
		"/lang/pl",
		"",
	)
	assertContainsAll(
		t,
		string(pageHTML),
		"<h3>pl</h3>",
		"content/pl/docs/test.md",
		"content/pl/docs/missing.md",
		"waiting-for-review",
		"#101",
		"#102",
		"#999",
		"Merge branch &#39;en-branch&#39;",
		"2020-01-06",
	)
	writeRenderedHTMLIfEnabled(t, "dashboard-pl-full.html", pageHTML)

	tableHTML := renderResponseBody(
		t,
		env.dashboardStore,
		http.MethodPost,
		"/lang/pl",
		url.Values{
			"itemsType": []string{"with-update-or-pr"},
			"filepath":  []string{"content/pl/docs"},
		}.Encode(),
	)
	assertContainsAll(
		t,
		string(tableHTML),
		"<table",
		"content/pl/docs/test.md",
		"content/pl/docs/missing.md",
		"#101",
		"#999",
	)
	writeRenderedHTMLIfEnabled(t, "dashboard-pl-table.html", tableHTML)

	filteredHTML := renderResponseBody(
		t,
		env.dashboardStore,
		http.MethodPost,
		"/lang/pl",
		url.Values{
			"itemsType": []string{"with-pr"},
			"filepath":  []string{"missing.md"},
		}.Encode(),
	)
	assertContainsAll(
		t,
		string(filteredHTML),
		"content/pl/docs/missing.md",
		"waiting-for-review",
		"#999",
	)
	writeRenderedHTMLIfEnabled(t, "dashboard-pl-filtered-missing.html", filteredHTML)
}

func newRefreshDashboardRenderEnv(
	t *testing.T,
	scenarioName string,
	prIndexByLang map[string]pullreq.FilePRIndexData,
) refreshDashboardRenderEnv {
	t.Helper()

	tmpDir := t.TempDir()
	scenarioDir := scenarioPath(t, scenarioName)

	runScenarioScript(t, tmpDir, scenarioDir, "init.sh")

	repoPath := filepath.Join(tmpDir, "repo")
	cacheDir := filepath.Join(tmpDir, "cache")

	gitRepo := git.NewRepo(repoPath)
	cacheStore := store.NewFileStore(cacheDir)
	gitRepoHist := githist.New(gitRepo, cacheStore)
	gitSeeker := gitseek.New(gitRepo, gitRepoHist, cacheStore)

	langCodesProvider := &langcnt.LangCodesProvider{
		RepoDir: repoPath,
	}

	contentPairProvider := filepairs.NewContentPairProvider(gitRepo)
	pairProviders := filepairs.NewPairProviders(contentPairProvider)

	dashboardStore := dashboard.NewStore(cacheStore)

	task := tasks.NewRefreshDashboardTask(
		langCodesProvider,
		pairProviders,
		gitSeeker,
		fakeFilePRIndex{data: prIndexByLang},
		dashboardStore,
	)

	return refreshDashboardRenderEnv{
		tmpDir:         tmpDir,
		scenarioDir:    scenarioDir,
		dashboardStore: dashboardStore,
		task:           task,
	}
}

type fakeFilePRIndex struct {
	data map[string]pullreq.FilePRIndexData
}

func (f fakeFilePRIndex) LangIndex(langCode string) (pullreq.FilePRIndexData, error) {
	if index, ok := f.data[langCode]; ok {
		return index, nil
	}

	return pullreq.FilePRIndexData{}, nil
}

func renderResponseBody(
	t *testing.T,
	dashboardStore *dashboard.Store,
	method string,
	targetURL string,
	formBody string,
) []byte {
	t.Helper()

	handler := web.NewHandler(dashboardStore)
	mux := http.NewServeMux()
	handler.Register(mux)

	var bodyReader io.Reader
	if formBody != "" {
		bodyReader = strings.NewReader(formBody)
	}

	request := httptest.NewRequest(method, targetURL, bodyReader)
	if formBody != "" {
		request.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	}

	recorder := httptest.NewRecorder()
	mux.ServeHTTP(recorder, request)

	response := recorder.Result()
	defer response.Body.Close()

	responseBody, err := io.ReadAll(response.Body)
	if err != nil {
		t.Fatalf("read response body: %v", err)
	}

	if response.StatusCode != http.StatusOK {
		t.Fatalf(
			"unexpected status code %d for %s %s\nbody:\n%s",
			response.StatusCode,
			method,
			targetURL,
			string(responseBody),
		)
	}

	return responseBody
}

func writeRenderedHTMLIfEnabled(t *testing.T, fileName string, htmlContent []byte) {
	t.Helper()

	if os.Getenv(renderTestHTMLEnvName) != "1" {
		return
	}

	outputDir := testOutputDir(t)

	if err := os.MkdirAll(outputDir, 0o755); err != nil {
		t.Fatalf("create output dir %s: %v", outputDir, err)
	}

	outputPath := filepath.Join(outputDir, fileName)

	if err := os.WriteFile(outputPath, htmlContent, 0o644); err != nil {
		t.Fatalf("write rendered html %s: %v", outputPath, err)
	}

	t.Logf("rendered html written to %s", outputPath)
}

func testOutputDir(t *testing.T) string {
	t.Helper()

	_, thisFile, _, ok := runtime.Caller(0)
	if !ok {
		t.Fatal("failed to determine current test file path")
	}

	tasksDir := filepath.Dir(thisFile)
	repoRootDir := filepath.Dir(tasksDir)

	return filepath.Join(repoRootDir, "test-output")
}

func assertContainsAll(t *testing.T, text string, fragments ...string) {
	t.Helper()

	for _, fragment := range fragments {
		if !strings.Contains(text, fragment) {
			t.Fatalf("expected output to contain %q\n\nfull output:\n%s", fragment, text)
		}
	}
}

func scenarioPath(t *testing.T, scenarioName string) string {
	t.Helper()

	_, thisFile, _, ok := runtime.Caller(0)
	if !ok {
		t.Fatal("failed to determine current test file path")
	}

	tasksDir := filepath.Dir(thisFile)
	repoRootDir := filepath.Dir(tasksDir)

	return filepath.Join(
		repoRootDir,
		"gitseek",
		"testdata",
		"scenarios",
		scenarioName,
	)
}

func runScenarioScript(t *testing.T, workDir string, scenarioDir string, scriptName string) {
	t.Helper()

	scriptPath := filepath.Join(scenarioDir, scriptName)

	command := exec.Command("bash", scriptPath)
	command.Dir = workDir
	command.Stdout = os.Stdout
	command.Stderr = os.Stderr

	if err := command.Run(); err != nil {
		t.Fatalf("script execution failed (%s): %v", scriptPath, err)
	}
}
