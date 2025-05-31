// Package proxycache provides a wrapper for your function to store results in local files,
// enabling memoization across executions.
package proxycache

//go:generate mockgen -typed -source=proxycache.go -destination=./internal/mocks/mocks.go -package=mocks

import (
	"context"
)

type Store interface {
	Read(bucket, key string, buff any) (bool, error)
	Write(bucket, key string, data any) error
}

// Get returns a value from local cache if it exists, or calls the block to retrieve it.
func Get[T any](
	ctx context.Context,
	store Store,
	bucket string,
	key string,
	isInvalid func(T) bool,
	block func(context.Context) (T, error),
) (T, error) {
	var buff T

	exists, err := store.Read(bucket, key, &buff)
	if err != nil {
		var zero T

		return zero, err
	}

	if exists {
		if isInvalid == nil || !isInvalid(buff) {
			return buff, nil
		}
	}

	result, err := block(ctx)
	if err != nil {
		return result, err
	}

	if err := store.Write(bucket, key, result); err != nil {
		var zero T

		return zero, err
	}

	return result, nil
}
