package repository

import (
	"context"

	"VincentLimarus/grpc-golang/internal/domain/entity"
)

type ProductRepository interface {
	Create(ctx context.Context, product *entity.Product) error
	GetByID(ctx context.Context, id uint64) (*entity.Product, error)
	List(ctx context.Context, page, limit int32) (products []*entity.Product, total int64, err error)
	Update(ctx context.Context, product *entity.Product) error
	Delete(ctx context.Context, id uint64) error
}
