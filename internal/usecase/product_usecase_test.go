package usecase_test

import (
	"context"
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	"VincentLimarus/grpc-golang/internal/domain"
	"VincentLimarus/grpc-golang/internal/domain/entity"
	"VincentLimarus/grpc-golang/internal/mapper"
	"VincentLimarus/grpc-golang/internal/mocks"
	"VincentLimarus/grpc-golang/internal/model/request"
	"VincentLimarus/grpc-golang/internal/usecase"
)

func newUC(t *testing.T) (*usecase.ProductUseCaseImpl, *mocks.ProductRepository) {
	t.Helper()
	repo := mocks.NewProductRepository()
	return usecase.NewProductUseCase(repo), repo
}

func TestCreateProduct(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		uc, repo := newUC(t)
		repo.On("Create", mock.Anything, mock.MatchedBy(func(p *entity.Product) bool {
			return p.Name == "Laptop" && p.Price == 1500 && p.Stock == 4
		})).Run(func(args mock.Arguments) {
			args.Get(1).(*entity.Product).ID = 1
		}).Return(nil).Once()

		req := request.CreateProductRequest{Name: "  Laptop  ", Price: 1500, Stock: 4}
		got, err := uc.CreateProduct(context.Background(), req)
		require.NoError(t, err)
		assert.Equal(t, uint64(1), got.ID)
		assert.Equal(t, "Laptop", got.Name)
		repo.AssertExpectations(t)
	})

	t.Run("invalid name", func(t *testing.T) {
		uc, repo := newUC(t)
		req := request.CreateProductRequest{Name: " ", Price: 1500, Stock: 4}
		got, err := uc.CreateProduct(context.Background(), req)
		require.ErrorIs(t, err, entity.ErrNameRequired)
		assert.Nil(t, got)
		repo.AssertNotCalled(t, "Create", mock.Anything, mock.Anything)
	})

	t.Run("invalid price", func(t *testing.T) {
		uc, repo := newUC(t)
		req := request.CreateProductRequest{Name: "Laptop", Price: 0, Stock: 4}
		got, err := uc.CreateProduct(context.Background(), req)
		require.ErrorIs(t, err, entity.ErrPriceInvalid)
		assert.Nil(t, got)
		repo.AssertNotCalled(t, "Create", mock.Anything, mock.Anything)
	})

	t.Run("repo error", func(t *testing.T) {
		uc, repo := newUC(t)
		repo.On("Create", mock.Anything, mock.AnythingOfType("*entity.Product")).Return(errors.New("db down")).Once()
		req := request.CreateProductRequest{Name: "Laptop", Price: 1500, Stock: 4}
		got, err := uc.CreateProduct(context.Background(), req)
		require.Error(t, err)
		assert.Nil(t, got)
	})
}

func TestGetProduct(t *testing.T) {
	t.Run("found", func(t *testing.T) {
		uc, repo := newUC(t)
		want := &entity.Product{ID: 3, Name: "Laptop", Price: 1500, Stock: 4}
		repo.On("GetByID", mock.Anything, uint64(3)).Return(want, nil).Once()

		got, err := uc.GetProduct(context.Background(), request.GetProductRequest{ID: 3})
		require.NoError(t, err)
		assert.Equal(t, mapper.ToProductResponse(want), got)
	})

	t.Run("not found", func(t *testing.T) {
		uc, repo := newUC(t)
		repo.On("GetByID", mock.Anything, uint64(9)).Return(nil, domain.ErrNotFound).Once()

		got, err := uc.GetProduct(context.Background(), request.GetProductRequest{ID: 9})
		require.ErrorIs(t, err, domain.ErrNotFound)
		assert.Nil(t, got)
	})

	t.Run("nil result no error", func(t *testing.T) {
		uc, repo := newUC(t)
		repo.On("GetByID", mock.Anything, uint64(9)).Return(nil, nil).Once()

		got, err := uc.GetProduct(context.Background(), request.GetProductRequest{ID: 9})
		require.ErrorIs(t, err, domain.ErrNotFound)
		assert.Nil(t, got)
	})

	t.Run("repo error", func(t *testing.T) {
		uc, repo := newUC(t)
		repo.On("GetByID", mock.Anything, uint64(1)).Return(nil, errors.New("db down")).Once()

		got, err := uc.GetProduct(context.Background(), request.GetProductRequest{ID: 1})
		require.Error(t, err)
		assert.Nil(t, got)
	})
}

func TestListProducts(t *testing.T) {
	t.Run("applies defaults", func(t *testing.T) {
		uc, repo := newUC(t)
		repo.On("List", mock.Anything, int32(1), int32(10)).Return([]*entity.Product{}, int64(0), nil).Once()

		got, err := uc.ListProducts(context.Background(), request.ListProductsRequest{})
		require.NoError(t, err)
		assert.Empty(t, got.Items)
		assert.Zero(t, got.Total)
	})

	t.Run("caps limit at 100", func(t *testing.T) {
		uc, repo := newUC(t)
		repo.On("List", mock.Anything, int32(5), int32(100)).Return([]*entity.Product{}, int64(0), nil).Once()

		got, err := uc.ListProducts(context.Background(), request.ListProductsRequest{Page: 5, Limit: 500})
		require.NoError(t, err)
		assert.Empty(t, got.Items)
	})

	t.Run("success", func(t *testing.T) {
		uc, repo := newUC(t)
		want := []*entity.Product{
			{ID: 1, Name: "Laptop", Price: 1500, Stock: 4},
			{ID: 2, Name: "Mouse", Price: 25, Stock: 10},
		}
		repo.On("List", mock.Anything, int32(2), int32(5)).Return(want, int64(2), nil).Once()

		got, err := uc.ListProducts(context.Background(), request.ListProductsRequest{Page: 2, Limit: 5})
		require.NoError(t, err)
		assert.Len(t, got.Items, 2)
		assert.Equal(t, int64(2), got.Total)
		assert.Equal(t, "Mouse", got.Items[1].Name)
	})

	t.Run("repo error", func(t *testing.T) {
		uc, repo := newUC(t)
		repo.On("List", mock.Anything, int32(1), int32(10)).Return(nil, int64(0), errors.New("db down")).Once()

		got, err := uc.ListProducts(context.Background(), request.ListProductsRequest{Page: 1, Limit: 10})
		require.Error(t, err)
		assert.Nil(t, got)
	})
}

func TestUpdateProduct(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		uc, repo := newUC(t)
		existing := &entity.Product{ID: 1, Name: "Old", Price: 100, Stock: 1}
		repo.On("GetByID", mock.Anything, uint64(1)).Return(existing, nil).Once()
		repo.On("Update", mock.Anything, mock.MatchedBy(func(p *entity.Product) bool {
			return p.Name == "New Name" && p.Price == 200 && p.Stock == 5
		})).Return(nil).Once()

		req := request.UpdateProductRequest{ID: 1, Name: "  New Name ", Price: 200, Stock: 5}
		got, err := uc.UpdateProduct(context.Background(), req)
		require.NoError(t, err)
		assert.Equal(t, "New Name", got.Name)
		assert.Equal(t, float64(200), got.Price)
		assert.Equal(t, int32(5), got.Stock)
		assert.False(t, got.UpdatedAt.IsZero())
	})

	t.Run("not found", func(t *testing.T) {
		uc, repo := newUC(t)
		repo.On("GetByID", mock.Anything, uint64(9)).Return(nil, domain.ErrNotFound).Once()

		req := request.UpdateProductRequest{ID: 9, Name: "N", Price: 1, Stock: 1}
		got, err := uc.UpdateProduct(context.Background(), req)
		require.ErrorIs(t, err, domain.ErrNotFound)
		assert.Nil(t, got)
		repo.AssertNotCalled(t, "Update", mock.Anything, mock.Anything)
	})

	t.Run("invalid after update", func(t *testing.T) {
		uc, repo := newUC(t)
		existing := &entity.Product{ID: 1, Name: "Old", Price: 100, Stock: 1}
		repo.On("GetByID", mock.Anything, uint64(1)).Return(existing, nil).Once()

		req := request.UpdateProductRequest{ID: 1, Name: "N", Price: -1, Stock: 1}
		got, err := uc.UpdateProduct(context.Background(), req)
		require.ErrorIs(t, err, entity.ErrPriceInvalid)
		assert.Nil(t, got)
		repo.AssertNotCalled(t, "Update", mock.Anything, mock.Anything)
	})

	t.Run("repo error", func(t *testing.T) {
		uc, repo := newUC(t)
		existing := &entity.Product{ID: 1, Name: "Old", Price: 100, Stock: 1}
		repo.On("GetByID", mock.Anything, uint64(1)).Return(existing, nil).Once()
		repo.On("Update", mock.Anything, mock.AnythingOfType("*entity.Product")).Return(errors.New("db down")).Once()

		req := request.UpdateProductRequest{ID: 1, Name: "N", Price: 1, Stock: 1}
		got, err := uc.UpdateProduct(context.Background(), req)
		require.Error(t, err)
		assert.Nil(t, got)
	})
}

func TestPatchProduct(t *testing.T) {
	t.Run("success partial patch", func(t *testing.T) {
		uc, repo := newUC(t)
		existing := &entity.Product{ID: 1, Name: "Laptop", Price: 1500, Stock: 4}
		repo.On("GetByID", mock.Anything, uint64(1)).Return(existing, nil).Once()
		repo.On("Update", mock.Anything, mock.MatchedBy(func(p *entity.Product) bool {
			return p.Name == "Gaming" && p.Price == 1500 && p.Stock == 8
		})).Return(nil).Once()

		name := "Gaming"
		stock := int32(8)
		req := request.PatchProductRequest{ID: 1, Name: &name, Stock: &stock}
		got, err := uc.PatchProduct(context.Background(), req)
		require.NoError(t, err)
		assert.Equal(t, "Gaming", got.Name)
		assert.Equal(t, int32(8), got.Stock)
		assert.Equal(t, float64(1500), got.Price)
	})

	t.Run("not found", func(t *testing.T) {
		uc, repo := newUC(t)
		repo.On("GetByID", mock.Anything, uint64(9)).Return(nil, domain.ErrNotFound).Once()

		name := "X"
		got, err := uc.PatchProduct(context.Background(), request.PatchProductRequest{ID: 9, Name: &name})
		require.ErrorIs(t, err, domain.ErrNotFound)
		assert.Nil(t, got)
	})

	t.Run("no fields", func(t *testing.T) {
		uc, repo := newUC(t)
		existing := &entity.Product{ID: 1, Name: "Laptop", Price: 1500, Stock: 4}
		repo.On("GetByID", mock.Anything, uint64(1)).Return(existing, nil).Once()

		got, err := uc.PatchProduct(context.Background(), request.PatchProductRequest{ID: 1})
		require.ErrorIs(t, err, entity.ErrNoPatchFields)
		assert.Nil(t, got)
	})

	t.Run("repo error", func(t *testing.T) {
		uc, repo := newUC(t)
		existing := &entity.Product{ID: 1, Name: "Laptop", Price: 1500, Stock: 4}
		repo.On("GetByID", mock.Anything, uint64(1)).Return(existing, nil).Once()
		repo.On("Update", mock.Anything, mock.AnythingOfType("*entity.Product")).Return(errors.New("db down")).Once()

		name := "Gaming"
		got, err := uc.PatchProduct(context.Background(), request.PatchProductRequest{ID: 1, Name: &name})
		require.Error(t, err)
		assert.Nil(t, got)
	})
}

func TestDeleteProduct(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		uc, repo := newUC(t)
		repo.On("Delete", mock.Anything, uint64(1)).Return(nil).Once()

		require.NoError(t, uc.DeleteProduct(context.Background(), request.DeleteProductRequest{ID: 1}))
	})

	t.Run("not found", func(t *testing.T) {
		uc, repo := newUC(t)
		repo.On("Delete", mock.Anything, uint64(9)).Return(domain.ErrNotFound).Once()

		err := uc.DeleteProduct(context.Background(), request.DeleteProductRequest{ID: 9})
		require.ErrorIs(t, err, domain.ErrNotFound)
	})
}
