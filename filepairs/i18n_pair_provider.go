package filepairs

import (
	"fmt"
	"path/filepath"
)

type I18NFileChecker interface {
	FileExists(path string) (bool, error)
}

type I18NPairProvider struct {
	files I18NFileChecker
}

func NewI18NPairProvider(files I18NFileChecker) *I18NPairProvider {
	return &I18NPairProvider{
		files: files,
	}
}

func (p *I18NPairProvider) Name() string {
	return i18nDirPrefix
}

func (p *I18NPairProvider) ListPairs(langCode string) ([]Pair, error) {
	langPath := i18nFilePath(langCode)
	enPath := i18nFilePath(enLangCode)

	exists, err := p.files.FileExists(langPath)
	if err != nil {
		return nil, fmt.Errorf("check i18n lang file %s: %w", langPath, err)
	}

	if exists {
		return []Pair{{
			EnPath:   enPath,
			LangPath: langPath,
		}}, nil
	}

	exists, err = p.files.FileExists(enPath)
	if err != nil {
		return nil, fmt.Errorf("check i18n EN file %s: %w", enPath, err)
	}

	if !exists {
		return []Pair{}, nil
	}

	return []Pair{{
		EnPath:   enPath,
		LangPath: langPath,
	}}, nil
}

func i18nFilePath(langCode string) string {
	return filepath.Join(i18nDirPrefix, langCode, langCode+".toml")
}
