package proxycache_test

import (
	"context"
	"errors"
	"reflect"
	"testing"

	"go-kweb-lang/proxycache"
	"go-kweb-lang/proxycache/internal/mocks"
	"go-kweb-lang/testing/storetests"

	"go.uber.org/mock/gomock"
)

type TestData struct {
	Value string
}

var errTest = errors.New("test error")

func TestGet(t *testing.T) {
	for _, tc := range []struct {
		name         string
		initMock     func(storeMock *mocks.MockStore)
		isInvalid    func(data TestData) bool
		block        func(ctx context.Context) (TestData, error)
		expectResult TestData
		expectErr    func(err error) bool
	}{
		{
			name: "value found in the cache",
			initMock: func(storeMock *mocks.MockStore) {
				storeMock.EXPECT().
					Read("bucket", "key", gomock.Any()).
					DoAndReturn(storetests.MockReadReturn(
						true,
						TestData{"value1"},
						nil),
					)
			},
			isInvalid: nil,
			block: func(ctx context.Context) (TestData, error) {
				return TestData{"value1"}, nil
			},
			expectResult: TestData{"value1"},
			expectErr: func(err error) bool {
				return err == nil
			},
		},
		{
			name: "error while reading from the store",
			initMock: func(storeMock *mocks.MockStore) {
				storeMock.EXPECT().
					Read("bucket", "key", gomock.Any()).
					DoAndReturn(storetests.MockReadReturn(
						true,
						TestData{},
						errTest),
					)
			},
			isInvalid: nil,
			block: func(ctx context.Context) (TestData, error) {
				return TestData{"value1"}, nil
			},
			expectResult: TestData{},
			expectErr: func(err error) bool {
				return errors.Is(err, errTest)
			},
		},
		{
			name: "value not found in the cache",
			initMock: func(storeMock *mocks.MockStore) {
				storeMock.EXPECT().
					Read("bucket", "key", gomock.Any()).
					DoAndReturn(storetests.MockReadReturn(
						false,
						TestData{},
						nil),
					)

				storeMock.EXPECT().
					Write("bucket", "key", TestData{"value1"}).Return(nil)
			},
			isInvalid: nil,
			block: func(ctx context.Context) (TestData, error) {
				return TestData{"value1"}, nil
			},
			expectResult: TestData{"value1"},
			expectErr: func(err error) bool {
				return err == nil
			},
		},
		{
			name: "error while running block",
			initMock: func(storeMock *mocks.MockStore) {
				storeMock.EXPECT().
					Read("bucket", "key", gomock.Any()).
					DoAndReturn(storetests.MockReadReturn(
						false,
						TestData{},
						nil),
					)
			},
			isInvalid: nil,
			block: func(ctx context.Context) (TestData, error) {
				return TestData{}, errTest
			},
			expectResult: TestData{},
			expectErr: func(err error) bool {
				return errors.Is(err, errTest)
			},
		},
		{
			name: "error while writing to the store",
			initMock: func(storeMock *mocks.MockStore) {
				storeMock.EXPECT().
					Read("bucket", "key", gomock.Any()).
					DoAndReturn(storetests.MockReadReturn(
						false,
						TestData{""},
						nil),
					)

				storeMock.EXPECT().
					Write("bucket", "key", TestData{"value1"}).Return(errTest)
			},
			isInvalid: nil,
			block: func(ctx context.Context) (TestData, error) {
				return TestData{"value1"}, nil
			},
			expectResult: TestData{},
			expectErr: func(err error) bool {
				return errors.Is(err, errTest)
			},
		},
		{
			name: "value found in the cache but is not valid",
			initMock: func(storeMock *mocks.MockStore) {
				storeMock.EXPECT().
					Read("bucket", "key", gomock.Any()).
					DoAndReturn(storetests.MockReadReturn(
						true,
						TestData{"value1"},
						nil),
					)

				storeMock.EXPECT().
					Write("bucket", "key", TestData{"value2"}).Return(nil)
			},
			isInvalid: func(data TestData) bool {
				return true
			},
			block: func(ctx context.Context) (TestData, error) {
				return TestData{"value2"}, nil
			},
			expectResult: TestData{"value2"},
			expectErr: func(err error) bool {
				return err == nil
			},
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			ctx := context.Background()

			ctrl := gomock.NewController(t)
			storeMock := mocks.NewMockStore(ctrl)

			tc.initMock(storeMock)

			result, err := proxycache.Get(
				ctx,
				storeMock,
				"bucket",
				"key",
				tc.isInvalid,
				tc.block,
			)

			if !tc.expectErr(err) {
				t.Errorf("unexpected error: %v", err)
			}

			if !reflect.DeepEqual(tc.expectResult, result) {
				t.Errorf("unexpected result\nexpected: %+v\nactual:  %+v",
					tc.expectResult, result)
			}
		})
	}
}
