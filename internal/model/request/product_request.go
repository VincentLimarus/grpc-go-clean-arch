package request

import "VincentLimarus/grpc-golang/internal/shared/pagination"

type GetProductRequest struct {
	ID uint64
}

type ListProductsRequest struct {
	Page  int32
	Limit int32
}

func (r ListProductsRequest) Pagination() pagination.Params {
	return pagination.New(r.Page, r.Limit)
}

type CreateProductRequest struct {
	Name  string
	Price float64
	Stock int32
}

type UpdateProductRequest struct {
	ID    uint64
	Name  string
	Price float64
	Stock int32
}

type PatchProductRequest struct {
	ID    uint64
	Name  *string
	Price *float64
	Stock *int32
}

type DeleteProductRequest struct {
	ID uint64
}
