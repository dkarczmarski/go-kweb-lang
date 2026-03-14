//nolint:testpackage
package runtime

import (
	"context"
	"errors"
	"net/http"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/dkarczmarski/go-kweb-lang/githubmon"
)

func TestRunCheckAndRefresh_RunOnce_CallsRetryCheck(t *testing.T) {
	t.Parallel()

	repoDir := t.TempDir()
	if err := os.Mkdir(filepath.Join(repoDir, ".git"), 0o755); err != nil {
		t.Fatal(err)
	}

	repo := &fakeRepoCreator{}
	monitor := &fakeRetryChecker{}

	var refreshTask githubmon.OnUpdateTask

	err := runCheckAndRefresh(
		t.Context(),
		repoDir,
		false,
		true,
		0,
		repo,
		monitor,
		refreshTask,
	)
	if err != nil {
		t.Fatalf("runCheckAndRefresh returned error: %v", err)
	}

	if repo.createCalls != 0 {
		t.Fatalf("expected repo.Create not to be called, got %d", repo.createCalls)
	}

	if monitor.retryCalls != 1 {
		t.Fatalf("expected RetryCheck to be called once, got %d", monitor.retryCalls)
	}

	if monitor.retryDelay != defaultRetryDelay {
		t.Fatalf("unexpected retry delay: got %v, want %v", monitor.retryDelay, defaultRetryDelay)
	}

	if monitor.retryTask != nil {
		t.Fatal("expected nil task passed to RetryCheck")
	}

	if monitor.intervalCalls != 0 {
		t.Fatalf("expected IntervalCheck not to be called, got %d", monitor.intervalCalls)
	}
}

func TestRunCheckAndRefresh_RunOnce_ReturnsRetryError(t *testing.T) {
	t.Parallel()

	repoDir := t.TempDir()
	if err := os.Mkdir(filepath.Join(repoDir, ".git"), 0o755); err != nil {
		t.Fatal(err)
	}

	wantErr := errors.New("retry failed")
	monitor := &fakeRetryChecker{retryErr: wantErr}

	var refreshTask githubmon.OnUpdateTask

	err := runCheckAndRefresh(
		t.Context(),
		repoDir,
		false,
		true,
		0,
		&fakeRepoCreator{},
		monitor,
		refreshTask,
	)
	if err == nil {
		t.Fatal("expected error, got nil")
	}

	if !errors.Is(err, wantErr) {
		t.Fatalf("expected wrapped retry error, got %v", err)
	}
}

func TestRunCheckAndRefresh_IntervalMode_CallsIntervalCheck(t *testing.T) {
	t.Parallel()

	repoDir := t.TempDir()
	if err := os.Mkdir(filepath.Join(repoDir, ".git"), 0o755); err != nil {
		t.Fatal(err)
	}

	ch := make(chan struct{})
	monitor := &fakeRetryChecker{intervalCh: ch}

	var refreshTask githubmon.OnUpdateTask

	err := runCheckAndRefresh(
		t.Context(),
		repoDir,
		false,
		false,
		7,
		&fakeRepoCreator{},
		monitor,
		refreshTask,
	)
	if err != nil {
		t.Fatalf("runCheckAndRefresh returned error: %v", err)
	}

	select {
	case <-ch:
	case <-time.After(500 * time.Millisecond):
		t.Fatal("IntervalCheck was not called")
	}

	if monitor.intervalCalls != 1 {
		t.Fatalf("expected IntervalCheck to be called once, got %d", monitor.intervalCalls)
	}

	wantDelay := 7 * time.Minute
	if monitor.intervalDelay != wantDelay {
		t.Fatalf("unexpected interval delay: got %v, want %v", monitor.intervalDelay, wantDelay)
	}

	if monitor.intervalTask != nil {
		t.Fatal("expected nil task passed to IntervalCheck")
	}
}

func TestRunCheckAndRefresh_CreatesRepoWhenMissing(t *testing.T) {
	t.Parallel()

	repoDir := t.TempDir()
	repo := &fakeRepoCreator{}
	monitor := &fakeRetryChecker{}

	var refreshTask githubmon.OnUpdateTask

	err := runCheckAndRefresh(
		t.Context(),
		repoDir,
		false,
		false,
		0,
		repo,
		monitor,
		refreshTask,
	)
	if err != nil {
		t.Fatalf("runCheckAndRefresh returned error: %v", err)
	}

	if repo.createCalls != 1 {
		t.Fatalf("expected repo.Create to be called once, got %d", repo.createCalls)
	}

	if repo.createURL != repoURL {
		t.Fatalf("unexpected repo URL: got %q, want %q", repo.createURL, repoURL)
	}
}

func TestRunCheckAndRefresh_SkipGitChecking_DoesNotCreateRepo(t *testing.T) {
	t.Parallel()

	repo := &fakeRepoCreator{}
	monitor := &fakeRetryChecker{}

	var refreshTask githubmon.OnUpdateTask

	err := runCheckAndRefresh(
		t.Context(),
		t.TempDir(),
		true,
		false,
		0,
		repo,
		monitor,
		refreshTask,
	)
	if err != nil {
		t.Fatalf("runCheckAndRefresh returned error: %v", err)
	}

	if repo.createCalls != 0 {
		t.Fatalf("expected repo.Create not to be called, got %d", repo.createCalls)
	}
}

func TestCreateRepoIfNotExists_CreatesRepo(t *testing.T) {
	t.Parallel()

	repoDir := filepath.Join(t.TempDir(), "repo")
	repo := &fakeRepoCreator{}

	err := createRepoIfNotExists(t.Context(), repoDir, repo)
	if err != nil {
		t.Fatalf("createRepoIfNotExists returned error: %v", err)
	}

	if repo.createCalls != 1 {
		t.Fatalf("expected repo.Create to be called once, got %d", repo.createCalls)
	}

	if _, err := os.Stat(repoDir); err != nil {
		t.Fatalf("expected repo dir to exist, stat error: %v", err)
	}
}

func TestCreateRepoIfNotExists_DoesNothingWhenGitDirExists(t *testing.T) {
	t.Parallel()

	repoDir := t.TempDir()
	if err := os.Mkdir(filepath.Join(repoDir, ".git"), 0o755); err != nil {
		t.Fatal(err)
	}

	repo := &fakeRepoCreator{}

	err := createRepoIfNotExists(t.Context(), repoDir, repo)
	if err != nil {
		t.Fatalf("createRepoIfNotExists returned error: %v", err)
	}

	if repo.createCalls != 0 {
		t.Fatalf("expected repo.Create not to be called, got %d", repo.createCalls)
	}
}

func TestRunWebServer_ReturnsListenError(t *testing.T) {
	t.Parallel()

	srv := &fakeHTTPServer{
		listenErr: errors.New("listen failed"),
	}

	err := runWebServer(t.Context(), srv)
	if err == nil {
		t.Fatal("expected error, got nil")
	}
}

func TestRunWebServer_IgnoresErrServerClosed(t *testing.T) {
	t.Parallel()

	srv := &fakeHTTPServer{
		listenErr: http.ErrServerClosed,
	}

	err := runWebServer(t.Context(), srv)
	if err != nil {
		t.Fatalf("expected nil error, got %v", err)
	}
}

func TestRunWebServer_ShutsDownOnContextCancel(t *testing.T) {
	t.Parallel()

	ctx, cancel := context.WithCancel(t.Context())
	srv := &fakeHTTPServer{
		listenStartedCh:  make(chan struct{}),
		shutdownCalledCh: make(chan struct{}),
	}

	go func() {
		<-srv.listenStartedCh
		cancel()
	}()

	errCh := make(chan error, 1)
	go func() {
		errCh <- runWebServer(ctx, srv)
	}()

	select {
	case <-srv.shutdownCalledCh:
	case <-time.After(500 * time.Millisecond):
		t.Fatal("Shutdown was not called")
	}

	select {
	case err := <-errCh:
		if err != nil {
			t.Fatalf("expected nil error, got %v", err)
		}
	case <-time.After(500 * time.Millisecond):
		t.Fatal("runWebServer did not return")
	}
}

func TestRunWebServer_NilServerWaitsForContextCancel(t *testing.T) {
	t.Parallel()

	ctx, cancel := context.WithCancel(t.Context())
	errCh := make(chan error, 1)

	go func() {
		errCh <- runWebServer(ctx, nil)
	}()

	cancel()

	select {
	case err := <-errCh:
		if err != nil {
			t.Fatalf("expected nil error, got %v", err)
		}
	case <-time.After(500 * time.Millisecond):
		t.Fatal("runWebServer did not return")
	}
}

func TestFileExists(t *testing.T) {
	t.Parallel()

	path := filepath.Join(t.TempDir(), "file.txt")
	if err := os.WriteFile(path, []byte("x"), 0o600); err != nil {
		t.Fatal(err)
	}

	exists, err := fileExists(path)
	if err != nil {
		t.Fatalf("fileExists returned error: %v", err)
	}

	if !exists {
		t.Fatal("expected file to exist")
	}
}

func TestFileExists_Missing(t *testing.T) {
	t.Parallel()

	exists, err := fileExists(filepath.Join(t.TempDir(), "missing.txt"))
	if err != nil {
		t.Fatalf("fileExists returned error: %v", err)
	}

	if exists {
		t.Fatal("expected file to not exist")
	}
}
