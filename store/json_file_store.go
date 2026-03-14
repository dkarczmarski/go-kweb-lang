package store

import (
	"crypto/sha1" //nolint:gosec
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
)

const (
	dirPerm  = 0o755
	filePerm = 0o600
)

// ErrNotDirectory indicates that a path exists but is not a directory.
var ErrNotDirectory = errors.New("path exists but is not a directory")

// JSONFileStore stores JSON-encoded values in files under a bucket/key layout.
type JSONFileStore struct {
	cacheDir string
}

// NewFileStore creates a new JSON-backed file store rooted at cacheDir.
func NewFileStore(cacheDir string) *JSONFileStore {
	return &JSONFileStore{cacheDir: cacheDir}
}

// Read reads a JSON value from the given bucket and key into dst.
// It returns false with a nil error when the entry does not exist.
func (fs *JSONFileStore) Read(bucket, key string, dst any) (bool, error) {
	cacheFile := keyFilePath(fs.cacheDir, bucket, key)

	exists, err := keyFileExists(cacheFile)
	if err != nil {
		return false, fmt.Errorf("check cache file %s: %w", cacheFile, err)
	}

	if !exists {
		return false, nil
	}

	if err := readJSONFromFile(cacheFile, dst); err != nil {
		return false, fmt.Errorf("read cache file %s: %w", cacheFile, err)
	}

	return true, nil
}

func keyFileExists(cacheFile string) (bool, error) {
	_, err := os.Stat(cacheFile)
	if err != nil {
		if os.IsNotExist(err) {
			return false, nil
		}

		return false, fmt.Errorf("stat cache file %s: %w", cacheFile, err)
	}

	return true, nil
}

func readJSONFromFile(path string, dst any) error {
	data, err := os.ReadFile(path)
	if err != nil {
		return fmt.Errorf("read file %s: %w", path, err)
	}

	if err := json.Unmarshal(data, dst); err != nil {
		return fmt.Errorf("unmarshal JSON from %s: %w", path, err)
	}

	return nil
}

// Write writes data as JSON to the given bucket and key.
func (fs *JSONFileStore) Write(bucket, key string, data any) error {
	cacheFile := keyFilePath(fs.cacheDir, bucket, key)

	if err := ensureDir(filepath.Dir(cacheFile)); err != nil {
		return fmt.Errorf("ensure cache directory for %s: %w", cacheFile, err)
	}

	if err := writeJSONToFile(cacheFile, data); err != nil {
		return fmt.Errorf("write cache file %s: %w", cacheFile, err)
	}

	return nil
}

func writeJSONToFile(path string, data any) error {
	jsonData, err := json.MarshalIndent(data, "", "\t")
	if err != nil {
		return fmt.Errorf("marshal JSON: %w", err)
	}

	tmpPath := path + ".tmp"

	if err := os.WriteFile(tmpPath, jsonData, filePerm); err != nil {
		return fmt.Errorf("write temporary file %s: %w", tmpPath, err)
	}

	if err := os.Rename(tmpPath, path); err != nil {
		return fmt.Errorf("rename temporary file %s to %s: %w", tmpPath, path, err)
	}

	return nil
}

func keyFileName(key string) string {
	if key == "" {
		return "_single.json"
	}

	return keyHash(key) + ".json"
}

// keyHash returns a stable hash suitable for cache file names.
// SHA-1 is used here for compact, deterministic file naming rather than
// for any cryptographic security purpose.
func keyHash(value string) string {
	//nolint:gosec
	hash := sha1.New()
	_, _ = hash.Write([]byte(value))

	return hex.EncodeToString(hash.Sum(nil))
}

func keyFilePath(cacheDir, bucket, key string) string {
	return filepath.Join(cacheDir, bucket, keyFileName(key))
}

func ensureDir(path string) error {
	info, err := os.Stat(path)
	if err != nil {
		if !os.IsNotExist(err) {
			return fmt.Errorf("stat directory %s: %w", path, err)
		}

		if err := os.MkdirAll(path, dirPerm); err != nil {
			return fmt.Errorf("create directory %s: %w", path, err)
		}

		return nil
	}

	if !info.IsDir() {
		return fmt.Errorf("%w: %s", ErrNotDirectory, path)
	}

	return nil
}

// Delete removes the entry for the given bucket and key.
// Missing files are ignored.
func (fs *JSONFileStore) Delete(bucket, key string) error {
	cacheFile := keyFilePath(fs.cacheDir, bucket, key)

	if err := removeFile(cacheFile); err != nil {
		return fmt.Errorf("delete cache file %s: %w", cacheFile, err)
	}

	return nil
}

func removeFile(path string) error {
	err := os.Remove(path)
	if err != nil {
		if os.IsNotExist(err) {
			return nil
		}

		return fmt.Errorf("remove file %s: %w", path, err)
	}

	return nil
}

// ListBuckets lists subdirectories under the given bucket path prefix.
func (fs *JSONFileStore) ListBuckets(bucketPath string) ([]string, error) {
	dirs, err := listSubdirectories(filepath.Join(fs.cacheDir, bucketPath))
	if err != nil {
		return nil, fmt.Errorf("list buckets under %s: %w", bucketPath, err)
	}

	return dirs, nil
}

func listSubdirectories(path string) ([]string, error) {
	entries, err := os.ReadDir(path)
	if err != nil {
		if os.IsNotExist(err) {
			return []string{}, nil
		}

		return nil, fmt.Errorf("read directory %s: %w", path, err)
	}

	dirs := make([]string, 0, len(entries))

	for _, entry := range entries {
		if entry.IsDir() {
			dirs = append(dirs, entry.Name())
		}
	}

	return dirs, nil
}
