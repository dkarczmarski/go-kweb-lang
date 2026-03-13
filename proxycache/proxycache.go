// Package proxycache provides a wrapper for your function to store results in local files,
// enabling memoization across executions.
package proxycache

//go:generate mockgen -typed -source=proxycache.go -destination=./internal/mocks/mocks.go -package=mocks

import (
	"context"
	"fmt"
)

type Store interface {
	Read(bucket, key string, dst any) (bool, error)
	Write(bucket, key string, data any) error
}

// Get returns a cached value when present and not stale.
// When the cached value is missing or stale, it loads a fresh value using loader,
// stores it in cache, and returns it.
// If isStale is nil, any existing cached value is treated as valid.
//
//nolint:ireturn
func Get[T any](
	ctx context.Context,
	store Store,
	bucket string,
	key string,
	isStale func(T) bool,
	loader func(context.Context) (T, error),
) (T, error) {
	var (
		zero   T
		cached T
	)

	exists, err := store.Read(bucket, key, &cached)
	if err != nil {
		return zero, fmt.Errorf("read cache %s/%s: %w", bucket, key, err)
	}

	if exists && (isStale == nil || !isStale(cached)) {
		return cached, nil
	}

	value, err := loader(ctx)
	if err != nil {
		return zero, fmt.Errorf("load value for cache %s/%s: %w", bucket, key, err)
	}

	if err := store.Write(bucket, key, value); err != nil {
		return zero, fmt.Errorf("write cache %s/%s: %w", bucket, key, err)
	}

	return value, nil
}
