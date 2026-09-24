package mocks

import (
	"context"

	"github.com/stretchr/testify/mock"

	"VincentLimarus/grpc-golang/internal/domain/entity"
	"VincentLimarus/grpc-golang/internal/domain/repository"
)

type ProductRepository struct {
	mock.Mock
}

func NewProductRepository() *ProductRepository {
	return new(ProductRepository)
}

var _ repository.ProductRepository = (*ProductRepository)(nil)

func (m *ProductRepository) Create(ctx context.Context, p *entity.Product) error {
	args := m.Called(ctx, p)
	return args.Error(0)
}

func (m *ProductRepository) GetByID(ctx context.Context, id uint64) (*entity.Product, error) {
	args := m.Called(ctx, id)
	p, _ := args.Get(0).(*entity.Product)
	return p, args.Error(1)
}

func (m *ProductRepository) List(ctx context.Context, page, limit int32) ([]*entity.Product, int64, error) {
	args := m.Called(ctx, page, limit)
	products, _ := args.Get(0).([]*entity.Product)
	total, _ := args.Get(1).(int64)
	return products, total, args.Error(2)
}

func (m *ProductRepository) Update(ctx context.Context, p *entity.Product) error {
	args := m.Called(ctx, p)
	return args.Error(0)
}

func (m *ProductRepository) Delete(ctx context.Context, id uint64) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}
