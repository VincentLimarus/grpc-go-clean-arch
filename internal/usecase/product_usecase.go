package usecase

import (
	"context"

	"VincentLimarus/grpc-golang/internal/domain/entity"
	"VincentLimarus/grpc-golang/internal/domain/repository"
	"VincentLimarus/grpc-golang/internal/mapper"
	"VincentLimarus/grpc-golang/internal/model/request"
	"VincentLimarus/grpc-golang/internal/model/response"
	"VincentLimarus/grpc-golang/internal/shared/result"
)

type ProductUseCaseImpl struct {
	repo repository.ProductRepository
}

func NewProductUseCase(repo repository.ProductRepository) *ProductUseCaseImpl {
	return &ProductUseCaseImpl{repo: repo}
}

var _ ProductUseCase = (*ProductUseCaseImpl)(nil)

func (uc *ProductUseCaseImpl) GetProduct(ctx context.Context, req request.GetProductRequest) (*response.ProductResponse, error) {
	p, err := result.NotNil(uc.repo.GetByID(ctx, req.ID))
	if err != nil {
		return nil, err
	}
	return mapper.ToProductResponse(p), nil
}

func (uc *ProductUseCaseImpl) ListProducts(ctx context.Context, req request.ListProductsRequest) (*response.ProductListResponse, error) {
	params := req.Pagination()
	products, total, err := uc.repo.List(ctx, params.Page, params.Limit)
	if err != nil {
		return nil, err
	}
	return response.NewListResponse(mapper.ToProductResponses(products), total), nil
}

func (uc *ProductUseCaseImpl) CreateProduct(ctx context.Context, req request.CreateProductRequest) (*response.ProductResponse, error) {
	p, err := entity.NewProduct(req.Name, req.Price, req.Stock)
	if err != nil {
		return nil, err
	}
	if err := uc.repo.Create(ctx, p); err != nil {
		return nil, err
	}
	return mapper.ToProductResponse(p), nil
}

func (uc *ProductUseCaseImpl) UpdateProduct(ctx context.Context, req request.UpdateProductRequest) (*response.ProductResponse, error) {
	p, err := result.NotNil(uc.repo.GetByID(ctx, req.ID))
	if err != nil {
		return nil, err
	}
	if err := p.ApplyUpdate(req.Name, req.Price, req.Stock); err != nil {
		return nil, err
	}
	if err := uc.repo.Update(ctx, p); err != nil {
		return nil, err
	}
	return mapper.ToProductResponse(p), nil
}

func (uc *ProductUseCaseImpl) PatchProduct(ctx context.Context, req request.PatchProductRequest) (*response.ProductResponse, error) {
	p, err := result.NotNil(uc.repo.GetByID(ctx, req.ID))
	if err != nil {
		return nil, err
	}
	if err := p.ApplyPatch(mapper.ToProductPatch(req)); err != nil {
		return nil, err
	}
	if err := uc.repo.Update(ctx, p); err != nil {
		return nil, err
	}
	return mapper.ToProductResponse(p), nil
}

func (uc *ProductUseCaseImpl) DeleteProduct(ctx context.Context, req request.DeleteProductRequest) error {
	return uc.repo.Delete(ctx, req.ID)
}
