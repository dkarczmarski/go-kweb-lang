//nolint:testpackage,containedctx
package runtime

import (
	"context"
	"time"

	"github.com/dkarczmarski/go-kweb-lang/githubmon"
)

type fakeRepoCreator struct {
	createCalls int
	createCtx   context.Context
	createURL   string
	createErr   error
}

func (f *fakeRepoCreator) Create(ctx context.Context, url string) error {
	f.createCalls++
	f.createCtx = ctx
	f.createURL = url

	return f.createErr
}

type fakeRetryChecker struct {
	retryCalls    int
	retryCtx      context.Context
	retryDelay    time.Duration
	retryTask     githubmon.OnUpdateTask
	retryErr      error
	intervalCalls int
	intervalCtx   context.Context
	intervalDelay time.Duration
	intervalTask  githubmon.OnUpdateTask
	intervalErr   error
	intervalCh    chan struct{}
}

func (f *fakeRetryChecker) RetryCheck(ctx context.Context, delay time.Duration, task githubmon.OnUpdateTask) error {
	f.retryCalls++
	f.retryCtx = ctx
	f.retryDelay = delay
	f.retryTask = task

	return f.retryErr
}

func (f *fakeRetryChecker) IntervalCheck(ctx context.Context, delay time.Duration, task githubmon.OnUpdateTask) error {
	f.intervalCalls++
	f.intervalCtx = ctx
	f.intervalDelay = delay
	f.intervalTask = task

	if f.intervalCh != nil {
		close(f.intervalCh)
	}

	return f.intervalErr
}

type fakeHTTPServer struct {
	listenErr        error
	shutdownErr      error
	listenCalls      int
	shutdownCalls    int
	listenStartedCh  chan struct{}
	shutdownCalledCh chan struct{}
}

func (f *fakeHTTPServer) ListenAndServe() error {
	f.listenCalls++

	if f.listenStartedCh != nil {
		close(f.listenStartedCh)
	}

	return f.listenErr
}

func (f *fakeHTTPServer) Shutdown(_ context.Context) error {
	f.shutdownCalls++

	if f.shutdownCalledCh != nil {
		close(f.shutdownCalledCh)
	}

	return f.shutdownErr
}
