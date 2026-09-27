package pagination_test

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"VincentLimarus/grpc-golang/internal/shared/pagination"
)

func TestNew(t *testing.T) {
	table := []struct {
		name      string
		page      int32
		limit     int32
		wantPage  int32
		wantLimit int32
	}{
		{name: "zero values fall back to defaults", page: 0, limit: 0, wantPage: 1, wantLimit: 10},
		{name: "negative values fall back to defaults", page: -3, limit: -1, wantPage: 1, wantLimit: 10},
		{name: "valid values are kept", page: 2, limit: 5, wantPage: 2, wantLimit: 5},
		{name: "limit is capped", page: 5, limit: 500, wantPage: 5, wantLimit: 100},
		{name: "limit at max is kept", page: 1, limit: 100, wantPage: 1, wantLimit: 100},
	}
	for _, tc := range table {
		t.Run(tc.name, func(t *testing.T) {
			got := pagination.New(tc.page, tc.limit)
			assert.Equal(t, tc.wantPage, got.Page)
			assert.Equal(t, tc.wantLimit, got.Limit)
		})
	}
}

func TestParams_Offset(t *testing.T) {
	table := []struct {
		name   string
		page   int32
		limit  int32
		offset int
		size   int
	}{
		{name: "first page", page: 1, limit: 10, offset: 0, size: 10},
		{name: "second page", page: 2, limit: 5, offset: 5, size: 5},
		{name: "fifth page", page: 5, limit: 100, offset: 400, size: 100},
	}
	for _, tc := range table {
		t.Run(tc.name, func(t *testing.T) {
			got := pagination.New(tc.page, tc.limit)
			assert.Equal(t, tc.offset, got.Offset())
			assert.Equal(t, tc.size, got.Size())
		})
	}
}

func TestParams_IsPaged(t *testing.T) {
	assert.True(t, pagination.New(1, 10).IsPaged())
	assert.False(t, pagination.Params{}.IsPaged())
	assert.False(t, pagination.Params{Page: 1}.IsPaged())
	assert.False(t, pagination.Params{Limit: 10}.IsPaged())
}
