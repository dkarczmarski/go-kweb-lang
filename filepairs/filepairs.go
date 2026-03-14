// Package filepairs provides utilities for working with language file pairs.
//
// A file pair is a relation between an English (EN) source file and its
// translated file for a specific language.
//
// The package contains components that:
//   - detect which pair pattern a file path belongs to,
//   - identify whether a path is an EN file or a language file,
//   - generate the corresponding language file path from an EN file,
//   - list pairs of EN and language files for a given language.
//
// Different pair patterns can be implemented using PairMatcher and
// PairProvider interfaces.
package filepairs
