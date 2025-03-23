package filecache

import (
	"context"
	"crypto/sha1"
	"encoding/hex"
	"path/filepath"
)

func KeyFile(key string) string {
	return key + ".json"
}

func KeyHash(value string) string {
	hash := sha1.New()
	hash.Write([]byte(value))

	return hex.EncodeToString(hash.Sum(nil))
}

func CacheWrapperCtx[T any](ctx context.Context, cacheDir string, key string, block func() (T, error)) (T, error) {
	select {
	case <-ctx.Done():
		var zero T
		return zero, ctx.Err()
	default:
	}

	return CacheWrapper(cacheDir, key, block)
}

func CacheWrapper[T any](cacheDir string, key string, block func() (T, error)) (T, error) {
	if err := EnsureDir(cacheDir); err != nil {
		var zero T
		return zero, err
	}

	hash := KeyHash(key)
	cachePath := filepath.Join(cacheDir, KeyFile(hash))
	if FileExists(cachePath) {
		var buff T
		if err := ReadJSONFromFile(cachePath, &buff); err != nil {
			return buff, err
		}

		return buff, nil
	}

	result, err := block()
	if err != nil {
		return result, err
	}

	if err := WriteJSONToFile(cachePath, result); err != nil {
		var zero T
		return zero, err
	}

	return result, nil
}
