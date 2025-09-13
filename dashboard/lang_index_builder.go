package dashboard

import "fmt"

type LangCodesProvider interface {
	LangCodes() ([]string, error)
}

func buildLangIndex(langCodesProvider LangCodesProvider) (*LangIndex, error) {
	langCodes, err := langCodesProvider.LangCodes()
	if err != nil {
		return nil, fmt.Errorf("failed to get available languages: %w", err)
	}

	items := make([]LangIndexItem, 0, len(langCodes))
	for _, langCode := range langCodes {
		items = append(items, LangIndexItem{
			LangCode: langCode,
		})
	}

	return &LangIndex{
		Items: items,
	}, nil
}
