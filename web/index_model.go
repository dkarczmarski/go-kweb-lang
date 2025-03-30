package web

import (
	"fmt"

	"go-kweb-lang/langcnt"
)

func BuildIndexModel(content *langcnt.Content) ([]LinkModel, error) {
	langCodes, err := content.LangCodes()
	if err != nil {
		return nil, fmt.Errorf("error while getting available languages: %w", err)
	}

	model := make([]LinkModel, 0, len(langCodes))
	for _, langCode := range langCodes {
		model = append(model, LinkModel{
			Text: langCode,
			URL:  "lang/" + langCode,
		})
	}

	return model, nil
}
