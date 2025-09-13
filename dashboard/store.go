package dashboard

import (
	"context"
	"fmt"

	"go-kweb-lang/proxycache"
)

// CacheStore is an interface used to decouple this package from the concrete store.FileStore implementation.
type CacheStore interface {
	Read(bucket, key string, buff any) (bool, error)
	Write(bucket, key string, data any) error
	Delete(bucket, key string) error
}

type Store struct {
	cacheStore CacheStore
}

func NewStore(cacheStore CacheStore) *Store {
	return &Store{
		cacheStore: cacheStore,
	}
}

const (
	langIndexBucket = "dashboard-index"
	singleKey       = ""
)

func (s *Store) WriteDashboard(dashboard *Dashboard) error {
	return s.cacheStore.Write(langDashboardBucket(dashboard.LangCode), singleKey, dashboard)
}

func (s *Store) ReadDashboard(langCode string) (*Dashboard, error) {
	return proxycache.Get(
		context.Background(), // this context is not used
		s.cacheStore,
		langDashboardBucket(langCode),
		singleKey,
		nil,
		func(_ context.Context) (*Dashboard, error) {
			return &Dashboard{}, nil
		},
	)
}

func (s *Store) WriteDashboardIndex(langIndex *LangIndex) error {
	return s.cacheStore.Write(langIndexBucket, singleKey, langIndex)
}

func (s *Store) ReadDashboardIndex() (*LangIndex, error) {
	return proxycache.Get(
		context.Background(), // this context is not used
		s.cacheStore,
		langIndexBucket,
		singleKey,
		nil,
		func(_ context.Context) (*LangIndex, error) {
			return &LangIndex{}, nil
		},
	)
}

func langDashboardBucket(langCode string) string {
	return fmt.Sprintf("lang/%s/dashboard", langCode)
}
