package filepairs

import (
	"errors"
	"fmt"
	"path/filepath"
)

const (
	contentDirPrefix    = "content"
	contentPathMinParts = 3
)

var (
	ErrInvalidLanguageCode = errors.New("invalid language code")
	ErrInvalidContentPath  = errors.New("invalid content path")
)

type ContentPairMatcher struct{}

func (r ContentPairMatcher) Name() string {
	return contentDirPrefix
}

func (r ContentPairMatcher) CheckPath(path string) (bool, string, error) {
	parts := splitPath(path)

	if len(parts) < contentPathMinParts {
		return false, "", nil
	}

	if parts[0] != contentDirPrefix {
		return false, "", nil
	}

	langCode := parts[1]
	if langCode == "" {
		return false, "", ErrInvalidLanguageCode
	}

	return true, langCode, nil
}

func (r ContentPairMatcher) LangPath(path string, langCode string) (string, error) {
	parts := splitPath(path)

	if len(parts) < contentPathMinParts {
		return "", fmt.Errorf("%w: %s", ErrInvalidContentPath, path)
	}

	rel := parts[2:]
	out := append([]string{contentDirPrefix, langCode}, rel...)

	return filepath.Join(out...), nil
}
