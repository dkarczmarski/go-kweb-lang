// Package langcnt provides information about the 'content' directory in the kubernetes repository.
package langcnt

import (
	"fmt"
	"os"
	"path/filepath"
	"slices"
)

type Content struct {
	RepoDir string

	allowedLangCodes []string
}

// SetLangCodes sets a filter for available lang codes.
func (c *Content) SetLangCodes(langCodes []string) {
	c.allowedLangCodes = langCodes
}

// LangCodes returns all lang codes based on the 'content' directory in the Kubernetes repository.
// If a filter is set via SetLangCodes, it omits lang codes that are not in the filter.
func (c *Content) LangCodes() ([]string, error) {
	allLangs, err := listDirectories(filepath.Join(c.RepoDir, "content"))
	if err != nil {
		return nil, err
	}

	if len(c.allowedLangCodes) == 0 {
		return allLangs, nil
	}

	langs := make([]string, 0, len(allLangs))
	for _, lang := range allLangs {
		if !slices.Contains(c.allowedLangCodes, lang) {
			continue
		}

		langs = append(langs, lang)
	}

	return langs, nil
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
