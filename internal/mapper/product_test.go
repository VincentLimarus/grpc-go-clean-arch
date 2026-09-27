package mapper_test

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"VincentLimarus/grpc-golang/internal/domain/entity"
	"VincentLimarus/grpc-golang/internal/mapper"
	"VincentLimarus/grpc-golang/internal/model/request"
)

func TestToProductResponse(t *testing.T) {
	t.Run("maps every field", func(t *testing.T) {
		ts := time.Date(2026, 9, 24, 12, 0, 0, 0, time.UTC)
		got := mapper.ToProductResponse(&entity.Product{
			ID: 7, Name: "Laptop", Price: 1500.5, Stock: 4, CreatedAt: ts, UpdatedAt: ts,
		})
		require.NotNil(t, got)
		assert.Equal(t, uint64(7), got.ID)
		assert.Equal(t, "Laptop", got.Name)
		assert.Equal(t, 1500.5, got.Price)
		assert.Equal(t, int32(4), got.Stock)
		assert.Equal(t, ts, got.CreatedAt)
		assert.Equal(t, ts, got.UpdatedAt)
	})

	t.Run("nil entity", func(t *testing.T) {
		assert.Nil(t, mapper.ToProductResponse(nil))
	})
}

func TestToProductResponses(t *testing.T) {
	got := mapper.ToProductResponses([]*entity.Product{
		{ID: 1, Name: "Laptop"},
		{ID: 2, Name: "Mouse"},
	})
	require.Len(t, got, 2)
	assert.Equal(t, "Laptop", got[0].Name)
	assert.Equal(t, "Mouse", got[1].Name)

	assert.Empty(t, mapper.ToProductResponses(nil))
	assert.NotNil(t, mapper.ToProductResponses(nil))
}

func TestToProductPatch(t *testing.T) {
	name := "Gaming"
	price := 999.0
	stock := int32(8)

	patch := mapper.ToProductPatch(request.PatchProductRequest{ID: 1, Name: &name, Price: &price, Stock: &stock})
	require.NotNil(t, patch.Name)
	require.NotNil(t, patch.Price)
	require.NotNil(t, patch.Stock)
	assert.Equal(t, "Gaming", *patch.Name)
	assert.Equal(t, 999.0, *patch.Price)
	assert.Equal(t, int32(8), *patch.Stock)

	empty := mapper.ToProductPatch(request.PatchProductRequest{ID: 1})
	assert.Nil(t, empty.Name)
	assert.Nil(t, empty.Price)
	assert.Nil(t, empty.Stock)
}
