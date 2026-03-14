// Package langcnt provides language code information from the content
// directory in the Kubernetes repository.
package langcnt

import (
	"fmt"
	"os"
	"path/filepath"
	"slices"
)

const contentDirName = "content"

type LangCodesProvider struct {
	RepoDir string

	langCodesFilter []string
}

// SetLangCodesFilter sets the allowed language codes.
// When the filter is empty, all detected language codes are returned.
func (p *LangCodesProvider) SetLangCodesFilter(langCodes []string) {
	p.langCodesFilter = langCodes
}

// LangCodes returns language codes found in the repository content directory.
//
// The "en" directory is excluded.
//
// If a filter was set with SetLangCodesFilter, only language codes present in
// that filter are returned.
func (p *LangCodesProvider) LangCodes() ([]string, error) {
	allLangCodes, err := listLangDirectories(filepath.Join(p.RepoDir, contentDirName))
	if err != nil {
		return nil, err
	}

	if len(p.langCodesFilter) == 0 {
		return allLangCodes, nil
	}

	langCodes := make([]string, 0, len(allLangCodes))

	for _, langCode := range allLangCodes {
		if !slices.Contains(p.langCodesFilter, langCode) {
			continue
		}

		langCodes = append(langCodes, langCode)
	}

	return langCodes, nil
}

func listLangDirectories(path string) ([]string, error) {
	entries, err := os.ReadDir(path)
	if err != nil {
		return nil, fmt.Errorf("read directory %s: %w", path, err)
	}

	langCodes := make([]string, 0, len(entries))

	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}

		langCode := entry.Name()
		if langCode == "en" {
			continue
		}

		langCodes = append(langCodes, langCode)
	}

	return langCodes, nil
}
