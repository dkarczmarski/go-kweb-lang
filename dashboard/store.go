package dashboard

import "fmt"

// CacheStorage decouples dashboard storage from the concrete cache implementation.
type CacheStorage interface {
	Read(bucket, key string, buff any) (bool, error)
	Write(bucket, key string, data any) error
}

type Store struct {
	cacheStorage CacheStorage
}

func NewStore(cacheStorage CacheStorage) *Store {
	return &Store{
		cacheStorage: cacheStorage,
	}
}

const (
	dashboardIndexBucketName = "dashboard-index"
	singleCacheKey           = ""
)

func LangIndexBucket() string {
	return dashboardIndexBucketName
}

func LangIndexKey() string {
	return singleCacheKey
}

func LangDashboardBucket(langCode string) string {
	return fmt.Sprintf("lang/%s/dashboard", langCode)
}

func LangDashboardKey() string {
	return singleCacheKey
}

func (s *Store) WriteDashboard(dashboard Dashboard) error {
	bucket := LangDashboardBucket(dashboard.LangCode)
	key := LangDashboardKey()

	err := s.cacheStorage.Write(bucket, key, &dashboard)
	if err != nil {
		return fmt.Errorf("write dashboard to cache store bucket=%q key=%q: %w", bucket, key, err)
	}

	return nil
}

func (s *Store) ReadDashboard(langCode string) (Dashboard, error) {
	var (
		dashboard Dashboard
		empty     Dashboard
	)

	bucket := LangDashboardBucket(langCode)
	key := LangDashboardKey()

	found, err := s.cacheStorage.Read(bucket, key, &dashboard)
	if err != nil {
		return empty, fmt.Errorf("read dashboard from cache store bucket=%q key=%q: %w", bucket, key, err)
	}

	if !found {
		return empty, nil
	}

	return dashboard, nil
}

func (s *Store) WriteDashboardIndex(langIndex LangIndex) error {
	bucket := LangIndexBucket()
	key := LangIndexKey()

	err := s.cacheStorage.Write(bucket, key, &langIndex)
	if err != nil {
		return fmt.Errorf("write dashboard index to cache store bucket=%q key=%q: %w", bucket, key, err)
	}

	return nil
}

func (s *Store) ReadDashboardIndex() (LangIndex, error) {
	var (
		langIndex LangIndex
		empty     LangIndex
	)

	bucket := LangIndexBucket()
	key := LangIndexKey()

	found, err := s.cacheStorage.Read(bucket, key, &langIndex)
	if err != nil {
		return empty, fmt.Errorf("read dashboard index from cache store bucket=%q key=%q: %w", bucket, key, err)
	}

	if !found {
		return empty, nil
	}

	return langIndex, nil
}
