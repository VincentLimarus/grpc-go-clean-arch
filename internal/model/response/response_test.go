package response_test

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"VincentLimarus/grpc-golang/internal/model/response"
)

func TestNewListResponse(t *testing.T) {
	t.Run("keeps items and total", func(t *testing.T) {
		items := []*response.ProductResponse{{ID: 1}, {ID: 2}}
		got := response.NewListResponse(items, 2)
		assert.Equal(t, items, got.Items)
		assert.Equal(t, int64(2), got.Total)
	})

	t.Run("nil items become empty slice", func(t *testing.T) {
		got := response.NewListResponse[*response.ProductResponse](nil, 0)
		assert.NotNil(t, got.Items)
		assert.Empty(t, got.Items)
		assert.Zero(t, got.Total)
	})
}
