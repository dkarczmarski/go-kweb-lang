package proxycache_test

import (
	"context"
	"errors"
	"reflect"
	"testing"

	"github.com/dkarczmarski/go-kweb-lang/proxycache"
	"github.com/dkarczmarski/go-kweb-lang/proxycache/internal/mocks"
	"github.com/dkarczmarski/go-kweb-lang/testing/storetests"
	"go.uber.org/mock/gomock"
)

type TestData struct {
	Value string
}

var errTest = errors.New("test error")

func noError(err error) bool {
	return err == nil
}

func TestGet(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name       string
		initMock   func(storeMock *mocks.MockStore)
		isInvalid  func(data TestData) bool
		block      func(ctx context.Context) (TestData, error)
		wantResult TestData
		wantErr    func(err error) bool
	}{
		{
			name: "value found in cache",
			initMock: func(storeMock *mocks.MockStore) {
				storeMock.EXPECT().
					Read("bucket", "key", gomock.Any()).
					DoAndReturn(storetests.MockReadReturn(
						true,
						TestData{Value: "value1"},
						nil,
					))
			},
			isInvalid: nil,
			block: func(_ context.Context) (TestData, error) {
				return TestData{Value: "value1"}, nil
			},
			wantResult: TestData{Value: "value1"},
			wantErr:    noError,
		},
		{
			name: "error while reading from store",
			initMock: func(storeMock *mocks.MockStore) {
				storeMock.EXPECT().
					Read("bucket", "key", gomock.Any()).
					DoAndReturn(storetests.MockReadReturn(
						true,
						TestData{},
						errTest,
					))
			},
			isInvalid: nil,
			block: func(_ context.Context) (TestData, error) {
				return TestData{Value: "value1"}, nil
			},
			wantResult: TestData{},
			wantErr: func(err error) bool {
				return errors.Is(err, errTest)
			},
		},
		{
			name: "value not found in cache",
			initMock: func(storeMock *mocks.MockStore) {
				storeMock.EXPECT().
					Read("bucket", "key", gomock.Any()).
					DoAndReturn(storetests.MockReadReturn(
						false,
						TestData{},
						nil,
					))

				storeMock.EXPECT().
					Write("bucket", "key", TestData{Value: "value1"}).
					Return(nil)
			},
			isInvalid: nil,
			block: func(_ context.Context) (TestData, error) {
				return TestData{Value: "value1"}, nil
			},
			wantResult: TestData{Value: "value1"},
			wantErr:    noError,
		},
		{
			name: "error while running block",
			initMock: func(storeMock *mocks.MockStore) {
				storeMock.EXPECT().
					Read("bucket", "key", gomock.Any()).
					DoAndReturn(storetests.MockReadReturn(
						false,
						TestData{},
						nil,
					))
			},
			isInvalid: nil,
			block: func(_ context.Context) (TestData, error) {
				return TestData{}, errTest
			},
			wantResult: TestData{},
			wantErr: func(err error) bool {
				return errors.Is(err, errTest)
			},
		},
		{
			name: "error while writing to store",
			initMock: func(storeMock *mocks.MockStore) {
				storeMock.EXPECT().
					Read("bucket", "key", gomock.Any()).
					DoAndReturn(storetests.MockReadReturn(
						false,
						TestData{},
						nil,
					))

				storeMock.EXPECT().
					Write("bucket", "key", TestData{Value: "value1"}).
					Return(errTest)
			},
			isInvalid: nil,
			block: func(_ context.Context) (TestData, error) {
				return TestData{Value: "value1"}, nil
			},
			wantResult: TestData{},
			wantErr: func(err error) bool {
				return errors.Is(err, errTest)
			},
		},
		{
			name: "cached value is invalid and gets refreshed",
			initMock: func(storeMock *mocks.MockStore) {
				storeMock.EXPECT().
					Read("bucket", "key", gomock.Any()).
					DoAndReturn(storetests.MockReadReturn(
						true,
						TestData{Value: "value1"},
						nil,
					))

				storeMock.EXPECT().
					Write("bucket", "key", TestData{Value: "value2"}).
					Return(nil)
			},
			isInvalid: func(_ TestData) bool {
				return true
			},
			block: func(_ context.Context) (TestData, error) {
				return TestData{Value: "value2"}, nil
			},
			wantResult: TestData{Value: "value2"},
			wantErr:    noError,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			ctx := t.Context()

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

			if !tc.wantErr(err) {
				t.Fatalf("unexpected error: %v", err)
			}

			if !reflect.DeepEqual(tc.wantResult, result) {
				t.Fatalf(
					"unexpected result\nwant: %+v\ngot:  %+v",
					tc.wantResult,
					result,
				)
			}
		})
	}
}
