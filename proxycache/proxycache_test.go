package proxycache_test

import (
	"context"
	"errors"
	"log"
	"testing"

	"go-kweb-lang/proxycache"
)

func TestGet(t *testing.T) {
	for _, tc := range []struct {
		name        string
		before      func(t *testing.T, cacheDir, category, key string)
		isInvalid   func(value string) bool
		block       func(ctx context.Context) (string, error)
		checkResult func(t *testing.T, value string, err error) bool
		after       func(t *testing.T, cacheDir, category, key string)
	}{
		{
			name: "first call should execute block and returns value",
			before: func(t *testing.T, cacheDir, category, key string) {
				t.Helper()

				if proxycacheKeyExists(t, cacheDir, category, key) {
					t.Fatal()
				}
			},
			isInvalid: nil,
			block: func(ctx context.Context) (string, error) {
				return "my-value", nil
			},
			checkResult: func(t *testing.T, value string, err error) bool {
				return value == "my-value" && err == nil
			},
			after: func(t *testing.T, cacheDir, category, key string) {
				t.Helper()

				if !proxycacheKeyExists(t, cacheDir, category, key) {
					t.Error("key should exist")
				}
			},
		},
		{
			name: "first call should execute block and returns error",
			before: func(t *testing.T, cacheDir, category, key string) {
				t.Helper()

				if proxycacheKeyExists(t, cacheDir, category, key) {
					t.Fatal()
				}
			},
			isInvalid: nil,
			block: func(ctx context.Context) (string, error) {
				return "", errors.New("my-error")
			},
			checkResult: func(t *testing.T, value string, err error) bool {
				return value == "" && err != nil
			},
			after: func(t *testing.T, cacheDir, category, key string) {
				t.Helper()

				if proxycacheKeyExists(t, cacheDir, category, key) {
					t.Error("key should not exist")
				}
			},
		},
		{
			name: "second call when no validation, should not execute block and hit cache",
			before: func(t *testing.T, cacheDir, category, key string) {
				t.Helper()

				if _, err := proxycache.Get(
					context.Background(),
					cacheDir,
					category,
					key,
					nil,
					func(ctx context.Context) (string, error) {
						return "my-value", nil
					},
				); err != nil {
					t.Fatal(err)
				}

				if !proxycacheKeyExists(t, cacheDir, category, key) {
					t.Fatal()
				}
			},
			isInvalid: nil,
			block: func(ctx context.Context) (string, error) {
				log.Fatal("it should not be run")

				return "", nil
			},
			checkResult: func(t *testing.T, value string, err error) bool {
				return value == "my-value" && err == nil
			},
			after: func(t *testing.T, cacheDir, category, key string) {
				t.Helper()

				if !proxycacheKeyExists(t, cacheDir, category, key) {
					t.Error("key should exist")
				}
			},
		},
		{
			name: "second call when is valid, should not execute block and hit cache",
			before: func(t *testing.T, cacheDir, category, key string) {
				t.Helper()

				if _, err := proxycache.Get(
					context.Background(),
					cacheDir,
					category,
					key,
					nil,
					func(ctx context.Context) (string, error) {
						return "my-value", nil
					},
				); err != nil {
					t.Fatal(err)
				}

				if !proxycacheKeyExists(t, cacheDir, category, key) {
					t.Fatal()
				}
			},
			isInvalid: func(value string) bool {
				return false
			},
			block: func(ctx context.Context) (string, error) {
				log.Fatal("it should not be run")

				return "", nil
			},
			checkResult: func(t *testing.T, value string, err error) bool {
				return value == "my-value" && err == nil
			},
			after: func(t *testing.T, cacheDir, category, key string) {
				t.Helper()

				if !proxycacheKeyExists(t, cacheDir, category, key) {
					t.Error("key should exist")
				}
			},
		},
		{
			name: "second call when is not valid, should execute block",
			before: func(t *testing.T, cacheDir, category, key string) {
				t.Helper()

				if _, err := proxycache.Get(
					context.Background(),
					cacheDir,
					category,
					key,
					nil,
					func(ctx context.Context) (string, error) {
						return "my-value", nil
					},
				); err != nil {
					t.Fatal(err)
				}

				if !proxycacheKeyExists(t, cacheDir, category, key) {
					t.Fatal()
				}
			},
			isInvalid: func(value string) bool {
				return true
			},
			block: func(ctx context.Context) (string, error) {
				return "my-value2", nil
			},
			checkResult: func(t *testing.T, value string, err error) bool {
				return value == "my-value2" && err == nil
			},
			after: func(t *testing.T, cacheDir, category, key string) {
				t.Helper()

				if !proxycacheKeyExists(t, cacheDir, category, key) {
					t.Error("key should exist")
				}
			},
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			cacheDir := t.TempDir()
			category := "my-category"
			key := "my-key"

			tc.before(t, cacheDir, category, key)

			value, err := proxycache.Get(
				context.Background(),
				cacheDir,
				category,
				key,
				tc.isInvalid,
				tc.block,
			)

			if !tc.checkResult(t, value, err) {
				t.Errorf("unexpected result: %v, %v", value, err)
			}

			tc.after(t, cacheDir, category, key)
		})
	}
}

func TestPut(t *testing.T) {
	for _, tc := range []struct {
		name   string
		before func(t *testing.T, cacheDir, category, key string)
	}{
		{
			name: "put value that not exists",
			before: func(t *testing.T, cacheDir, category, key string) {
				t.Helper()

				if proxycacheKeyExists(t, cacheDir, category, key) {
					t.Fatal("key should not exist")
				}
			},
		},
		{
			name: "put override value that exists",
			before: func(t *testing.T, cacheDir, category, key string) {
				t.Helper()

				if err := proxycache.Put(cacheDir, category, key, ""); err != nil {
					t.Fatal(err)
				}

				if !proxycacheKeyExists(t, cacheDir, category, key) {
					t.Fatal("key should exist")
				}
			},
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			cacheDir := t.TempDir()
			category := "my-category"
			key := "my-key"
			value := "my-value"

			tc.before(t, cacheDir, category, key)

			err := proxycache.Put(cacheDir, category, key, value)
			if err != nil {
				t.Fatal(err)
			}

			if !proxycacheKeyExists(t, cacheDir, category, key) {
				t.Error("key should exist")
			}

			valueFromCache, err := proxycache.Get(
				context.Background(),
				cacheDir,
				category,
				key,
				nil,
				func(ctx context.Context) (string, error) {
					log.Fatal("should not call it")
					return "", nil
				},
			)
			if err != nil {
				t.Fatal(err)
			}

			if value != valueFromCache {
				t.Errorf("unexpected value: %v", valueFromCache)
			}
		})
	}
}

func TestInvalidateKey(t *testing.T) {
	for _, tc := range []struct {
		name   string
		before func(t *testing.T, cacheDir, category, key string)
	}{
		{
			name: "invalidate non-existent key",
		},
		{
			name: "invalidate existing key",
			before: func(t *testing.T, cacheDir, category, key string) {
				t.Helper()

				if err := proxycache.Put(cacheDir, category, key, ""); err != nil {
					t.Fatal(err)
				}
			},
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			cacheDir := t.TempDir()
			category := "my-category"
			key := "my-key"

			if err := proxycache.InvalidateKey(
				cacheDir,
				category,
				key,
			); err != nil {
				t.Fatal(err)
			}

			if proxycacheKeyExists(t, cacheDir, category, key) {
				t.Error("invalidated key should not exists")
			}
		})
	}
}

func proxycacheKeyExists(t *testing.T, cacheDir, category, key string) bool {
	t.Helper()

	exists, err := proxycache.KeyExists(cacheDir, category, key)
	if err != nil {
		t.Fatal(err)
	}

	return exists
}
