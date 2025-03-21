package github

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strings"
)

type PRMonitor struct {
	cacheDir string
	tasks    []OnPRUpdateTask
}

type OnPRUpdateTask interface {
	Run(ctx context.Context) error
}

type searchResult struct {
	Items []pullRequest `json:"items"`
}

type pullRequest struct {
	Number    int    `json:"number"`
	UpdatedAt string `json:"updated_at"`
}

func (mon *PRMonitor) maxUpdatedAt(langCode string) (string, error) {
	baseURL := "https://api.github.com/search/issues"

	q := strings.Join(
		[]string{
			"repo:kubernetes/website",
			"is:pr",
			"label:language/" + langCode,
		},
		"+",
	)

	query := url.Values{}
	query.Set("sort", "updated")
	query.Set("order", "desc")
	query.Set("per_page", "1")

	u, err := url.Parse(baseURL)
	if err != nil {
		return "", fmt.Errorf("error while parsing base URL: %v", err)
	}

	u.RawQuery = fmt.Sprintf("q=%s&%s", q, query.Encode())

	urlStr := u.String()

	resp, err := http.Get(urlStr)
	if err != nil {
		return "", fmt.Errorf("error while sending request to GitHub API: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("error while reading GitHub API response: status %s", resp.Status)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("error while reading response body: %v", err)
	}

	var result searchResult
	if err := json.Unmarshal(body, &result); err != nil {
		return "", fmt.Errorf("error while parsing JSON response: %v", err)
	}

	if len(result.Items) == 0 {
		return "", fmt.Errorf("no PRs found")
	}

	return result.Items[0].UpdatedAt, nil
}

func (mon *PRMonitor) lastUpdatedAtFile() string {
	return filepath.Join(mon.cacheDir, "github", "last-updated-at.txt")
}

func (mon *PRMonitor) lastMaxUpdatedAt() (string, error) {
	path := mon.lastUpdatedAtFile()

	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return "", nil
		}

		return "", fmt.Errorf("error while reading file %s: %w", path, err)
	}

	return string(data), nil
}

func (mon *PRMonitor) setLastMaxUpdatedAt(maxUpdatedAt string) error {
	path := mon.lastUpdatedAtFile()

	if err := os.WriteFile(path, []byte(maxUpdatedAt), 0644); err != nil {
		return fmt.Errorf("error while writing to file %s: %w", path, err)
	}

	return nil
}

func (mon *PRMonitor) Check(ctx context.Context, langCode string) (bool, error) {
	lastMaxUpdatedAt, err := mon.lastMaxUpdatedAt()
	if err != nil {
		return false, fmt.Errorf("error while getting the last updatedAt value: %w", err)
	}

	maxUpdatedAt, err := mon.maxUpdatedAt(langCode)
	if err != nil {
		return false, fmt.Errorf("error while getting the maximal updatedAt value: %w", err)
	}

	if lastMaxUpdatedAt == maxUpdatedAt {
		return false, nil
	}

	for _, task := range mon.tasks {
		if err := task.Run(ctx); err != nil {
			log.Printf("task error: %v", err)
		}
	}

	if err := mon.setLastMaxUpdatedAt(maxUpdatedAt); err != nil {
		return true, fmt.Errorf("error while saving the last PRs: %w", err)
	}

	return true, nil
}
