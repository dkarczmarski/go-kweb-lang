package langcnt

import (
	"os"
	"path/filepath"
	"slices"
)

type Content struct {
	RepoDir string

	allowedLang []string
}

func (c *Content) SetAllowedLang(allowedLang []string) {
	c.allowedLang = allowedLang
}

func (c *Content) Langs() ([]string, error) {
	allLangs, err := listDirectories(filepath.Join(c.RepoDir, "content"))
	if err != nil {
		return nil, err
	}
	if len(c.allowedLang) == 0 {
		return allLangs, nil
	}

	var langs []string
	for _, lang := range allLangs {
		if !slices.Contains(c.allowedLang, lang) {
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
		return nil, err
	}

	for _, file := range files {
		if file.IsDir() {
			dirs = append(dirs, file.Name())
		}
	}

	return dirs, nil
}
