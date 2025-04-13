// Package langcnt provides information about the 'content' directory in the kubernetes repository.
package langcnt

import (
	"fmt"
	"os"
	"path/filepath"
	"slices"
)

type LangCodesProvider struct {
	RepoDir string

	langCodesFilter []string
}

// SetLangCodesFilter sets a filter for available lang codes.
func (c *LangCodesProvider) SetLangCodesFilter(langCodes []string) {
	c.langCodesFilter = langCodes
}

// LangCodes returns all lang codes based on the 'content' directory in the Kubernetes repository.
// If a filter is set via SetLangCodesFilter, it omits lang codes that are not in the filter.
func (c *LangCodesProvider) LangCodes() ([]string, error) {
	allLangCodes, err := listDirectories(filepath.Join(c.RepoDir, "content"))
	if err != nil {
		return nil, err
	}

	if len(c.langCodesFilter) == 0 {
		return allLangCodes, nil
	}

	langCodes := make([]string, 0, len(allLangCodes))
	for _, lang := range allLangCodes {
		if !slices.Contains(c.langCodesFilter, lang) {
			continue
		}

		langCodes = append(langCodes, lang)
	}

	return langCodes, nil
}

func listDirectories(path string) ([]string, error) {
	var dirs []string

	files, err := os.ReadDir(path)
	if err != nil {
		return nil, fmt.Errorf("error while listing directory %s: %w", path, err)
	}

	for _, file := range files {
		if file.IsDir() {
			dirs = append(dirs, file.Name())
		}
	}

	return dirs, nil
}
