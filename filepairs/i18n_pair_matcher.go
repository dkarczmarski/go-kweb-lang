package filepairs

import (
	"errors"
	"fmt"
	"path/filepath"
)

const (
	i18nDirPrefix    = "i18n"
	i18nPathPartsLen = 3
)

var ErrInvalidI18NPath = errors.New("invalid i18n path")

type I18NPairMatcher struct{}

func (m I18NPairMatcher) Name() string {
	return i18nDirPrefix
}

func (m I18NPairMatcher) CheckPath(path string) (bool, string, error) {
	parts := splitPath(path)

	if len(parts) != i18nPathPartsLen {
		return false, "", nil
	}

	if parts[0] != i18nDirPrefix {
		return false, "", nil
	}

	langCode := parts[1]
	if langCode == "" {
		return false, "", ErrInvalidLanguageCode
	}

	expectedFileName := langCode + ".toml"
	if parts[2] != expectedFileName {
		return false, "", nil
	}

	return true, langCode, nil
}

func (m I18NPairMatcher) LangPath(path string, langCode string) (string, error) {
	match, _, err := m.CheckPath(path)
	if err != nil {
		return "", err
	}

	if !match {
		return "", fmt.Errorf("%w: %s", ErrInvalidI18NPath, path)
	}

	return filepath.Join(i18nDirPrefix, langCode, langCode+".toml"), nil
}
