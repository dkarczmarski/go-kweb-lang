//nolint:testpackage
package dashboard

import (
	"errors"
	"strings"
	"testing"
)

type stubLangCodesProvider struct {
	langCodes []string
	err       error
}

func (s stubLangCodesProvider) LangCodes() ([]string, error) {
	if s.err != nil {
		return nil, s.err
	}

	return s.langCodes, nil
}

func TestBuildLangIndex(t *testing.T) {
	t.Parallel()

	t.Run("builds lang index from provider", func(t *testing.T) {
		t.Parallel()

		provider := stubLangCodesProvider{
			langCodes: []string{"pl", "de", "fr"},
		}

		got, err := BuildLangIndex(provider)
		if err != nil {
			t.Fatalf("BuildLangIndex returned error: %v", err)
		}

		if len(got.Items) != 3 {
			t.Fatalf("expected 3 items, got %d", len(got.Items))
		}

		if got.Items[0].LangCode != "pl" {
			t.Fatalf("expected first lang code to be pl, got %q", got.Items[0].LangCode)
		}

		if got.Items[1].LangCode != "de" {
			t.Fatalf("expected second lang code to be de, got %q", got.Items[1].LangCode)
		}

		if got.Items[2].LangCode != "fr" {
			t.Fatalf("expected third lang code to be fr, got %q", got.Items[2].LangCode)
		}
	})

	t.Run("returns wrapped error when provider fails", func(t *testing.T) {
		t.Parallel()

		providerErr := errors.New("boom")
		provider := stubLangCodesProvider{
			err: providerErr,
		}

		got, err := BuildLangIndex(provider)
		if err == nil {
			t.Fatal("expected error, got nil")
		}

		if len(got.Items) != 0 {
			t.Fatalf("expected zero items, got %d", len(got.Items))
		}

		if !strings.Contains(err.Error(), "failed to get available languages") {
			t.Fatalf("expected wrapped error message, got %q", err.Error())
		}

		if !errors.Is(err, providerErr) {
			t.Fatalf("expected error to wrap providerErr, got %v", err)
		}
	})

	t.Run("returns empty index for empty provider result", func(t *testing.T) {
		t.Parallel()

		provider := stubLangCodesProvider{
			langCodes: []string{},
		}

		got, err := BuildLangIndex(provider)
		if err != nil {
			t.Fatalf("BuildLangIndex returned error: %v", err)
		}

		if got.Items == nil {
			t.Fatal("expected non-nil items slice")
		}

		if len(got.Items) != 0 {
			t.Fatalf("expected 0 items, got %d", len(got.Items))
		}
	})
}
