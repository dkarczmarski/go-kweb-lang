package proxycache

import (
	"context"
	"crypto/sha1"
	"encoding/hex"
	"fmt"
	"os"
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

func GetCtx[T any](
	ctx context.Context,
	cacheDir string,
	key string,
	isInvalid func(T) bool,
	block func() (T, error),
) (T, error) {
	select {
	case <-ctx.Done():
		var zero T
		return zero, ctx.Err()
	default:
	}

	if err := EnsureDir(cacheDir); err != nil {
		var zero T
		return zero, err
	}

	key = KeyHash(key)
	cachePath := filepath.Join(cacheDir, KeyFile(key))
	if FileExists(cachePath) {
		var buff T
		if err := ReadJSONFromFile(cachePath, &buff); err != nil {
			return buff, err
		}

		if isInvalid == nil || !isInvalid(buff) {
			return buff, nil
		}
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

func InvalidateKey(cacheDir string, key string) error {
	cacheFile := filepath.Join(cacheDir, KeyFile(key))

	if err := removeFile(cacheFile); err != nil {
		return fmt.Errorf("error while removing file %v: %w", cacheFile, err)
	}

	return nil
}

func removeFile(path string) error {
	_, err := os.Stat(path)
	if err != nil {
		if os.IsNotExist(err) {
			// skip silently
			return nil
		}
		return fmt.Errorf("failed to check file: %w", err)
	}

	err = os.Remove(path)
	if err != nil {
		return fmt.Errorf("failed to remove file %s: %w", path, err)
	}

	return nil
}
