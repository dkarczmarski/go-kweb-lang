package store_test

import (
	"fmt"
	"os"
	"path/filepath"
	"reflect"
	"testing"

	"github.com/dkarczmarski/go-kweb-lang/store"
)

const testFilePerm = 0o600

func TestFileStore_Delete(t *testing.T) {
	t.Parallel()

	for _, tc := range []struct {
		name        string
		bucket      string
		key         string
		expectedErr func(err error) bool
		checkAfter  func(tb testing.TB, dir string) bool
	}{
		{
			name:   "delete key that exists",
			bucket: "a/b/c",
			key:    "key1",
			expectedErr: func(err error) bool {
				return err == nil
			},
			checkAfter: func(tb testing.TB, dir string) bool {
				tb.Helper()

				path := filepath.Join(dir, "a/b/c/1073ab6cda4b991cd29f9e83a307f34004ae9327.json")
				exists, err := fileExists(path)
				if err != nil {
					tb.Fatal(err)
				}

				return !exists
			},
		},
		{
			name:   "delete key that does not exist",
			bucket: "a/b/c",
			key:    "nonexistent-key",
			expectedErr: func(err error) bool {
				return err == nil
			},
		},
		{
			name:   "delete key in a bucket that does not exist",
			bucket: "x1/x2/x3",
			key:    "key",
			expectedErr: func(err error) bool {
				return err == nil
			},
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			dir := t.TempDir()

			filePath := "a/b/c/1073ab6cda4b991cd29f9e83a307f34004ae9327.json"
			fileContent := []byte("{\"Value\": \"Text\"}")

			if err := os.MkdirAll(filepath.Join(dir, filepath.Dir(filePath)), 0o700); err != nil {
				t.Fatal(err)
			}

			if err := os.WriteFile(filepath.Join(dir, filePath), fileContent, testFilePerm); err != nil {
				t.Fatal(err)
			}

			storage := store.NewFileStore(dir)

			err := storage.Delete(tc.bucket, tc.key)
			if !tc.expectedErr(err) {
				t.Errorf("unexpected error: %v", err)
			}

			if tc.checkAfter != nil && !tc.checkAfter(t, dir) {
				t.Errorf("unexpected error while checking after delete: %v", err)
			}
		})
	}
}

func fileExists(path string) (bool, error) {
	_, err := os.Stat(path)
	if os.IsNotExist(err) {
		return false, nil
	}

	if err != nil {
		return false, fmt.Errorf("failed to stat file: %w", err)
	}

	return true, nil
}

func TestFileStore_Read(t *testing.T) {
	t.Parallel()

	type testStruct struct {
		Value string
	}

	dir := t.TempDir()

	filePath := "a/b/c/1073ab6cda4b991cd29f9e83a307f34004ae9327.json"
	fileContent := []byte("{\"Value\": \"Text\"}")

	if err := os.MkdirAll(filepath.Join(dir, filepath.Dir(filePath)), 0o700); err != nil {
		t.Fatal(err)
	}

	if err := os.WriteFile(filepath.Join(dir, filePath), fileContent, testFilePerm); err != nil {
		t.Fatal(err)
	}

	storage := store.NewFileStore(dir)

	for _, tc := range []struct {
		name           string
		bucket         string
		key            string
		expectedExists bool
		expectedData   testStruct
	}{
		{
			name:           "bucket and key that exist",
			bucket:         "a/b/c",
			key:            "key1",
			expectedExists: true,
			expectedData: testStruct{
				Value: "Text",
			},
		},
		{
			name:           "nonexistent key",
			bucket:         "a/b/c",
			key:            "nonexistent-key",
			expectedExists: false,
			expectedData:   testStruct{},
		},
		{
			name:           "nonexistent bucket",
			bucket:         "x1/x2/x3",
			key:            "key",
			expectedExists: false,
			expectedData:   testStruct{},
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			var dst testStruct

			exists, err := storage.Read(tc.bucket, tc.key, &dst)
			if err != nil {
				t.Errorf("unexpected error: %v", err)
			}

			if tc.expectedExists != exists {
				t.Errorf("unexpected exists value: got %v, want %v", exists, tc.expectedExists)
			}

			if !reflect.DeepEqual(tc.expectedData, dst) {
				t.Errorf("unexpected data: got %+v, want %+v", dst, tc.expectedData)
			}
		})
	}
}

func TestFileStore_ListBuckets(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()

	mustMkDirAll(t, filepath.Join(dir, "a/b1"))
	mustMkDirAll(t, filepath.Join(dir, "a/b2"))
	mustMkDirAll(t, filepath.Join(dir, "a/b3"))

	storage := store.NewFileStore(dir)

	for _, tc := range []struct {
		name       string
		bucketPath string
		expected   []string
	}{
		{
			name:       "bucket path exists",
			bucketPath: "a",
			expected:   []string{"b1", "b2", "b3"},
		},
		{
			name:       "nonexistent bucket path",
			bucketPath: "nonexistent/1/2",
			expected:   []string{},
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			buckets, err := storage.ListBuckets(tc.bucketPath)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			if !reflect.DeepEqual(tc.expected, buckets) {
				t.Errorf("unexpected outcome: got %v, want %v", buckets, tc.expected)
			}
		})
	}
}

func mustMkDirAll(tb testing.TB, path string) {
	tb.Helper()

	if err := os.MkdirAll(path, 0o700); err != nil {
		tb.Fatal(err)
	}
}

func TestFileStore_Write(t *testing.T) {
	t.Parallel()

	type testData struct {
		Value string
	}

	for _, tc := range []struct {
		name     string
		data     any
		expected string
	}{
		{
			name:     "struct",
			data:     testData{Value: "Text"},
			expected: "{\n\t\"Value\": \"Text\"\n}",
		},
		{
			name:     "pointer to struct",
			data:     &testData{Value: "Text"},
			expected: "{\n\t\"Value\": \"Text\"\n}",
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			dir := t.TempDir()
			fileStore := store.NewFileStore(dir)

			err := fileStore.Write("a/b/c", "key1", tc.data)
			if err != nil {
				t.Fatal(err)
			}

			b, err := os.ReadFile(filepath.Join(dir, "a/b/c/1073ab6cda4b991cd29f9e83a307f34004ae9327.json"))
			if err != nil {
				t.Fatal(err)
			}

			actual := string(b)
			if actual != tc.expected {
				t.Errorf("unexpected result: got %s, want %s", actual, tc.expected)
			}
		})
	}
}
