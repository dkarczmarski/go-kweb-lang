// Package proxycache provides a wrapper for your function to store results in local files,
// enabling memoization across executions.
package proxycache

import (
	"context"
	"crypto/sha1"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
)

// Get returns a value from local cache if it exists, or calls the block to retrieve it.
func Get[T any](
	ctx context.Context,
	cacheDir string,
	category string,
	key string,
	isInvalid func(T) bool,
	block func(context.Context) (T, error),
) (T, error) {
	select {
	case <-ctx.Done():
		var zero T

		return zero, fmt.Errorf("error while getting key %v: %w", key, ctx.Err())
	default:
	}

	cacheFile := keyFilePath(cacheDir, category, key)

	if err := ensureDir(filepath.Dir(cacheFile)); err != nil {
		var zero T

		return zero, err
	}

	if keyFileExists(cacheFile) {
		var buff T
		if err := readJSONFromFile(cacheFile, &buff); err != nil {
			return buff, err
		}

		if isInvalid == nil || !isInvalid(buff) {
			return buff, nil
		}
	}

	result, err := block(ctx)
	if err != nil {
		return result, err
	}

	if err := writeJSONToFile(cacheFile, result); err != nil {
		var zero T

		return zero, err
	}

	return result, nil
}

// InvalidateKey removes the item from the cache for the given key.
func InvalidateKey(cacheDir, category, key string) error {
	cacheFile := keyFilePath(cacheDir, category, key)

	if err := removeFile(cacheFile); err != nil {
		return fmt.Errorf("error while removing file %v: %w", cacheFile, err)
	}

	return nil
}

// Put inserts the item into the cache.
func Put[T any](cacheDir, category, key string, value T) error {
	cacheFile := keyFilePath(cacheDir, category, key)

	if err := ensureDir(filepath.Dir(cacheFile)); err != nil {
		return err
	}

	return writeJSONToFile(cacheFile, value)
}

// KeyExists checks if the item exists for the given key.
func KeyExists(cacheDir, category, key string) bool {
	return keyFileExists(keyFilePath(cacheDir, category, key))
}

func keyFileExists(cacheFile string) bool {
	_, err := os.Stat(cacheFile)
	if os.IsNotExist(err) {
		return false
	}

	return err == nil
}

func keyFileName(key string) string {
	return key + ".json"
}

func keyHash(value string) string {
	hash := sha1.New()
	hash.Write([]byte(value))

	return hex.EncodeToString(hash.Sum(nil))
}

func keyFilePath(cacheDir string, category string, key string) string {
	return filepath.Join(cacheDir, category, keyFileName(keyHash(key)))
}

func ensureDir(path string) error {
	info, err := os.Stat(path)
	if err != nil {
		if !os.IsNotExist(err) {
			return fmt.Errorf("failed to check directory: %w", err)
		}

		if err := os.MkdirAll(path, 0o755); err != nil {
			return fmt.Errorf("failed to create directory %s: %w", path, err)
		}

		return nil
	}

	if !info.IsDir() {
		return fmt.Errorf("%s already exists but is not a directory: %w", path, os.ErrNotExist)
	}

	return nil
}

func readJSONFromFile(path string, v any) error {
	b, err := os.ReadFile(path)
	if err != nil {
		return err
	}

	if err := json.Unmarshal(b, v); err != nil {
		return fmt.Errorf("error while unmarhalling: %w", err)
	}

	return nil
}

func writeJSONToFile(path string, v any) error {
	data, err := json.MarshalIndent(v, "", "\t")
	if err != nil {
		return fmt.Errorf("json marshal error: %w", err)
	}

	tmpPath := path + ".tmp"
	if err := os.WriteFile(tmpPath, data, 0o644); err != nil {
		return fmt.Errorf("write file %s error: %w", tmpPath, err)
	}

	if err := os.Rename(tmpPath, path); err != nil {
		return fmt.Errorf("rename file %s error: %w", tmpPath, err)
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
