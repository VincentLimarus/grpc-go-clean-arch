package usecase

import (
	"context"

	"VincentLimarus/grpc-golang/internal/domain/entity"
)

type CreateProductInput struct {
	Name  string
	Price float64
	Stock int32
}

type UpdateProductInput struct {
	ID    uint64
	Name  string
	Price float64
	Stock int32
}

type PatchProductInput struct {
	ID    uint64
	Patch entity.ProductPatch
}

type ProductUseCase interface {
	GetProduct(ctx context.Context, id uint64) (*entity.Product, error)
	ListProducts(ctx context.Context, page, limit int32) ([]*entity.Product, int64, error)
	CreateProduct(ctx context.Context, in CreateProductInput) (*entity.Product, error)
	UpdateProduct(ctx context.Context, in UpdateProductInput) (*entity.Product, error)
	PatchProduct(ctx context.Context, in PatchProductInput) (*entity.Product, error)
	DeleteProduct(ctx context.Context, id uint64) error
}
