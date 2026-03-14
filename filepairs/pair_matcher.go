package filepairs

// PairMatcher recognizes file paths that belong to a specific
// language pair pattern.
//
// It can detect whether a path is part of a known pair structure
// and return the language code from the matched path.
//
// For English files, the returned language code is "en".
//
// It can also generate the corresponding language file path
// for a given matched path and language code.
type PairMatcher interface {
	// Name returns the name of the matcher.
	Name() string

	// CheckPath checks if the given path matches the pair pattern.
	// It returns whether the path matches and the language code
	// from the matched path.
	CheckPath(path string) (match bool, langCode string, err error)

	// LangPath returns the language file path for the given path
	// and language code.
	LangPath(path string, langCode string) (string, error)
}
