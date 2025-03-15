package web

import (
	"go-kweb-lang/langcnt"
)

func BuildIndexModel(content *langcnt.Content) ([]LinkModel, error) {
	langs, err := content.Langs()
	if err != nil {
		return nil, err
	}

	var model []LinkModel

	for _, lang := range langs {
		model = append(model, LinkModel{
			Text: lang,
			Url:  "lang/" + lang,
		})
	}

	return model, nil
}
