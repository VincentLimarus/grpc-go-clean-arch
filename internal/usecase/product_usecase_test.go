package usecase_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	"VincentLimarus/grpc-golang/internal/domain"
	"VincentLimarus/grpc-golang/internal/domain/entity"
	"VincentLimarus/grpc-golang/internal/mocks"
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

		in := usecase.CreateProductInput{Name: "  Laptop  ", Price: 1500, Stock: 4}
		got, err := uc.CreateProduct(context.Background(), in)
		require.NoError(t, err)
		assert.Equal(t, uint64(1), got.ID)
		assert.Equal(t, "Laptop", got.Name)
		repo.AssertExpectations(t)
	})

	t.Run("invalid name", func(t *testing.T) {
		uc, repo := newUC(t)
		in := usecase.CreateProductInput{Name: " ", Price: 1500, Stock: 4}
		got, err := uc.CreateProduct(context.Background(), in)
		require.ErrorIs(t, err, entity.ErrNameRequired)
		assert.Nil(t, got)
		repo.AssertNotCalled(t, "Create", mock.Anything, mock.Anything)
	})

	t.Run("invalid price", func(t *testing.T) {
		uc, repo := newUC(t)
		in := usecase.CreateProductInput{Name: "Laptop", Price: 0, Stock: 4}
		got, err := uc.CreateProduct(context.Background(), in)
		require.ErrorIs(t, err, entity.ErrPriceInvalid)
		assert.Nil(t, got)
		repo.AssertNotCalled(t, "Create", mock.Anything, mock.Anything)
	})

	t.Run("repo error", func(t *testing.T) {
		uc, repo := newUC(t)
		repo.On("Create", mock.Anything, mock.AnythingOfType("*entity.Product")).Return(errors.New("db down")).Once()
		in := usecase.CreateProductInput{Name: "Laptop", Price: 1500, Stock: 4}
		got, err := uc.CreateProduct(context.Background(), in)
		require.Error(t, err)
		assert.Nil(t, got)
	})
}

func TestGetProduct(t *testing.T) {
	t.Run("found", func(t *testing.T) {
		uc, repo := newUC(t)
		want := &entity.Product{ID: 3, Name: "Laptop", Price: 1500, Stock: 4}
		repo.On("GetByID", mock.Anything, uint64(3)).Return(want, nil).Once()

		got, err := uc.GetProduct(context.Background(), 3)
		require.NoError(t, err)
		assert.Same(t, want, got)
	})

	t.Run("not found", func(t *testing.T) {
		uc, repo := newUC(t)
		repo.On("GetByID", mock.Anything, uint64(9)).Return(nil, domain.ErrNotFound).Once()

		got, err := uc.GetProduct(context.Background(), 9)
		require.ErrorIs(t, err, domain.ErrNotFound)
		assert.Nil(t, got)
	})

	t.Run("nil result no error", func(t *testing.T) {
		uc, repo := newUC(t)
		repo.On("GetByID", mock.Anything, uint64(9)).Return(nil, nil).Once()

		got, err := uc.GetProduct(context.Background(), 9)
		require.ErrorIs(t, err, domain.ErrNotFound)
		assert.Nil(t, got)
	})

	t.Run("repo error", func(t *testing.T) {
		uc, repo := newUC(t)
		repo.On("GetByID", mock.Anything, uint64(1)).Return(nil, errors.New("db down")).Once()

		got, err := uc.GetProduct(context.Background(), 1)
		require.Error(t, err)
		assert.Nil(t, got)
	})
}

func TestListProducts(t *testing.T) {
	t.Run("applies defaults", func(t *testing.T) {
		uc, repo := newUC(t)
		repo.On("List", mock.Anything, int32(1), int32(10)).Return([]*entity.Product{}, int64(0), nil).Once()

		products, total, err := uc.ListProducts(context.Background(), 0, 0)
		require.NoError(t, err)
		assert.Empty(t, products)
		assert.Zero(t, total)
	})

	t.Run("caps limit at 100", func(t *testing.T) {
		uc, repo := newUC(t)
		repo.On("List", mock.Anything, int32(5), int32(100)).Return([]*entity.Product{}, int64(0), nil).Once()

		_, _, err := uc.ListProducts(context.Background(), 5, 500)
		require.NoError(t, err)
	})

	t.Run("success", func(t *testing.T) {
		uc, repo := newUC(t)
		want := []*entity.Product{
			{ID: 1, Name: "Laptop", Price: 1500, Stock: 4},
			{ID: 2, Name: "Mouse", Price: 25, Stock: 10},
		}
		repo.On("List", mock.Anything, int32(2), int32(5)).Return(want, int64(2), nil).Once()

		products, total, err := uc.ListProducts(context.Background(), 2, 5)
		require.NoError(t, err)
		assert.Len(t, products, 2)
		assert.Equal(t, int64(2), total)
	})

	t.Run("repo error", func(t *testing.T) {
		uc, repo := newUC(t)
		repo.On("List", mock.Anything, int32(1), int32(10)).Return(nil, int64(0), errors.New("db down")).Once()

		products, total, err := uc.ListProducts(context.Background(), 1, 10)
		require.Error(t, err)
		assert.Nil(t, products)
		assert.Zero(t, total)
	})
}

func TestUpdateProduct(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		uc, repo := newUC(t)
		existing := &entity.Product{ID: 1, Name: "Old", Price: 100, Stock: 1, UpdatedAt: time.Now()}
		repo.On("GetByID", mock.Anything, uint64(1)).Return(existing, nil).Once()
		repo.On("Update", mock.Anything, mock.MatchedBy(func(p *entity.Product) bool {
			return p.Name == "New Name" && p.Price == 200 && p.Stock == 5
		})).Return(nil).Once()

		in := usecase.UpdateProductInput{ID: 1, Name: "  New Name ", Price: 200, Stock: 5}
		got, err := uc.UpdateProduct(context.Background(), in)
		require.NoError(t, err)
		assert.Equal(t, "New Name", got.Name)
		assert.Equal(t, float64(200), got.Price)
		assert.Equal(t, int32(5), got.Stock)
		assert.NotZero(t, got.UpdatedAt)
	})

	t.Run("not found", func(t *testing.T) {
		uc, repo := newUC(t)
		repo.On("GetByID", mock.Anything, uint64(9)).Return(nil, domain.ErrNotFound).Once()

		in := usecase.UpdateProductInput{ID: 9, Name: "N", Price: 1, Stock: 1}
		got, err := uc.UpdateProduct(context.Background(), in)
		require.ErrorIs(t, err, domain.ErrNotFound)
		assert.Nil(t, got)
		repo.AssertNotCalled(t, "Update", mock.Anything, mock.Anything)
	})

	t.Run("invalid after update", func(t *testing.T) {
		uc, repo := newUC(t)
		existing := &entity.Product{ID: 1, Name: "Old", Price: 100, Stock: 1}
		repo.On("GetByID", mock.Anything, uint64(1)).Return(existing, nil).Once()

		in := usecase.UpdateProductInput{ID: 1, Name: "N", Price: -1, Stock: 1}
		got, err := uc.UpdateProduct(context.Background(), in)
		require.ErrorIs(t, err, entity.ErrPriceInvalid)
		assert.Nil(t, got)
		repo.AssertNotCalled(t, "Update", mock.Anything, mock.Anything)
	})

	t.Run("repo error", func(t *testing.T) {
		uc, repo := newUC(t)
		existing := &entity.Product{ID: 1, Name: "Old", Price: 100, Stock: 1}
		repo.On("GetByID", mock.Anything, uint64(1)).Return(existing, nil).Once()
		repo.On("Update", mock.Anything, mock.AnythingOfType("*entity.Product")).Return(errors.New("db down")).Once()

		in := usecase.UpdateProductInput{ID: 1, Name: "N", Price: 1, Stock: 1}
		got, err := uc.UpdateProduct(context.Background(), in)
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
		in := usecase.PatchProductInput{
			ID:    1,
			Patch: entity.ProductPatch{Name: &name, Stock: &stock},
		}
		got, err := uc.PatchProduct(context.Background(), in)
		require.NoError(t, err)
		assert.Equal(t, "Gaming", got.Name)
		assert.Equal(t, int32(8), got.Stock)
		assert.Equal(t, float64(1500), got.Price)
	})

	t.Run("not found", func(t *testing.T) {
		uc, repo := newUC(t)
		repo.On("GetByID", mock.Anything, uint64(9)).Return(nil, domain.ErrNotFound).Once()

		name := "X"
		in := usecase.PatchProductInput{ID: 9, Patch: entity.ProductPatch{Name: &name}}
		got, err := uc.PatchProduct(context.Background(), in)
		require.ErrorIs(t, err, domain.ErrNotFound)
		assert.Nil(t, got)
	})

	t.Run("no fields", func(t *testing.T) {
		uc, repo := newUC(t)
		existing := &entity.Product{ID: 1, Name: "Laptop", Price: 1500, Stock: 4}
		repo.On("GetByID", mock.Anything, uint64(1)).Return(existing, nil).Once()

		in := usecase.PatchProductInput{ID: 1}
		got, err := uc.PatchProduct(context.Background(), in)
		require.ErrorIs(t, err, entity.ErrNoPatchFields)
		assert.Nil(t, got)
	})

	t.Run("repo error", func(t *testing.T) {
		uc, repo := newUC(t)
		existing := &entity.Product{ID: 1, Name: "Laptop", Price: 1500, Stock: 4}
		repo.On("GetByID", mock.Anything, uint64(1)).Return(existing, nil).Once()
		repo.On("Update", mock.Anything, mock.AnythingOfType("*entity.Product")).Return(errors.New("db down")).Once()

		name := "Gaming"
		in := usecase.PatchProductInput{ID: 1, Patch: entity.ProductPatch{Name: &name}}
		got, err := uc.PatchProduct(context.Background(), in)
		require.Error(t, err)
		assert.Nil(t, got)
	})
}

func TestDeleteProduct(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		uc, repo := newUC(t)
		repo.On("Delete", mock.Anything, uint64(1)).Return(nil).Once()

		require.NoError(t, uc.DeleteProduct(context.Background(), 1))
	})

	t.Run("not found", func(t *testing.T) {
		uc, repo := newUC(t)
		repo.On("Delete", mock.Anything, uint64(9)).Return(domain.ErrNotFound).Once()

		require.ErrorIs(t, uc.DeleteProduct(context.Background(), 9), domain.ErrNotFound)
	})
}
