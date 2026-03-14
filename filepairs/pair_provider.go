package filepairs

// PairProvider lists file pairs for a given language.
//
// A pair connects an English (EN) file with its corresponding
// translated file for the specified language.
type PairProvider interface {
	// Name returns the name of the pair provider.
	Name() string

	// ListPairs returns all file pairs for the given language code.
	ListPairs(langCode string) ([]Pair, error)
}
