//nolint:nilnil,goconst,dupl
package gitseek_test

import (
	"context"
	"errors"
	"reflect"
	"testing"

	"github.com/dkarczmarski/go-kweb-lang/git"
	"github.com/dkarczmarski/go-kweb-lang/gitseek"
)

type fakeGitRepo struct {
	findFileLastCommitFunc   func(ctx context.Context, path string) (git.CommitInfo, error)
	findFileCommitsAfterFunc func(ctx context.Context, path string, commitIDFrom string) ([]git.CommitInfo, error)
	fileExistsFunc           func(path string) (bool, error)

	findFileLastCommitCalls   []string
	findFileCommitsAfterCalls []findFileCommitsAfterCall
	fileExistsCalls           []string
}

type findFileCommitsAfterCall struct {
	Path         string
	CommitIDFrom string
}

func (f *fakeGitRepo) FindFileLastCommit(ctx context.Context, path string) (git.CommitInfo, error) {
	f.findFileLastCommitCalls = append(f.findFileLastCommitCalls, path)

	if f.findFileLastCommitFunc == nil {
		return git.CommitInfo{}, errors.New("unexpected call to FindFileLastCommit")
	}

	return f.findFileLastCommitFunc(ctx, path)
}

func (f *fakeGitRepo) FindFileCommitsAfter(
	ctx context.Context,
	path string,
	commitIDFrom string,
) ([]git.CommitInfo, error) {
	f.findFileCommitsAfterCalls = append(f.findFileCommitsAfterCalls, findFileCommitsAfterCall{
		Path:         path,
		CommitIDFrom: commitIDFrom,
	})

	if f.findFileCommitsAfterFunc == nil {
		return nil, errors.New("unexpected call to FindFileCommitsAfter")
	}

	return f.findFileCommitsAfterFunc(ctx, path, commitIDFrom)
}

func (f *fakeGitRepo) FileExists(path string) (bool, error) {
	f.fileExistsCalls = append(f.fileExistsCalls, path)

	if f.fileExistsFunc == nil {
		return false, errors.New("unexpected call to FileExists")
	}

	return f.fileExistsFunc(path)
}

type fakeGitRepoHist struct {
	findForkCommitFunc  func(ctx context.Context, commitID string) (*git.CommitInfo, error)
	findMergeCommitFunc func(ctx context.Context, commitID string) (*git.CommitInfo, error)

	findForkCommitCalls  []string
	findMergeCommitCalls []string
}

func (f *fakeGitRepoHist) FindForkCommit(ctx context.Context, commitID string) (*git.CommitInfo, error) {
	f.findForkCommitCalls = append(f.findForkCommitCalls, commitID)

	if f.findForkCommitFunc == nil {
		return nil, errors.New("unexpected call to FindForkCommit")
	}

	return f.findForkCommitFunc(ctx, commitID)
}

func (f *fakeGitRepoHist) FindMergeCommit(ctx context.Context, commitID string) (*git.CommitInfo, error) {
	f.findMergeCommitCalls = append(f.findMergeCommitCalls, commitID)

	if f.findMergeCommitFunc == nil {
		return nil, errors.New("unexpected call to FindMergeCommit")
	}

	return f.findMergeCommitFunc(ctx, commitID)
}

type fakeCacheStorage struct {
	readFunc   func(bucket, key string, buff any) (bool, error)
	writeFunc  func(bucket, key string, data any) error
	deleteFunc func(bucket, key string) error

	readCalls   []cacheReadCall
	writeCalls  []cacheWriteCall
	deleteCalls []cacheKey
}

type cacheKey struct {
	Bucket string
	Key    string
}

type cacheReadCall struct {
	Bucket string
	Key    string
	Buff   any
}

type cacheWriteCall struct {
	Bucket string
	Key    string
	Data   any
}

func (f *fakeCacheStorage) Read(bucket, key string, buff any) (bool, error) {
	f.readCalls = append(f.readCalls, cacheReadCall{
		Bucket: bucket,
		Key:    key,
		Buff:   buff,
	})

	if f.readFunc == nil {
		return false, errors.New("unexpected call to Read")
	}

	return f.readFunc(bucket, key, buff)
}

func (f *fakeCacheStorage) Write(bucket, key string, data any) error {
	f.writeCalls = append(f.writeCalls, cacheWriteCall{
		Bucket: bucket,
		Key:    key,
		Data:   data,
	})

	if f.writeFunc == nil {
		return errors.New("unexpected call to Write")
	}

	return f.writeFunc(bucket, key, data)
}

func (f *fakeCacheStorage) Delete(bucket, key string) error {
	f.deleteCalls = append(f.deleteCalls, cacheKey{Bucket: bucket, Key: key})

	if f.deleteFunc == nil {
		return errors.New("unexpected call to Delete")
	}

	return f.deleteFunc(bucket, key)
}

func TestFileInfoCacheBucket(t *testing.T) {
	t.Parallel()

	got := gitseek.FileInfoCacheBucket("pl")
	want := "lang/pl/git-file-info"

	if got != want {
		t.Fatalf("unexpected bucket: got %q, want %q", got, want)
	}
}

func TestGitSeek_CheckLang_ReturnsCachedValue(t *testing.T) {
	t.Parallel()

	expected := gitseek.FileInfo{
		LangPath:   "content/pl/foo.md",
		FileStatus: gitseek.StatusEnFileUpdated,
		EnUpdates: []gitseek.EnUpdate{
			{
				Commit: git.CommitInfo{CommitID: "en-1"},
			},
		},
	}

	cache := &fakeCacheStorage{
		readFunc: func(_, _ string, buff any) (bool, error) {
			ptr, ok := buff.(*gitseek.FileInfo)
			if !ok {
				t.Fatalf("unexpected buff type: %T", buff)
			}
			*ptr = expected

			return true, nil
		},
		writeFunc: func(_, _ string, _ any) error {
			t.Fatalf("Write should not be called on cache hit")

			return nil
		},
		deleteFunc: func(_, _ string) error {
			return nil
		},
	}

	repo := &fakeGitRepo{}
	hist := &fakeGitRepoHist{}

	gs := gitseek.New(repo, hist, cache)

	got, err := gs.CheckLang(t.Context(), "pl", gitseek.Pair{
		EnPath:   "content/en/foo.md",
		LangPath: "content/pl/foo.md",
	})
	if err != nil {
		t.Fatalf("CheckLang returned error: %v", err)
	}

	if !reflect.DeepEqual(got, expected) {
		t.Fatalf("unexpected result:\n got: %#v\nwant: %#v", got, expected)
	}

	if len(repo.findFileLastCommitCalls) != 0 {
		t.Fatalf("git repo should not be called on cache hit")
	}

	if len(hist.findForkCommitCalls) != 0 || len(hist.findMergeCommitCalls) != 0 {
		t.Fatalf("git repo hist should not be called on cache hit")
	}

	if len(cache.readCalls) != 1 {
		t.Fatalf("expected one cache read, got %d", len(cache.readCalls))
	}

	if len(cache.writeCalls) != 0 {
		t.Fatalf("expected no cache writes, got %d", len(cache.writeCalls))
	}
}

func TestGitSeek_CheckLang_ReturnsErrorWhenCacheReadFails(t *testing.T) {
	t.Parallel()

	expectedErr := errors.New("cache read failed")

	cache := &fakeCacheStorage{
		readFunc: func(_, _ string, _ any) (bool, error) {
			return false, expectedErr
		},
		writeFunc: func(_, _ string, _ any) error {
			t.Fatalf("Write should not be called when cache read fails")

			return nil
		},
		deleteFunc: func(_, _ string) error {
			return nil
		},
	}

	gs := gitseek.New(&fakeGitRepo{}, &fakeGitRepoHist{}, cache)

	_, err := gs.CheckLang(t.Context(), "pl", gitseek.Pair{
		EnPath:   "content/en/foo.md",
		LangPath: "content/pl/foo.md",
	})

	if err == nil {
		t.Fatalf("expected error, got nil")
	}

	if !errors.Is(err, expectedErr) {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestGitSeek_CheckLang_ComputesAndWritesCacheOnMiss(t *testing.T) {
	t.Parallel()

	langLastCommit := git.CommitInfo{CommitID: "lang-last"}
	langMergeCommit := &git.CommitInfo{CommitID: "merge-lang"}
	forkCommit := &git.CommitInfo{CommitID: "fork-1"}
	enCommit1 := git.CommitInfo{CommitID: "en-1"}
	enCommit2 := git.CommitInfo{CommitID: "en-2"}
	mergeEn1 := &git.CommitInfo{CommitID: "merge-en-1"}
	mergeEn2 := &git.CommitInfo{CommitID: "merge-en-2"}

	cache := &fakeCacheStorage{
		readFunc: func(_, _ string, _ any) (bool, error) {
			return false, nil
		},
		writeFunc: func(_, _ string, _ any) error {
			return nil
		},
		deleteFunc: func(_, _ string) error {
			return nil
		},
	}

	repo := &fakeGitRepo{
		findFileLastCommitFunc: func(_ context.Context, _ string) (git.CommitInfo, error) {
			return langLastCommit, nil
		},
		findFileCommitsAfterFunc: func(_ context.Context, _ string, _ string) ([]git.CommitInfo, error) {
			return []git.CommitInfo{enCommit1, enCommit2}, nil
		},
		fileExistsFunc: func(_ string) (bool, error) {
			return true, nil
		},
	}

	hist := &fakeGitRepoHist{
		findForkCommitFunc: func(_ context.Context, _ string) (*git.CommitInfo, error) {
			return forkCommit, nil
		},
		findMergeCommitFunc: func(_ context.Context, commitID string) (*git.CommitInfo, error) {
			switch commitID {
			case "lang-last":
				return langMergeCommit, nil
			case "en-1":
				return mergeEn1, nil
			case "en-2":
				return mergeEn2, nil
			default:
				t.Fatalf("unexpected commit id in FindMergeCommit: %s", commitID)

				return nil, nil
			}
		},
	}

	gs := gitseek.New(repo, hist, cache)

	got, err := gs.CheckLang(t.Context(), "pl", gitseek.Pair{
		EnPath:   "content/en/foo.md",
		LangPath: "content/pl/foo.md",
	})
	if err != nil {
		t.Fatalf("CheckLang returned error: %v", err)
	}

	want := gitseek.FileInfo{
		LangPath:        "content/pl/foo.md",
		LangLastCommit:  langLastCommit,
		LangMergeCommit: langMergeCommit,
		LangForkCommit:  forkCommit,
		FileStatus:      gitseek.StatusEnFileUpdated,
		EnUpdates: []gitseek.EnUpdate{
			{Commit: enCommit1, MergePoint: mergeEn1},
			{Commit: enCommit2, MergePoint: mergeEn2},
		},
	}

	if !reflect.DeepEqual(got, want) {
		t.Fatalf("unexpected result:\n got: %#v\nwant: %#v", got, want)
	}

	if len(cache.writeCalls) != 1 {
		t.Fatalf("expected one cache write, got %d", len(cache.writeCalls))
	}

	writeCall := cache.writeCalls[0]

	if writeCall.Bucket != gitseek.FileInfoCacheBucket("pl") {
		t.Fatalf("unexpected bucket written: %s", writeCall.Bucket)
	}

	if writeCall.Key != "content/pl/foo.md" {
		t.Fatalf("unexpected key written: %s", writeCall.Key)
	}

	written, ok := writeCall.Data.(gitseek.FileInfo)
	if !ok {
		t.Fatalf("unexpected written data type: %T", writeCall.Data)
	}

	if !reflect.DeepEqual(written, want) {
		t.Fatalf("unexpected cache write value:\n got: %#v\nwant: %#v", written, want)
	}
}

func TestGitSeek_CheckLang_ReturnsErrorWhenCacheWriteFails(t *testing.T) {
	t.Parallel()

	expectedErr := errors.New("cache write failed")

	cache := &fakeCacheStorage{
		readFunc: func(_, _ string, _ any) (bool, error) {
			return false, nil
		},
		writeFunc: func(_, _ string, _ any) error {
			return expectedErr
		},
		deleteFunc: func(_, _ string) error {
			return nil
		},
	}

	repo := &fakeGitRepo{
		findFileLastCommitFunc: func(_ context.Context, _ string) (git.CommitInfo, error) {
			return git.CommitInfo{CommitID: "lang-last"}, nil
		},
		findFileCommitsAfterFunc: func(_ context.Context, _, _ string) ([]git.CommitInfo, error) {
			return nil, nil
		},
		fileExistsFunc: func(_ string) (bool, error) {
			return true, nil
		},
	}

	hist := &fakeGitRepoHist{
		findForkCommitFunc: func(_ context.Context, _ string) (*git.CommitInfo, error) {
			return nil, nil
		},
		findMergeCommitFunc: func(_ context.Context, _ string) (*git.CommitInfo, error) {
			return nil, nil
		},
	}

	gs := gitseek.New(repo, hist, cache)

	_, err := gs.CheckLang(t.Context(), "pl", gitseek.Pair{
		EnPath:   "content/en/foo.md",
		LangPath: "content/pl/foo.md",
	})

	if err == nil {
		t.Fatalf("expected error, got nil")
	}

	if !errors.Is(err, expectedErr) {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestGitSeek_CheckLang_UsesForkCommitAsStartPoint(t *testing.T) {
	t.Parallel()

	cache := &fakeCacheStorage{
		readFunc: func(_, _ string, _ any) (bool, error) {
			return false, nil
		},
		writeFunc: func(_, _ string, _ any) error {
			return nil
		},
		deleteFunc: func(_, _ string) error {
			return nil
		},
	}

	repo := &fakeGitRepo{
		findFileLastCommitFunc: func(_ context.Context, _ string) (git.CommitInfo, error) {
			return git.CommitInfo{CommitID: "lang-last"}, nil
		},
		findFileCommitsAfterFunc: func(_ context.Context, _ string, commitIDFrom string) ([]git.CommitInfo, error) {
			if commitIDFrom != "fork-1" {
				t.Fatalf("unexpected commitIDFrom: got %q, want %q", commitIDFrom, "fork-1")
			}

			return nil, nil
		},
		fileExistsFunc: func(_ string) (bool, error) {
			return true, nil
		},
	}

	hist := &fakeGitRepoHist{
		findForkCommitFunc: func(_ context.Context, _ string) (*git.CommitInfo, error) {
			return &git.CommitInfo{CommitID: "fork-1"}, nil
		},
		findMergeCommitFunc: func(_ context.Context, _ string) (*git.CommitInfo, error) {
			return nil, nil
		},
	}

	gs := gitseek.New(repo, hist, cache)

	_, err := gs.CheckLang(t.Context(), "pl", gitseek.Pair{
		EnPath:   "content/en/foo.md",
		LangPath: "content/pl/foo.md",
	})
	if err != nil {
		t.Fatalf("CheckLang returned error: %v", err)
	}
}

func TestGitSeek_CheckLang_UsesLangLastCommitAsStartPointWhenForkIsNil(t *testing.T) {
	t.Parallel()

	cache := &fakeCacheStorage{
		readFunc: func(_, _ string, _ any) (bool, error) {
			return false, nil
		},
		writeFunc: func(_, _ string, _ any) error {
			return nil
		},
		deleteFunc: func(_, _ string) error {
			return nil
		},
	}

	repo := &fakeGitRepo{
		findFileLastCommitFunc: func(_ context.Context, _ string) (git.CommitInfo, error) {
			return git.CommitInfo{CommitID: "lang-last"}, nil
		},
		findFileCommitsAfterFunc: func(_ context.Context, _ string, commitIDFrom string) ([]git.CommitInfo, error) {
			if commitIDFrom != "lang-last" {
				t.Fatalf("unexpected commitIDFrom: got %q, want %q", commitIDFrom, "lang-last")
			}

			return nil, nil
		},
		fileExistsFunc: func(_ string) (bool, error) {
			return true, nil
		},
	}

	hist := &fakeGitRepoHist{
		findForkCommitFunc: func(_ context.Context, _ string) (*git.CommitInfo, error) {
			return nil, nil
		},
		findMergeCommitFunc: func(_ context.Context, _ string) (*git.CommitInfo, error) {
			return nil, nil
		},
	}

	gs := gitseek.New(repo, hist, cache)

	_, err := gs.CheckLang(t.Context(), "pl", gitseek.Pair{
		EnPath:   "content/en/foo.md",
		LangPath: "content/pl/foo.md",
	})
	if err != nil {
		t.Fatalf("CheckLang returned error: %v", err)
	}
}

func TestGitSeek_CheckLang_SetsStatusEnFileDoesNotExist(t *testing.T) {
	t.Parallel()

	got := runCheckLangForStatus(t, false, nil)

	if got.FileStatus != gitseek.StatusEnFileDoesNotExist {
		t.Fatalf("unexpected status: got %q, want %q", got.FileStatus, gitseek.StatusEnFileDoesNotExist)
	}
}

func TestGitSeek_CheckLang_SetsStatusEnFileNoLongerExists(t *testing.T) {
	t.Parallel()

	got := runCheckLangForStatus(t, false, []git.CommitInfo{{CommitID: "en-1"}})

	if got.FileStatus != gitseek.StatusEnFileNoLongerExists {
		t.Fatalf("unexpected status: got %q, want %q", got.FileStatus, gitseek.StatusEnFileNoLongerExists)
	}
}

func TestGitSeek_CheckLang_SetsStatusEnFileUpdated(t *testing.T) {
	t.Parallel()

	got := runCheckLangForStatus(t, true, []git.CommitInfo{{CommitID: "en-1"}})

	if got.FileStatus != gitseek.StatusEnFileUpdated {
		t.Fatalf("unexpected status: got %q, want %q", got.FileStatus, gitseek.StatusEnFileUpdated)
	}
}

func TestGitSeek_CheckLang_LeavesStatusEmptyWhenEnFileExistsAndHasNoUpdates(t *testing.T) {
	t.Parallel()

	got := runCheckLangForStatus(t, true, nil)

	if got.FileStatus != "" {
		t.Fatalf("unexpected status: got %q, want empty", got.FileStatus)
	}
}

func TestGitSeek_CheckLang_ReturnsErrorWhenFindFileLastCommitFails(t *testing.T) {
	t.Parallel()

	expectedErr := errors.New("find last commit failed")

	cache := &fakeCacheStorage{
		readFunc: func(_, _ string, _ any) (bool, error) {
			return false, nil
		},
		writeFunc: func(_, _ string, _ any) error {
			return nil
		},
		deleteFunc: func(_, _ string) error {
			return nil
		},
	}

	repo := &fakeGitRepo{
		findFileLastCommitFunc: func(_ context.Context, _ string) (git.CommitInfo, error) {
			return git.CommitInfo{}, expectedErr
		},
	}
	hist := &fakeGitRepoHist{}

	gs := gitseek.New(repo, hist, cache)

	_, err := gs.CheckLang(t.Context(), "pl", gitseek.Pair{
		EnPath:   "content/en/foo.md",
		LangPath: "content/pl/foo.md",
	})

	if err == nil {
		t.Fatalf("expected error, got nil")
	}

	if !errors.Is(err, expectedErr) {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestGitSeek_CheckLang_ReturnsErrorWhenFindMergeCommitForLangFails(t *testing.T) {
	t.Parallel()

	expectedErr := errors.New("find merge failed")

	cache := &fakeCacheStorage{
		readFunc: func(_, _ string, _ any) (bool, error) {
			return false, nil
		},
		writeFunc: func(_, _ string, _ any) error {
			return nil
		},
		deleteFunc: func(_, _ string) error {
			return nil
		},
	}

	repo := &fakeGitRepo{
		findFileLastCommitFunc: func(_ context.Context, _ string) (git.CommitInfo, error) {
			return git.CommitInfo{CommitID: "lang-last"}, nil
		},
	}
	hist := &fakeGitRepoHist{
		findMergeCommitFunc: func(_ context.Context, _ string) (*git.CommitInfo, error) {
			return nil, expectedErr
		},
	}

	gs := gitseek.New(repo, hist, cache)

	_, err := gs.CheckLang(t.Context(), "pl", gitseek.Pair{
		EnPath:   "content/en/foo.md",
		LangPath: "content/pl/foo.md",
	})

	if err == nil {
		t.Fatalf("expected error, got nil")
	}

	if !errors.Is(err, expectedErr) {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestGitSeek_CheckLang_ReturnsErrorWhenFindForkCommitFails(t *testing.T) {
	t.Parallel()

	expectedErr := errors.New("find fork failed")

	cache := &fakeCacheStorage{
		readFunc: func(_, _ string, _ any) (bool, error) {
			return false, nil
		},
		writeFunc: func(_, _ string, _ any) error {
			return nil
		},
		deleteFunc: func(_, _ string) error {
			return nil
		},
	}

	repo := &fakeGitRepo{
		findFileLastCommitFunc: func(_ context.Context, _ string) (git.CommitInfo, error) {
			return git.CommitInfo{CommitID: "lang-last"}, nil
		},
	}
	hist := &fakeGitRepoHist{
		findMergeCommitFunc: func(_ context.Context, _ string) (*git.CommitInfo, error) {
			return nil, nil
		},
		findForkCommitFunc: func(_ context.Context, _ string) (*git.CommitInfo, error) {
			return nil, expectedErr
		},
	}

	gs := gitseek.New(repo, hist, cache)

	_, err := gs.CheckLang(t.Context(), "pl", gitseek.Pair{
		EnPath:   "content/en/foo.md",
		LangPath: "content/pl/foo.md",
	})

	if err == nil {
		t.Fatalf("expected error, got nil")
	}

	if !errors.Is(err, expectedErr) {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestGitSeek_CheckLang_ReturnsErrorWhenFindFileCommitsAfterFails(t *testing.T) {
	t.Parallel()

	expectedErr := errors.New("find commits after failed")

	cache := &fakeCacheStorage{
		readFunc: func(_, _ string, _ any) (bool, error) {
			return false, nil
		},
		writeFunc: func(_, _ string, _ any) error {
			return nil
		},
		deleteFunc: func(_, _ string) error {
			return nil
		},
	}

	repo := &fakeGitRepo{
		findFileLastCommitFunc: func(_ context.Context, _ string) (git.CommitInfo, error) {
			return git.CommitInfo{CommitID: "lang-last"}, nil
		},
		findFileCommitsAfterFunc: func(_ context.Context, _, _ string) ([]git.CommitInfo, error) {
			return nil, expectedErr
		},
	}
	hist := &fakeGitRepoHist{
		findMergeCommitFunc: func(_ context.Context, _ string) (*git.CommitInfo, error) {
			return nil, nil
		},
		findForkCommitFunc: func(_ context.Context, _ string) (*git.CommitInfo, error) {
			return nil, nil
		},
	}

	gs := gitseek.New(repo, hist, cache)

	_, err := gs.CheckLang(t.Context(), "pl", gitseek.Pair{
		EnPath:   "content/en/foo.md",
		LangPath: "content/pl/foo.md",
	})

	if err == nil {
		t.Fatalf("expected error, got nil")
	}

	if !errors.Is(err, expectedErr) {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestGitSeek_CheckLang_ReturnsErrorWhenFileExistsFails(t *testing.T) {
	t.Parallel()

	expectedErr := errors.New("file exists failed")

	cache := &fakeCacheStorage{
		readFunc: func(_, _ string, _ any) (bool, error) {
			return false, nil
		},
		writeFunc: func(_, _ string, _ any) error {
			return nil
		},
		deleteFunc: func(_, _ string) error {
			return nil
		},
	}

	repo := &fakeGitRepo{
		findFileLastCommitFunc: func(_ context.Context, _ string) (git.CommitInfo, error) {
			return git.CommitInfo{CommitID: "lang-last"}, nil
		},
		findFileCommitsAfterFunc: func(_ context.Context, _, _ string) ([]git.CommitInfo, error) {
			return nil, nil
		},
		fileExistsFunc: func(_ string) (bool, error) {
			return false, expectedErr
		},
	}
	hist := &fakeGitRepoHist{
		findMergeCommitFunc: func(_ context.Context, _ string) (*git.CommitInfo, error) {
			return nil, nil
		},
		findForkCommitFunc: func(_ context.Context, _ string) (*git.CommitInfo, error) {
			return nil, nil
		},
	}

	gs := gitseek.New(repo, hist, cache)

	_, err := gs.CheckLang(t.Context(), "pl", gitseek.Pair{
		EnPath:   "content/en/foo.md",
		LangPath: "content/pl/foo.md",
	})

	if err == nil {
		t.Fatalf("expected error, got nil")
	}

	if !errors.Is(err, expectedErr) {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestGitSeek_CheckLang_ReturnsErrorWhenFindingENUpdateMergeCommitFails(t *testing.T) {
	t.Parallel()

	expectedErr := errors.New("find EN merge failed")

	cache := &fakeCacheStorage{
		readFunc: func(_, _ string, _ any) (bool, error) {
			return false, nil
		},
		writeFunc: func(_, _ string, _ any) error {
			return nil
		},
		deleteFunc: func(_, _ string) error {
			return nil
		},
	}

	repo := &fakeGitRepo{
		findFileLastCommitFunc: func(_ context.Context, _ string) (git.CommitInfo, error) {
			return git.CommitInfo{CommitID: "lang-last"}, nil
		},
		findFileCommitsAfterFunc: func(_ context.Context, _, _ string) ([]git.CommitInfo, error) {
			return []git.CommitInfo{{CommitID: "en-1"}}, nil
		},
		fileExistsFunc: func(_ string) (bool, error) {
			return true, nil
		},
	}
	hist := &fakeGitRepoHist{
		findMergeCommitFunc: func(_ context.Context, commitID string) (*git.CommitInfo, error) {
			if commitID == "lang-last" {
				return nil, nil
			}
			if commitID == "en-1" {
				return nil, expectedErr
			}
			t.Fatalf("unexpected commit id: %s", commitID)

			return nil, nil
		},
		findForkCommitFunc: func(_ context.Context, _ string) (*git.CommitInfo, error) {
			return nil, nil
		},
	}

	gs := gitseek.New(repo, hist, cache)

	_, err := gs.CheckLang(t.Context(), "pl", gitseek.Pair{
		EnPath:   "content/en/foo.md",
		LangPath: "content/pl/foo.md",
	})

	if err == nil {
		t.Fatalf("expected error, got nil")
	}

	if !errors.Is(err, expectedErr) {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestGitSeek_InvalidateFile_DeletesCacheEntry(t *testing.T) {
	t.Parallel()

	cache := &fakeCacheStorage{
		readFunc: func(_, _ string, _ any) (bool, error) {
			return false, nil
		},
		writeFunc: func(_, _ string, _ any) error {
			return nil
		},
		deleteFunc: func(_, _ string) error {
			return nil
		},
	}

	gs := gitseek.New(&fakeGitRepo{}, &fakeGitRepoHist{}, cache)

	err := gs.InvalidateFile("pl", "content/pl/foo.md")
	if err != nil {
		t.Fatalf("InvalidateFile returned error: %v", err)
	}

	if len(cache.deleteCalls) != 1 {
		t.Fatalf("expected one delete call, got %d", len(cache.deleteCalls))
	}

	got := cache.deleteCalls[0]
	want := cacheKey{
		Bucket: gitseek.FileInfoCacheBucket("pl"),
		Key:    "content/pl/foo.md",
	}

	if !reflect.DeepEqual(got, want) {
		t.Fatalf("unexpected delete call:\n got: %#v\nwant: %#v", got, want)
	}
}

func TestGitSeek_InvalidateFile_ReturnsErrorWhenDeleteFails(t *testing.T) {
	t.Parallel()

	expectedErr := errors.New("delete failed")

	cache := &fakeCacheStorage{
		readFunc: func(_, _ string, _ any) (bool, error) {
			return false, nil
		},
		writeFunc: func(_, _ string, _ any) error {
			return nil
		},
		deleteFunc: func(_, _ string) error {
			return expectedErr
		},
	}

	gs := gitseek.New(&fakeGitRepo{}, &fakeGitRepoHist{}, cache)

	err := gs.InvalidateFile("pl", "content/pl/foo.md")

	if err == nil {
		t.Fatalf("expected error, got nil")
	}

	if !errors.Is(err, expectedErr) {
		t.Fatalf("unexpected error: %v", err)
	}
}

func runCheckLangForStatus(t *testing.T, fileExists bool, enCommitsAfter []git.CommitInfo) gitseek.FileInfo {
	t.Helper()

	cache := &fakeCacheStorage{
		readFunc: func(_, _ string, _ any) (bool, error) {
			return false, nil
		},
		writeFunc: func(_, _ string, _ any) error {
			return nil
		},
		deleteFunc: func(_, _ string) error {
			return nil
		},
	}

	repo := &fakeGitRepo{
		findFileLastCommitFunc: func(_ context.Context, _ string) (git.CommitInfo, error) {
			return git.CommitInfo{CommitID: "lang-last"}, nil
		},
		findFileCommitsAfterFunc: func(_ context.Context, _, _ string) ([]git.CommitInfo, error) {
			return enCommitsAfter, nil
		},
		fileExistsFunc: func(_ string) (bool, error) {
			return fileExists, nil
		},
	}

	hist := &fakeGitRepoHist{
		findMergeCommitFunc: func(_ context.Context, _ string) (*git.CommitInfo, error) {
			return nil, nil
		},
		findForkCommitFunc: func(_ context.Context, _ string) (*git.CommitInfo, error) {
			return nil, nil
		},
	}

	gs := gitseek.New(repo, hist, cache)

	got, err := gs.CheckLang(t.Context(), "pl", gitseek.Pair{
		EnPath:   "content/en/foo.md",
		LangPath: "content/pl/foo.md",
	})
	if err != nil {
		t.Fatalf("CheckLang returned error: %v", err)
	}

	return got
}
