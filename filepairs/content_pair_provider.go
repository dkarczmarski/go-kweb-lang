package filepairs

import "fmt"

type ContentFilesLister interface {
	ListFiles(path string) ([]string, error)
}

type ContentPairProvider struct {
	fileLister  ContentFilesLister
	pairMatcher ContentPairMatcher
}

func NewContentPairProvider(fileLister ContentFilesLister) *ContentPairProvider {
	return &ContentPairProvider{
		fileLister:  fileLister,
		pairMatcher: ContentPairMatcher{},
	}
}

func (p *ContentPairProvider) Name() string {
	return contentDirPrefix
}

func (p *ContentPairProvider) ListPairs(langCode string) ([]Pair, error) {
	enBasePath := contentDirPrefix + "/" + enLangCode
	langBasePath := contentDirPrefix + "/" + langCode

	langPaths, err := p.fileLister.ListFiles(langBasePath)
	if err != nil {
		return nil, fmt.Errorf("list content files for lang %s: %w", langCode, err)
	}

	enPaths, err := p.fileLister.ListFiles(enBasePath)
	if err != nil {
		return nil, fmt.Errorf("list content files for EN: %w", err)
	}

	pairs := make([]Pair, 0, len(langPaths)+len(enPaths))
	addedLangPaths := make(map[string]struct{}, len(langPaths)+len(enPaths))

	pairs, err = p.appendPairsFromLangPaths(pairs, addedLangPaths, langBasePath, langPaths)
	if err != nil {
		return nil, err
	}

	pairs, err = p.appendMissingPairsFromEnPaths(pairs, addedLangPaths, enBasePath, enPaths, langCode)
	if err != nil {
		return nil, err
	}

	return pairs, nil
}

func (p *ContentPairProvider) appendPairsFromLangPaths(
	pairs []Pair,
	addedLangPaths map[string]struct{},
	langBasePath string,
	langPaths []string,
) ([]Pair, error) {
	for _, langPath := range langPaths {
		if !p.acceptLangPath(langPath) {
			continue
		}

		fullLangPath := langBasePath + "/" + langPath

		enPath, err := p.pairMatcher.LangPath(fullLangPath, enLangCode)
		if err != nil {
			return nil, fmt.Errorf("resolve EN path for %s: %w", fullLangPath, err)
		}

		pairs = append(pairs, Pair{
			EnPath:   enPath,
			LangPath: fullLangPath,
		})

		addedLangPaths[fullLangPath] = struct{}{}
	}

	return pairs, nil
}

func (p *ContentPairProvider) appendMissingPairsFromEnPaths(
	pairs []Pair,
	addedLangPaths map[string]struct{},
	enBasePath string,
	enPaths []string,
	langCode string,
) ([]Pair, error) {
	for _, enPath := range enPaths {
		if !p.acceptLangPath(enPath) {
			continue
		}

		fullEnPath := enBasePath + "/" + enPath

		langPath, err := p.pairMatcher.LangPath(fullEnPath, langCode)
		if err != nil {
			return nil, fmt.Errorf("resolve lang path for %s: %w", fullEnPath, err)
		}

		if _, exists := addedLangPaths[langPath]; exists {
			continue
		}

		pairs = append(pairs, Pair{
			EnPath:   fullEnPath,
			LangPath: langPath,
		})

		addedLangPaths[langPath] = struct{}{}
	}

	return pairs, nil
}

//nolint:gosimple
func (p *ContentPairProvider) acceptLangPath(langPath string) bool {
	if langPath == "OWNERS" {
		return false
	}

	return true
}
