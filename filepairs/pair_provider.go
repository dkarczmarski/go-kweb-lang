package filepairs

// PairProvider lists file pairs for a given language.
//
// A pair connects an English (EN) file with its corresponding
// translated file for the specified language.
//
// A pair may be returned even if only one file exists (either the EN file
// or the language file). At least one of the files in the pair must exist
// for the pair to be included in the result.
type PairProvider interface {
	// Name returns the name of the pair provider.
	Name() string

	// ListPairs returns all file pairs for the given language code.
	ListPairs(langCode string) ([]Pair, error)
}
