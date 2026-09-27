package mocks

import (
	"context"

	"github.com/stretchr/testify/mock"

	"VincentLimarus/grpc-golang/internal/model/request"
	"VincentLimarus/grpc-golang/internal/model/response"
	"VincentLimarus/grpc-golang/internal/usecase"
)

type ProductUseCase struct {
	mock.Mock
}

func NewProductUseCase() *ProductUseCase {
	return new(ProductUseCase)
}

var _ usecase.ProductUseCase = (*ProductUseCase)(nil)

func (m *ProductUseCase) GetProduct(ctx context.Context, req request.GetProductRequest) (*response.ProductResponse, error) {
	args := m.Called(ctx, req)
	res, _ := args.Get(0).(*response.ProductResponse)
	return res, args.Error(1)
}

func (m *ProductUseCase) ListProducts(ctx context.Context, req request.ListProductsRequest) (*response.ProductListResponse, error) {
	args := m.Called(ctx, req)
	res, _ := args.Get(0).(*response.ProductListResponse)
	return res, args.Error(1)
}

func (m *ProductUseCase) CreateProduct(ctx context.Context, req request.CreateProductRequest) (*response.ProductResponse, error) {
	args := m.Called(ctx, req)
	res, _ := args.Get(0).(*response.ProductResponse)
	return res, args.Error(1)
}

func (m *ProductUseCase) UpdateProduct(ctx context.Context, req request.UpdateProductRequest) (*response.ProductResponse, error) {
	args := m.Called(ctx, req)
	res, _ := args.Get(0).(*response.ProductResponse)
	return res, args.Error(1)
}

func (m *ProductUseCase) PatchProduct(ctx context.Context, req request.PatchProductRequest) (*response.ProductResponse, error) {
	args := m.Called(ctx, req)
	res, _ := args.Get(0).(*response.ProductResponse)
	return res, args.Error(1)
}

func (m *ProductUseCase) DeleteProduct(ctx context.Context, req request.DeleteProductRequest) error {
	args := m.Called(ctx, req)
	return args.Error(0)
}
