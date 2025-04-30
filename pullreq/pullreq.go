// Package pullreq provides information about pull requests
package pullreq

//go:generate mockgen -typed -source=pullreq.go -destination=../mocks/mock_pullreq.go -package=mocks

import (
	"context"
	"fmt"
	"log"
	"slices"
	"sort"

	"go-kweb-lang/pullreq/internal"

	"go-kweb-lang/github"
	"go-kweb-lang/proxycache"
)

type FilePRFinderConfig struct {
	Storage FilePRFinderStorage
	PerPage int
}

type FilePRFinder struct {
	gitHub   github.GitHub
	cacheDir string
	storage  FilePRFinderStorage

	perPage int
}

type FilePRFinderStorage interface {
	LangIndex(langCode string) (map[string][]int, error)
	StoreLangIndex(langCode string, filePRs map[string][]int) error
}

func NewFilePRFinder(gitHub github.GitHub, cacheDir string, opts ...func(config *FilePRFinderConfig)) *FilePRFinder {
	config := FilePRFinderConfig{
		Storage: &FilePRFinderFileStorage{
			cacheDir: cacheDir,
		},
		PerPage: 100,
	}

	for _, opt := range opts {
		opt(&config)
	}

	return &FilePRFinder{
		gitHub:   gitHub,
		cacheDir: cacheDir,
		storage:  config.Storage,
		perPage:  config.PerPage,
	}
}

func (p *FilePRFinder) fetchLangOpenedPRs(ctx context.Context, langCode string) ([]github.PRItem, error) {
	var prs []github.PRItem

	var maxUpdatedAt string
	for safetyCounter := 20; safetyCounter >= 0; safetyCounter-- {
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
				PerPage: p.perPage,
			},
		)
		if err != nil {
			return nil, fmt.Errorf("error while searching for pull requests: %w", err)
		}
		if len(result.Items) == 0 {
			break
		}

		prs = append(prs, result.Items...)

		maxUpdatedAt = result.Items[len(result.Items)-1].UpdatedAt

		if safetyCounter == 0 {
			log.Fatal("too many requests. probably something is wrong")
		}
	}

	return prs, nil
}

func (p *FilePRFinder) fetchPRCommits(ctx context.Context, pr github.PRItem) ([]string, error) {
	key := fmt.Sprintf("%v", pr.Number)
	commits, err := proxycache.Get(
		ctx,
		p.cacheDir,
		internal.CategoryPrCommits,
		key,
		func(cachedPrCommits internal.PRCommits) bool {
			isInvalid := cachedPrCommits.UpdatedAt != pr.UpdatedAt

			if isInvalid {
				log.Printf("PR #%v hit cache but is out of date: %v", pr.Number, cachedPrCommits.UpdatedAt)
			}

			return isInvalid
		},
		func(ctx context.Context) (internal.PRCommits, error) {
			log.Printf("fetching commit list for PR #%v", pr.Number)

			commitIds, err := p.gitHub.GetPRCommits(ctx, pr.Number)
			if err != nil {
				return internal.PRCommits{}, err
			}

			return internal.PRCommits{
				UpdatedAt: pr.UpdatedAt,
				CommitIds: commitIds,
			}, nil
		},
	)
	if err != nil {
		return nil, err
	}

	return commits.CommitIds, nil
}

func (p *FilePRFinder) fetchCommitFiles(ctx context.Context, commitID string) (*github.CommitFiles, error) {
	return proxycache.Get(
		ctx,
		p.cacheDir,
		internal.CategoryCommitFiles,
		commitID,
		nil,
		func(ctx context.Context) (*github.CommitFiles, error) {
			return p.gitHub.GetCommitFiles(ctx, commitID)
		},
	)
}

func (p *FilePRFinder) convertToFilePRs(prsFiles map[int][]string) map[string][]int {
	filePRs := make(map[string][]int)

	for pr, files := range prsFiles {
		for _, file := range files {
			if !slices.Contains(filePRs[file], pr) {
				filePRs[file] = append(filePRs[file], pr)
			}
		}
	}

	for file, prs := range filePRs {
		sort.Sort(sort.Reverse(sort.IntSlice(prs)))
		filePRs[file] = prs
	}

	return filePRs
}

// Update updates the file-to-PR index for the given langCode.
func (p *FilePRFinder) Update(ctx context.Context, langCode string) error {
	log.Printf("[%v] updating the index of PR files", langCode)

	prs, err := p.fetchLangOpenedPRs(ctx, langCode)
	if err != nil {
		return fmt.Errorf("error while getting pull requests: %w", err)
	}

	prsLen := len(prs)
	log.Printf("[%v] fetched a PR list of size %v", langCode, prsLen)

	prsFiles := make(map[int][]string)
	for prIndex, pr := range prs {
		log.Printf("[%v][%v/%v] fetched info for PR #%v. updated at: %v",
			langCode, prIndex, prsLen, pr.Number, pr.UpdatedAt)

		log.Printf("[%v][%v/%v] getting commit ids for PR #%v",
			langCode, prIndex, prsLen, pr.Number)

		commitIds, err := p.fetchPRCommits(ctx, pr)
		if err != nil {
			return fmt.Errorf("error while getting commits for pr %v: %w", pr.Number, err)
		}

		commitIdsLen := len(commitIds)

		log.Printf("[%v][%v/%v] for PR #%v got commit list of size %v",
			langCode, prIndex, prsLen, pr.Number, commitIdsLen)

		for commitIndex, commitID := range commitIds {
			log.Printf("[%v][%v/%v][%v/%v] getting file list for commit: %v",
				langCode, prIndex, prsLen, commitIndex, commitIdsLen, commitID)

			commitFiles, err := p.fetchCommitFiles(ctx, commitID)
			if err != nil {
				return fmt.Errorf("error while getting files for commit id %v: %w", commitID, err)
			}

			log.Printf("[%v][%v/%v][%v/%v] for commit %v got file list of size %v",
				langCode, prIndex, prsLen, commitIndex, commitIdsLen, commitID, len(commitFiles.Files))

			prsFiles[pr.Number] = append(prsFiles[pr.Number], commitFiles.Files...)
		}
	}

	filePRs := p.convertToFilePRs(prsFiles)

	if err := p.storage.StoreLangIndex(langCode, filePRs); err != nil {
		return fmt.Errorf("error while storing %v index: %w", langCode, err)
	}

	return nil
}

// LangIndex returns a map from file names to a list of pull request indices for the given langCode.
func (p *FilePRFinder) LangIndex(langCode string) (map[string][]int, error) {
	return p.storage.LangIndex(langCode)
}

type FilePRFinderFileStorage struct {
	cacheDir string
}

func (sto *FilePRFinderFileStorage) LangIndex(langCode string) (map[string][]int, error) {
	return proxycache.Get(
		context.Background(),
		sto.cacheDir,
		internal.CategoryFilePrsIndex,
		langCode,
		nil,
		func(ctx context.Context) (map[string][]int, error) {
			return nil, nil
		},
	)
}

func (sto *FilePRFinderFileStorage) StoreLangIndex(langCode string, filePRs map[string][]int) error {
	return proxycache.Put(
		sto.cacheDir,
		internal.CategoryFilePrsIndex,
		langCode,
		filePRs,
	)
}
