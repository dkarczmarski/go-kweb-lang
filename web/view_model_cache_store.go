package web

import (
	"context"
	"fmt"

	"go-kweb-lang/proxycache"
	"go-kweb-lang/web/internal/view"
)

// CacheStore is an interface used to decouple this package from the concrete store.FileStore implementation.
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

func (s *ViewModelCacheStore) GetLangCodes() (*view.LangCodesViewModel, error) {
	return proxycache.Get(
		context.Background(), // this context is not used
		s.cacheStore,
		bucketLangCodesView,
		singleKey,
		nil,
		func(_ context.Context) (*view.LangCodesViewModel, error) {
			return &view.LangCodesViewModel{}, nil
		},
	)
}

func (s *ViewModelCacheStore) SetLangCodes(model *view.LangCodesViewModel) error {
	return s.cacheStore.Write(bucketLangCodesView, singleKey, model)
}

func (s *ViewModelCacheStore) GetLangDashboardFiles(langCode string) (view.LangDashboardFilesModel, error) {
	return proxycache.Get(
		context.Background(), // this context is not used
		s.cacheStore,
		langDashboardFilesBucketName(langCode),
		singleKey,
		nil,
		func(_ context.Context) (view.LangDashboardFilesModel, error) {
			return view.LangDashboardFilesModel{}, nil
		},
	)
}

func (s *ViewModelCacheStore) SetLangDashboardFiles(langCode string, langDashboardFiles view.LangDashboardFilesModel) error {
	return s.cacheStore.Write(langDashboardFilesBucketName(langCode), singleKey, langDashboardFiles)
}

func langDashboardFilesBucketName(langCode string) string {
	return fmt.Sprintf("lang/%s/view-lang-dashboard-files", langCode)
}
