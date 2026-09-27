package usecase

import (
	"context"

	"VincentLimarus/grpc-golang/internal/model/request"
	"VincentLimarus/grpc-golang/internal/model/response"
)

type ProductUseCase interface {
	GetProduct(ctx context.Context, req request.GetProductRequest) (*response.ProductResponse, error)
	ListProducts(ctx context.Context, req request.ListProductsRequest) (*response.ProductListResponse, error)
	CreateProduct(ctx context.Context, req request.CreateProductRequest) (*response.ProductResponse, error)
	UpdateProduct(ctx context.Context, req request.UpdateProductRequest) (*response.ProductResponse, error)
	PatchProduct(ctx context.Context, req request.PatchProductRequest) (*response.ProductResponse, error)
	DeleteProduct(ctx context.Context, req request.DeleteProductRequest) error
}
