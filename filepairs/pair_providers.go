package filepairs

import "fmt"

type PairProviders struct {
	providers []PairProvider
}

func NewPairProviders(providers ...PairProvider) *PairProviders {
	return &PairProviders{
		providers: providers,
	}
}

func (p *PairProviders) ListPairs(langCode string) ([]Pair, error) {
	var out []Pair

	for _, provider := range p.providers {
		pairs, err := provider.ListPairs(langCode)
		if err != nil {
			return nil, fmt.Errorf("pair provider %s: %w", provider.Name(), err)
		}

		out = append(out, pairs...)
	}

	return out, nil
}
