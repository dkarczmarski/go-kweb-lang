package pullreq

import (
	"encoding/json"
	"fmt"
	"go-kweb-lang/filecache"
	"go-kweb-lang/github"
	"log"
	"os"
	"path/filepath"
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

func (p *PullRequests) fetchPRCommits(pr github.PRItem) ([]string, error) {
	key := fmt.Sprintf("%v", pr.Number)
	commits, err := filecache.CacheWrapper(
		filepath.Join(p.CacheDir, "pr", "pr-commits"),
		key,
		func(cachedPrCommits prCommits) bool {
			isInvalid := cachedPrCommits.UpdatedAt != pr.UpdatedAt

			if isInvalid {
				log.Printf("PR #%v hit cache but is out of date: %v", pr.Number, cachedPrCommits.UpdatedAt)
			}

			return isInvalid
		},
		func() (prCommits, error) {
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
	return filecache.CacheWrapper(
		filepath.Join(p.CacheDir, "pr", "commit-files"),
		commitID,
		nil,
		func() (*github.CommitFiles, error) {
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

	return filePRs
}

func (p *PullRequests) Update(langCode string) error {
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

		commitIds, err := p.fetchPRCommits(pr)
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

type filePRsRecord struct {
	Path string // this field is not necessary
	PRs  []int
}

func (p *PullRequests) storeAll(filePRs map[string][]int) error {
	for path, prs := range filePRs {
		if err := p.store(path, prs); err != nil {
			return fmt.Errorf("error while storing PRs for file %v: %w", path, err)
		}
	}

	return nil
}

func (p *PullRequests) storageFile(path string) string {
	return filepath.Join(p.CacheDir, "pr", "file-prs", filecache.KeyFile(filecache.KeyHash(path)))
}

func (p *PullRequests) store(path string, prs []int) error {
	record := filePRsRecord{
		Path: path,
		PRs:  prs,
	}

	b, err := json.MarshalIndent(&record, "", "\t")
	if err != nil {
		return fmt.Errorf("error while marshalling data: %w", err)
	}

	storageFile := p.storageFile(path)
	if err := filecache.EnsureDir(filepath.Dir(storageFile)); err != nil {
		return fmt.Errorf("error while checking parent directories for %v: %w",
			storageFile, err)
	}
	if err := os.WriteFile(storageFile, b, 0644); err != nil {
		return fmt.Errorf("error while writing file %v: %w", storageFile, err)
	}

	return nil
}

func (p *PullRequests) load(path string) ([]int, error) {
	storageFile := p.storageFile(path)

	b, err := os.ReadFile(storageFile)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}

		return nil, fmt.Errorf("error while reading file %v: %w", storageFile, err)
	}

	var buff filePRsRecord
	if err := json.Unmarshal(b, &buff); err != nil {
		return nil, fmt.Errorf("error while unmashalling data: %w", err)
	}

	return buff.PRs, nil
}

func (p *PullRequests) ListPRs(file string) ([]int, error) {
	return p.load(file)
}
