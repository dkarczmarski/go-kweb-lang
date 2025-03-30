// Package pullreq provides information about pull requests
package pullreq

import (
	"context"
	"fmt"
	"go-kweb-lang/github"
	"go-kweb-lang/proxycache"
	"log"
	"sort"
)

const (
	categoryPrCommits   = "pr-pr-commits"
	categoryCommitFiles = "pr-commit-files"
	categoryFilePrs     = "pr-fileprs"
)

type PullRequests struct {
	GitHub   github.GitHub
	CacheDir string
	PerPage  int
}

func (p *PullRequests) fetchLangOpenedPRs(langCode string) ([]github.PRItem, error) {
	var prs []github.PRItem

	var maxUpdatedAt string
	for safetyCounter := 20; safetyCounter >= 0; safetyCounter-- {
		result, err := p.GitHub.PRSearch(
			github.PRSearchFilter{
				LangCode:    langCode,
				UpdatedFrom: maxUpdatedAt,
				OnlyOpen:    true,
			},
			github.PageRequest{
				Sort:    "updated",
				Order:   "asc",
				PerPage: p.PerPage,
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

type prCommits struct {
	UpdatedAt string
	CommitIds []string
}

func (p *PullRequests) fetchPRCommits(ctx context.Context, pr github.PRItem) ([]string, error) {
	key := fmt.Sprintf("%v", pr.Number)
	commits, err := proxycache.Get(
		ctx,
		p.CacheDir,
		categoryPrCommits,
		key,
		func(cachedPrCommits prCommits) bool {
			isInvalid := cachedPrCommits.UpdatedAt != pr.UpdatedAt

			if isInvalid {
				log.Printf("PR #%v hit cache but is out of date: %v", pr.Number, cachedPrCommits.UpdatedAt)
			}

			return isInvalid
		},
		func(ctx context.Context) (prCommits, error) {
			log.Printf("fetching commit list for PR #%v", pr.Number)

			commitIds, err := p.GitHub.GetPRCommits(pr.Number)
			if err != nil {
				return prCommits{}, err
			}

			return prCommits{pr.UpdatedAt, commitIds}, nil
		},
	)

	if err != nil {
		return nil, err
	}

	return commits.CommitIds, nil
}

func (p *PullRequests) fetchCommitFiles(commitID string) (*github.CommitFiles, error) {
	return proxycache.Get(
		context.Background(), // todo:
		p.CacheDir,
		categoryCommitFiles,
		commitID,
		nil,
		func(ctx context.Context) (*github.CommitFiles, error) {
			return p.GitHub.GetCommitFiles(commitID)
		},
	)
}

func (p *PullRequests) convertToFilePRs(prsFiles map[int][]string) map[string][]int {
	filePRs := make(map[string][]int)

	for pr, files := range prsFiles {
		for _, file := range files {
			filePRs[file] = append(filePRs[file], pr)
		}
	}

	for file, prs := range filePRs {
		sort.Sort(sort.Reverse(sort.IntSlice(prs)))
		filePRs[file] = prs
	}

	return filePRs
}

func (p *PullRequests) Update(ctx context.Context, langCode string) error {
	log.Printf("[%v] updating the index of PR files", langCode)

	prs, err := p.fetchLangOpenedPRs(langCode)
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

			commitFiles, err := p.fetchCommitFiles(commitID)
			if err != nil {
				return fmt.Errorf("error while getting files for commit id %v: %w", commitID, err)
			}

			log.Printf("[%v][%v/%v][%v/%v] for commit %v got file list of size %v",
				langCode, prIndex, prsLen, commitIndex, commitIdsLen, commitID, len(commitFiles.Files))

			prsFiles[pr.Number] = append(prsFiles[pr.Number], commitFiles.Files...)
		}
	}

	filePRs := p.convertToFilePRs(prsFiles)

	if err := p.storeAll(filePRs); err != nil {
		return fmt.Errorf("error while storing the file PRs index: %w", err)
	}

	return nil
}

func (p *PullRequests) storeAll(filePRs map[string][]int) error {
	for path, prs := range filePRs {
		if err := p.store(path, prs); err != nil {
			return fmt.Errorf("error while storing PRs for file %v: %w", path, err)
		}
	}

	return nil
}

func (p *PullRequests) store(path string, prs []int) error {
	return proxycache.Put(
		p.CacheDir,
		categoryFilePrs,
		path,
		prs,
	)
}

func (p *PullRequests) load(path string) ([]int, error) {
	return proxycache.Get(
		context.Background(),
		p.CacheDir,
		categoryFilePrs,
		path,
		nil,
		func(ctx context.Context) ([]int, error) {
			return []int{}, nil
		},
	)
}

func (p *PullRequests) ListPRs(path string) ([]int, error) {
	return p.load(path)
}
