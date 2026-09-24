package usecase

import (
	"context"
	"strings"
	"time"

	"VincentLimarus/grpc-golang/internal/domain"
	"VincentLimarus/grpc-golang/internal/domain/entity"
	"VincentLimarus/grpc-golang/internal/domain/repository"
)

const (
	defaultPage  = 1
	defaultLimit = 10
	maxLimit     = 100
)

type ProductUseCaseImpl struct {
	repo repository.ProductRepository
}

func NewProductUseCase(repo repository.ProductRepository) *ProductUseCaseImpl {
	return &ProductUseCaseImpl{repo: repo}
}

var _ ProductUseCase = (*ProductUseCaseImpl)(nil)

func (uc *ProductUseCaseImpl) GetProduct(ctx context.Context, id uint64) (*entity.Product, error) {
	p, err := uc.repo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}
	if p == nil {
		return nil, domain.ErrNotFound
	}
	return p, nil
}

func (uc *ProductUseCaseImpl) ListProducts(ctx context.Context, page, limit int32) ([]*entity.Product, int64, error) {
	if page < 1 {
		page = defaultPage
	}
	if limit < 1 {
		limit = defaultLimit
	}
	if limit > maxLimit {
		limit = maxLimit
	}
	return uc.repo.List(ctx, page, limit)
}

func (uc *ProductUseCaseImpl) CreateProduct(ctx context.Context, in CreateProductInput) (*entity.Product, error) {
	p, err := entity.NewProduct(in.Name, in.Price, in.Stock)
	if err != nil {
		return nil, err
	}
	if err := uc.repo.Create(ctx, p); err != nil {
		return nil, err
	}
	return p, nil
}

func (uc *ProductUseCaseImpl) UpdateProduct(ctx context.Context, in UpdateProductInput) (*entity.Product, error) {
	p, err := uc.GetProduct(ctx, in.ID)
	if err != nil {
		return nil, err
	}
	p.Name = strings.TrimSpace(in.Name)
	p.Price = in.Price
	p.Stock = in.Stock
	p.UpdatedAt = time.Now().UTC()
	if err := p.Validate(); err != nil {
		return nil, err
	}
	if err := uc.repo.Update(ctx, p); err != nil {
		return nil, err
	}
	return p, nil
}

func (uc *ProductUseCaseImpl) PatchProduct(ctx context.Context, in PatchProductInput) (*entity.Product, error) {
	p, err := uc.GetProduct(ctx, in.ID)
	if err != nil {
		return nil, err
	}
	if err := p.ApplyPatch(in.Patch); err != nil {
		return nil, err
	}
	if err := uc.repo.Update(ctx, p); err != nil {
		return nil, err
	}
	return p, nil
}

func (uc *ProductUseCaseImpl) DeleteProduct(ctx context.Context, id uint64) error {
	return uc.repo.Delete(ctx, id)
}
