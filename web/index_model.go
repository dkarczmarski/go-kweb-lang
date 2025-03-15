package web

import (
	"fmt"
	"go-kweb-lang/langcnt"
)

func BuildIndexModel(content *langcnt.Content) ([]LinkModel, error) {
	langs, err := content.Langs()
	if err != nil {
		return nil, fmt.Errorf("error while getting available languages: %w", err)
	}

	model := make([]LinkModel, 0, len(langs))
	for _, lang := range langs {
		model = append(model, LinkModel{
			Text: lang,
			URL:  "lang/" + lang,
		})
	}

	return model, nil
}
