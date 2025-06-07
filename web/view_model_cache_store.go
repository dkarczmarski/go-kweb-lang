package web

import (
	"context"
	"fmt"

	"go-kweb-lang/proxycache"
)

// CacheStore is an interface used to decouple this package from the concrete store implementation
type CacheStore interface {
	Read(bucket, key string, buff any) (bool, error)
	Write(bucket, key string, data any) error
	Delete(bucket, key string) error
}

type ViewModelCacheStore struct {
	cacheStore CacheStore
}

func NewViewModelCacheStore(cacheStore CacheStore) *ViewModelCacheStore {
	return &ViewModelCacheStore{
		cacheStore: cacheStore,
	}
}

const (
	bucketLangCodesView = "view-lang-codes"
	singleKey           = ""
)

func (s *ViewModelCacheStore) GetLangCodes() (*LangCodesViewModel, error) {
	return proxycache.Get(
		context.Background(),
		s.cacheStore,
		bucketLangCodesView,
		singleKey,
		nil,
		func(ctx context.Context) (*LangCodesViewModel, error) {
			return &LangCodesViewModel{}, nil
		},
	)
}

func (s *ViewModelCacheStore) SetLangCodes(model *LangCodesViewModel) error {
	return s.cacheStore.Write(bucketLangCodesView, singleKey, model)
}

func (s *ViewModelCacheStore) GetLangDashboard(langCode string) (*LangDashboardViewModel, error) {
	return proxycache.Get(
		context.Background(),
		s.cacheStore,
		langDashboardBucketName(langCode),
		singleKey,
		nil,
		func(ctx context.Context) (*LangDashboardViewModel, error) {
			return nil, nil
		},
	)
}

func (s *ViewModelCacheStore) SetLangDashboard(langCode string, model *LangDashboardViewModel) error {
	return s.cacheStore.Write(langDashboardBucketName(langCode), singleKey, model)
}

func langDashboardBucketName(langCode string) string {
	return fmt.Sprintf("lang/%v/view-lang-dashboard", langCode)
}
