package entity_test

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"VincentLimarus/grpc-golang/internal/domain/entity"
)

func TestNewProduct(t *testing.T) {
	t.Run("valid", func(t *testing.T) {
		p, err := entity.NewProduct("  Laptop  ", 1500.5, 4)
		require.NoError(t, err)
		assert.Equal(t, "Laptop", p.Name)
		assert.Equal(t, 1500.5, p.Price)
		assert.Equal(t, int32(4), p.Stock)
		assert.NotZero(t, p.CreatedAt)
		assert.NotZero(t, p.UpdatedAt)
	})

	t.Run("empty name", func(t *testing.T) {
		p, err := entity.NewProduct("   ", 10, 1)
		require.ErrorIs(t, err, entity.ErrNameRequired)
		assert.Nil(t, p)
	})

	t.Run("zero price", func(t *testing.T) {
		p, err := entity.NewProduct("Laptop", 0, 1)
		require.ErrorIs(t, err, entity.ErrPriceInvalid)
		assert.Nil(t, p)
	})

	t.Run("negative stock", func(t *testing.T) {
		p, err := entity.NewProduct("Laptop", 10, -1)
		require.ErrorIs(t, err, entity.ErrStockInvalid)
		assert.Nil(t, p)
	})
}

func TestProduct_Validate(t *testing.T) {
	p := &entity.Product{Name: "Mouse", Price: 25, Stock: 10}
	require.NoError(t, p.Validate())

	table := []struct {
		name  string
		mut   func(*entity.Product)
		error error
	}{
		{name: "blank name", mut: func(p *entity.Product) { p.Name = " " }, error: entity.ErrNameRequired},
		{name: "negative price", mut: func(p *entity.Product) { p.Price = -5 }, error: entity.ErrPriceInvalid},
		{name: "negative stock", mut: func(p *entity.Product) { p.Stock = -2 }, error: entity.ErrStockInvalid},
	}
	for _, tc := range table {
		t.Run(tc.name, func(t *testing.T) {
			cp := *p
			tc.mut(&cp)
			err := cp.Validate()
			assert.True(t, errors.Is(err, tc.error))
		})
	}
}

func TestProduct_ApplyPatch(t *testing.T) {
	p, err := entity.NewProduct("Laptop", 1500, 4)
	require.NoError(t, err)

	t.Run("partial fields", func(t *testing.T) {
		name := "Gaming Laptop"
		price := 1299.0
		oldUpdated := p.UpdatedAt
		p.ApplyPatch(entity.ProductPatch{Name: &name, Price: &price})
		assert.Equal(t, name, p.Name)
		assert.Equal(t, price, p.Price)
		assert.Equal(t, int32(4), p.Stock)
		assert.True(t, p.UpdatedAt.After(oldUpdated) || !p.UpdatedAt.Equal(oldUpdated))
	})

	t.Run("no fields", func(t *testing.T) {
		err := p.ApplyPatch(entity.ProductPatch{})
		require.ErrorIs(t, err, entity.ErrNoPatchFields)
	})

	t.Run("invalid after patch", func(t *testing.T) {
		price := -1.0
		err := p.ApplyPatch(entity.ProductPatch{Price: &price})
		require.ErrorIs(t, err, entity.ErrPriceInvalid)
	})
}
