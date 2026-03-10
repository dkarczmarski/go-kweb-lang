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
	basePath := contentDirPrefix + "/" + langCode

	langPaths, err := p.fileLister.ListFiles(basePath)
	if err != nil {
		return nil, fmt.Errorf("list content files for lang %s: %w", langCode, err)
	}

	pairs := make([]Pair, 0, len(langPaths))

	for _, langPath := range langPaths {
		if !p.acceptLangPath(langPath) {
			continue
		}

		langPath = basePath + "/" + langPath

		enPath, err := p.pairMatcher.LangPath(langPath, "en")
		if err != nil {
			return nil, fmt.Errorf("resolve EN path for %s: %w", langPath, err)
		}

		pairs = append(pairs, Pair{
			EnPath:   enPath,
			LangPath: langPath,
		})
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
