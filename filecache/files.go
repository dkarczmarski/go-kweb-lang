package filecache

import (
	"encoding/json"
	"fmt"
	"os"
)

func FileExists(path string) bool {
	_, err := os.Stat(path)
	if os.IsNotExist(err) {
		return false
	}

	return err == nil
}

func EnsureDir(path string) error {
	info, err := os.Stat(path)
	if err != nil {
		if !os.IsNotExist(err) {
			return fmt.Errorf("failed to check directory: %w", err)
		}

		if err := os.MkdirAll(path, 0755); err != nil {
			return fmt.Errorf("failed to create directory %s: %w", path, err)
		}

		return nil
	}

	if !info.IsDir() {
		return fmt.Errorf("%s already exists but is not a directory", path)
	}

	return nil
}

func ReadJSONFromFile(path string, v any) error {
	b, err := os.ReadFile(path)
	if err != nil {
		return err
	}

	return json.Unmarshal(b, v)
}

func WriteJSONToFile(path string, v any) error {
	b, err := json.MarshalIndent(v, "", "\t")
	if err != nil {
		return fmt.Errorf("json marshal error: %w", err)
	}
	tmpPath := path + ".tmp"
	if err := os.WriteFile(tmpPath, b, 0644); err != nil {
		return fmt.Errorf("write file %s error: %w", tmpPath, err)
	}
	if err := os.Rename(tmpPath, path); err != nil {
		return fmt.Errorf("rename file %s error: %w", tmpPath, err)
	}

	return nil
}
