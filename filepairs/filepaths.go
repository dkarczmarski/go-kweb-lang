package filepairs

import (
	"errors"
	"fmt"
	"path/filepath"
	"strings"
)

var ErrPairMatcherNotFound = errors.New("pair matcher not found")

type FilePaths struct {
	pairMatchers []PairMatcher
}

func New() *FilePaths {
	return &FilePaths{
		pairMatchers: []PairMatcher{
			ContentPairMatcher{},
			I18NPairMatcher{},
		},
	}
}

func (fp *FilePaths) CheckPath(path string) (*PathInfo, error) {
	cleanPath := filepath.Clean(path)

	for _, pairMatcher := range fp.pairMatchers {
		match, langCode, err := pairMatcher.CheckPath(cleanPath)
		if err != nil {
			return nil, fmt.Errorf("pair matcher %s: %w", pairMatcher.Name(), err)
		}

		if match {
			return &PathInfo{
				Path:            cleanPath,
				PairMatcherName: pairMatcher.Name(),
				LangCode:        langCode,
				pairMatcher:     pairMatcher,
			}, nil
		}
	}

	return nil, fmt.Errorf("%w: %s", ErrPairMatcherNotFound, cleanPath)
}

func splitPath(path string) []string {
	return strings.Split(filepath.ToSlash(filepath.Clean(path)), "/")
}
