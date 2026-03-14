package filepairs

import (
	"errors"
	"fmt"
)

const enLangCode = "en"

var (
	ErrLangPathRequiresEnPath = errors.New("lang path can be resolved only for EN path")
	ErrInvalidInput           = errors.New("invalid input")
)

type PathInfo struct {
	Path            string
	PairMatcherName string
	LangCode        string

	pairMatcher PairMatcher
}

func (pi *PathInfo) IsEnPath() bool {
	return pi != nil && pi.LangCode == enLangCode
}

func (pi *PathInfo) LangPath(langCode string) (string, error) {
	if pi == nil {
		return "", fmt.Errorf("%w: path info is nil", ErrInvalidInput)
	}

	if !pi.IsEnPath() {
		return "", fmt.Errorf("%w: %s", ErrLangPathRequiresEnPath, pi.Path)
	}

	if pi.pairMatcher == nil {
		return "", fmt.Errorf("%w: pair matcher is nil for path %s", ErrInvalidInput, pi.Path)
	}

	langPath, err := pi.pairMatcher.LangPath(pi.Path, langCode)
	if err != nil {
		return "", fmt.Errorf("pair matcher %s: %w", pi.PairMatcherName, err)
	}

	return langPath, nil
}
