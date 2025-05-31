package store

import (
	"crypto/sha1"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
)

type FileStore struct {
	cacheDir string
}

func NewFileStore(cacheDir string) *FileStore {
	return &FileStore{cacheDir: cacheDir}
}

func (fs *FileStore) Read(bucket, key string, buff any) (bool, error) {
	cacheFile := keyFilePath(fs.cacheDir, bucket, key)

	exists, err := keyFileExists(cacheFile)
	if err != nil {
		return false, err
	}

	if err := ensureDir(filepath.Dir(cacheFile)); err != nil {
		return false, err
	}

	if !exists {
		return false, nil
	}

	if err := readJSONFromFile(cacheFile, buff); err != nil {
		return false, err
	}

	return true, nil
}

func keyFileExists(cacheFile string) (bool, error) {
	_, err := os.Stat(cacheFile)
	if err != nil {
		if os.IsNotExist(err) {
			return false, nil
		}

		return false, err
	}

	return true, nil
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

func (fs *FileStore) Write(bucket, key string, v any) error {
	cacheFile := keyFilePath(fs.cacheDir, bucket, key)

	if err := ensureDir(filepath.Dir(cacheFile)); err != nil {
		return err
	}

	return writeJSONToFile(cacheFile, v)
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

func (fs *FileStore) Delete(bucket, key string) error {
	cacheFile := keyFilePath(fs.cacheDir, bucket, key)

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

func (fs *FileStore) ListBuckets(bucketPth string) ([]string, error) {
	return listSubdirectories(filepath.Join(fs.cacheDir, bucketPth))
}

func listSubdirectories(path string) ([]string, error) {
	entries, err := os.ReadDir(path)
	if err != nil {
		if os.IsNotExist(err) {
			return []string{}, nil
		}

		return nil, fmt.Errorf("failed to read directory %q: %w", path, err)
	}

	var dirs []string
	for _, entry := range entries {
		if entry.IsDir() {
			dirs = append(dirs, entry.Name())
		}
	}

	return dirs, nil
}
