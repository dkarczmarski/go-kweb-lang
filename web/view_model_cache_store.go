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

func (s *ViewModelCacheStore) GetLangDashboardFiles(langCode string) ([]FileModel, error) {
	return proxycache.Get(
		context.Background(),
		s.cacheStore,
		langDashboardFilesBucketName(langCode),
		singleKey,
		nil,
		func(ctx context.Context) ([]FileModel, error) {
			return nil, nil
		},
	)
}

func (s *ViewModelCacheStore) SetLangDashboardFiles(langCode string, files []FileModel) error {
	return s.cacheStore.Write(langDashboardFilesBucketName(langCode), singleKey, files)
}

func langDashboardFilesBucketName(langCode string) string {
	return fmt.Sprintf("lang/%v/view-lang-dashboard-files", langCode)
}
