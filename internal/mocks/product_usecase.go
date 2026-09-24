package mocks

import (
	"context"

	"github.com/stretchr/testify/mock"

	"VincentLimarus/grpc-golang/internal/domain/entity"
	"VincentLimarus/grpc-golang/internal/usecase"
)

type ProductUseCase struct {
	mock.Mock
}

func NewProductUseCase() *ProductUseCase {
	return new(ProductUseCase)
}

var _ usecase.ProductUseCase = (*ProductUseCase)(nil)

func (m *ProductUseCase) GetProduct(ctx context.Context, id uint64) (*entity.Product, error) {
	args := m.Called(ctx, id)
	p, _ := args.Get(0).(*entity.Product)
	return p, args.Error(1)
}

func (m *ProductUseCase) ListProducts(ctx context.Context, page, limit int32) ([]*entity.Product, int64, error) {
	args := m.Called(ctx, page, limit)
	products, _ := args.Get(0).([]*entity.Product)
	total, _ := args.Get(1).(int64)
	return products, total, args.Error(2)
}

func (m *ProductUseCase) CreateProduct(ctx context.Context, in usecase.CreateProductInput) (*entity.Product, error) {
	args := m.Called(ctx, in)
	p, _ := args.Get(0).(*entity.Product)
	return p, args.Error(1)
}

func (m *ProductUseCase) UpdateProduct(ctx context.Context, in usecase.UpdateProductInput) (*entity.Product, error) {
	args := m.Called(ctx, in)
	p, _ := args.Get(0).(*entity.Product)
	return p, args.Error(1)
}

func (m *ProductUseCase) PatchProduct(ctx context.Context, in usecase.PatchProductInput) (*entity.Product, error) {
	args := m.Called(ctx, in)
	p, _ := args.Get(0).(*entity.Product)
	return p, args.Error(1)
}

func (m *ProductUseCase) DeleteProduct(ctx context.Context, id uint64) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}
