// Package pullreq provides information about pull requests
package pullreq

//go:generate mockgen -typed -source=pullreq.go -destination=./internal/mocks/mocks.go -package=mocks

import (
	"context"
	"errors"
	"fmt"
	"log"
	"sort"
	"strconv"

	"github.com/dkarczmarski/go-kweb-lang/filepairs"
	"github.com/dkarczmarski/go-kweb-lang/github"
	"github.com/dkarczmarski/go-kweb-lang/proxycache"
	"github.com/dkarczmarski/go-kweb-lang/pullreq/internal/cachetypes"
)

type GitHub interface {
	PRSearch(ctx context.Context, filter github.PRSearchFilter, page github.PageRequest) (*github.PRSearchResult, error)
	GetPRCommits(ctx context.Context, prNumber int) ([]string, error)
	GetCommitFiles(ctx context.Context, commitID string) (*github.CommitFiles, error)
}

type CacheStorage interface {
	Read(bucket, key string, buff any) (bool, error)
	Write(bucket, key string, data any) error
	Delete(bucket, key string) error
}

type FilePRIndexData map[string][]int

type FilePRIndex struct {
	gitHub       GitHub
	cacheStorage CacheStorage
	filePaths    *filepairs.FilePaths
	perPage      int
}

const (
	bucketPRCommits    = "pr-pr-commits"
	bucketCommitFiles  = "pr-commit-files"
	bucketFilePRsIndex = "pr-fileprs-index"
	maxPages           = 20
)

var (
	ErrPaginationDidNotAdvance = errors.New("pagination did not advance")
	ErrOpenPRPageLimitExceeded = errors.New("open pull request pagination exceeded page limit")
	ErrLangIndexNotFound       = errors.New("language pull request index not found")
)

func NewFilePRIndex(gitHub GitHub, cacheStore CacheStorage, perPage int) *FilePRIndex {
	return &FilePRIndex{
		gitHub:       gitHub,
		cacheStorage: cacheStore,
		filePaths:    filepairs.New(),
		perPage:      perPage,
	}
}

// PRCommitsCacheBucket returns the cache bucket used for PR commit lists
// for the given language.
func PRCommitsCacheBucket(langCode string) string {
	return fmt.Sprintf("lang/%s/%s", langCode, bucketPRCommits)
}

// PRCommitsCacheKey returns the cache key used for a PR commit list.
func PRCommitsCacheKey(prNumber int) string {
	return strconv.Itoa(prNumber)
}

// CommitFilesCacheBucket returns the cache bucket used for commit file lists
// for the given language.
func CommitFilesCacheBucket(langCode string) string {
	return fmt.Sprintf("lang/%s/%s", langCode, bucketCommitFiles)
}

// CommitFilesCacheKey returns the cache key used for a commit file list.
func CommitFilesCacheKey(commitID string) string {
	return commitID
}

// FilePRsIndexCacheBucket returns the cache bucket used for the file-to-PR index
// for the given language.
func FilePRsIndexCacheBucket(langCode string) string {
	return fmt.Sprintf("lang/%s/%s", langCode, bucketFilePRsIndex)
}

// FilePRsIndexCacheKey returns the cache key used for the file-to-PR index
// for the given language.
func FilePRsIndexCacheKey(langCode string) string {
	return langCode
}

const firstPage = 1

func (p *FilePRIndex) fetchOpenPRsForLang(ctx context.Context, langCode string) ([]github.PRItem, error) {
	var (
		pullRequests []github.PRItem
		maxUpdatedAt string
	)

	for page := range maxPages {
		log.Printf(
			"[pullreq][%s] fetching open PRs page %d/%d (updatedFrom=%q)",
			langCode, page+1, maxPages, maxUpdatedAt,
		)

		result, err := p.gitHub.PRSearch(
			ctx,
			github.PRSearchFilter{
				LangCode:    langCode,
				UpdatedFrom: maxUpdatedAt,
				OnlyOpen:    true,
			},
			github.PageRequest{
				Sort:    "updated",
				Order:   "asc",
				Page:    firstPage,
				PerPage: p.perPage,
			},
		)
		if err != nil {
			return nil, fmt.Errorf("search open pull requests for %s: %w", langCode, err)
		}

		if len(result.Items) == 0 {
			log.Printf("[pullreq][%s] no more open PRs to fetch", langCode)

			return pullRequests, nil
		}

		pullRequests = append(pullRequests, result.Items...)

		nextUpdatedAt := result.Items[len(result.Items)-1].UpdatedAt
		if nextUpdatedAt == maxUpdatedAt {
			return nil, fmt.Errorf(
				"%w: lang=%s updatedAt=%s",
				ErrPaginationDidNotAdvance,
				langCode,
				maxUpdatedAt,
			)
		}

		maxUpdatedAt = nextUpdatedAt
	}

	return nil, fmt.Errorf(
		"%w: lang=%s limit=%d",
		ErrOpenPRPageLimitExceeded,
		langCode,
		maxPages,
	)
}

func (p *FilePRIndex) fetchPRCommits(
	ctx context.Context,
	langCode string,
	pullRequest github.PRItem,
) ([]string, error) {
	commits, err := proxycache.Get(
		ctx,
		p.cacheStorage,
		PRCommitsCacheBucket(langCode),
		PRCommitsCacheKey(pullRequest.Number),
		func(cachedPRCommits cachetypes.PRCommits) bool {
			isStale := cachedPRCommits.UpdatedAt != pullRequest.UpdatedAt

			if isStale {
				log.Printf(
					"[pullreq][%s][pr:%d] commit cache stale: cached=%s current=%s",
					langCode,
					pullRequest.Number,
					cachedPRCommits.UpdatedAt,
					pullRequest.UpdatedAt,
				)
			}

			return isStale
		},
		func(ctx context.Context) (cachetypes.PRCommits, error) {
			log.Printf("[pullreq][%s][pr:%d] fetching commit IDs", langCode, pullRequest.Number)

			commitIDs, err := p.gitHub.GetPRCommits(ctx, pullRequest.Number)
			if err != nil {
				return cachetypes.PRCommits{}, fmt.Errorf(
					"fetch commit IDs for PR #%d in %s: %w",
					pullRequest.Number,
					langCode,
					err,
				)
			}

			return cachetypes.PRCommits{
				UpdatedAt: pullRequest.UpdatedAt,
				CommitIDs: commitIDs,
			}, nil
		},
	)
	if err != nil {
		return nil, fmt.Errorf(
			"load commit IDs for PR #%d in %s: %w",
			pullRequest.Number,
			langCode,
			err,
		)
	}

	return commits.CommitIDs, nil
}

func (p *FilePRIndex) fetchCommitFiles(
	ctx context.Context,
	langCode string,
	commitID string,
) (*github.CommitFiles, error) {
	commitFiles, err := proxycache.Get(
		ctx,
		p.cacheStorage,
		CommitFilesCacheBucket(langCode),
		CommitFilesCacheKey(commitID),
		nil,
		func(ctx context.Context) (*github.CommitFiles, error) {
			log.Printf("[pullreq][%s][commit:%s] fetching files", langCode, commitID)

			files, err := p.gitHub.GetCommitFiles(ctx, commitID)
			if err != nil {
				return nil, fmt.Errorf("fetch files for commit %s in %s: %w", commitID, langCode, err)
			}

			return files, nil
		},
	)
	if err != nil {
		return nil, fmt.Errorf("load files for commit %s in %s: %w", commitID, langCode, err)
	}

	return commitFiles, nil
}

func (p *FilePRIndex) fetchFilesForCommitList(
	ctx context.Context,
	langCode string,
	pullRequest github.PRItem,
	pullRequestIndex int,
	pullRequestsCount int,
	commitIDs []string,
) ([]string, error) {
	files := make([]string, 0)
	commitCount := len(commitIDs)

	for commitIndex, commitID := range commitIDs {
		log.Printf(
			"[pullreq][%s][%d/%d][pr:%d][%d/%d] loading files for commit %s",
			langCode,
			pullRequestIndex+1,
			pullRequestsCount,
			pullRequest.Number,
			commitIndex+1,
			commitCount,
			commitID,
		)

		commitFiles, err := p.fetchCommitFiles(ctx, langCode, commitID)
		if err != nil {
			return nil, fmt.Errorf(
				"load files for commit %s in PR #%d for %s: %w",
				commitID,
				pullRequest.Number,
				langCode,
				err,
			)
		}

		log.Printf(
			"[pullreq][%s][%d/%d][pr:%d][%d/%d] commit %s contains %d files",
			langCode,
			pullRequestIndex+1,
			pullRequestsCount,
			pullRequest.Number,
			commitIndex+1,
			commitCount,
			commitID,
			len(commitFiles.Files),
		)

		files = append(files, commitFiles.Files...)
	}

	return files, nil
}

func (p *FilePRIndex) fetchFilesForPR(
	ctx context.Context,
	langCode string,
	pullRequest github.PRItem,
	pullRequestIndex int,
	pullRequestsCount int,
) ([]string, error) {
	log.Printf(
		"[pullreq][%s][%d/%d][pr:%d] loading PR details (updatedAt=%s)",
		langCode,
		pullRequestIndex+1,
		pullRequestsCount,
		pullRequest.Number,
		pullRequest.UpdatedAt,
	)

	commitIDs, err := p.fetchPRCommits(ctx, langCode, pullRequest)
	if err != nil {
		return nil, fmt.Errorf(
			"load commit IDs for PR #%d in %s: %w",
			pullRequest.Number,
			langCode,
			err,
		)
	}

	log.Printf(
		"[pullreq][%s][%d/%d][pr:%d] found %d commits",
		langCode,
		pullRequestIndex+1,
		pullRequestsCount,
		pullRequest.Number,
		len(commitIDs),
	)

	files, err := p.fetchFilesForCommitList(
		ctx,
		langCode,
		pullRequest,
		pullRequestIndex,
		pullRequestsCount,
		commitIDs,
	)
	if err != nil {
		return nil, fmt.Errorf("load files for PR #%d in %s: %w", pullRequest.Number, langCode, err)
	}

	return files, nil
}

func (p *FilePRIndex) fetchPRFiles(
	ctx context.Context,
	langCode string,
	pullRequests []github.PRItem,
) (map[int][]string, error) {
	prsFiles := make(map[int][]string, len(pullRequests))
	pullRequestsCount := len(pullRequests)

	for pullRequestIndex, pullRequest := range pullRequests {
		files, err := p.fetchFilesForPR(
			ctx,
			langCode,
			pullRequest,
			pullRequestIndex,
			pullRequestsCount,
		)
		if err != nil {
			return nil, fmt.Errorf(
				"load files for PR #%d in %s: %w",
				pullRequest.Number,
				langCode,
				err,
			)
		}

		prsFiles[pullRequest.Number] = files
	}

	return prsFiles, nil
}

func (p *FilePRIndex) filterFilesForLang(files []string, langCode string) []string {
	filtered := make([]string, 0, len(files))

	for _, file := range files {
		pathInfo, err := p.filePaths.CheckPath(file)
		if err != nil {
			log.Printf("[pullreq][%s] skipping file %q: path check failed: %v", langCode, file, err)

			continue
		}

		if pathInfo.LangCode != langCode {
			continue
		}

		filtered = append(filtered, file)
	}

	return filtered
}

func (p *FilePRIndex) buildFilePRIndex(prsFiles map[int][]string, langCode string) FilePRIndexData {
	filePRs := make(FilePRIndexData, len(prsFiles))
	seen := make(map[string]map[int]struct{}, len(prsFiles))

	for prNumber, files := range prsFiles {
		langFiles := p.filterFilesForLang(files, langCode)

		for _, file := range langFiles {
			if seen[file] == nil {
				seen[file] = make(map[int]struct{})
			}

			if _, exists := seen[file][prNumber]; exists {
				continue
			}

			seen[file][prNumber] = struct{}{}

			filePRs[file] = append(filePRs[file], prNumber)
		}
	}

	for file, prs := range filePRs {
		sort.Sort(sort.Reverse(sort.IntSlice(prs)))
		filePRs[file] = prs
	}

	return filePRs
}

func (p *FilePRIndex) writeLangIndex(langCode string, filePRs FilePRIndexData) error {
	if err := p.cacheStorage.Write(
		FilePRsIndexCacheBucket(langCode),
		FilePRsIndexCacheKey(langCode),
		filePRs,
	); err != nil {
		return fmt.Errorf("write file PR index for %s: %w", langCode, err)
	}

	return nil
}

// RefreshIndex fetches current PR data and rebuilds the file-to-PR index
// for the given language.
func (p *FilePRIndex) RefreshIndex(ctx context.Context, langCode string) error {
	log.Printf("[pullreq][%s] refreshing file PR index", langCode)

	pullRequests, err := p.fetchOpenPRsForLang(ctx, langCode)
	if err != nil {
		return fmt.Errorf("fetch open pull requests for %s: %w", langCode, err)
	}

	log.Printf("[pullreq][%s] fetched %d open pull requests", langCode, len(pullRequests))

	prsFiles, err := p.fetchPRFiles(ctx, langCode, pullRequests)
	if err != nil {
		return fmt.Errorf("fetch pull request files for %s: %w", langCode, err)
	}

	filePRs := p.buildFilePRIndex(prsFiles, langCode)

	log.Printf("[pullreq][%s] built file PR index with %d files", langCode, len(filePRs))

	if err := p.writeLangIndex(langCode, filePRs); err != nil {
		return fmt.Errorf("store file PR index for %s: %w", langCode, err)
	}

	log.Printf("[pullreq][%s] file PR index refreshed", langCode)

	return nil
}

// LangIndex returns a map from file names to a list of pull request indices for
// the given langCode.
func (p *FilePRIndex) LangIndex(langCode string) (FilePRIndexData, error) {
	bucket := FilePRsIndexCacheBucket(langCode)
	key := FilePRsIndexCacheKey(langCode)

	var filePRs FilePRIndexData

	exists, err := p.cacheStorage.Read(bucket, key, &filePRs)
	if err != nil {
		return nil, fmt.Errorf("read file PR index for %s: %w", langCode, err)
	}

	if !exists {
		return nil, fmt.Errorf("%w: %s", ErrLangIndexNotFound, langCode)
	}

	return filePRs, nil
}
