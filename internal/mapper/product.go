package mapper

import (
	"VincentLimarus/grpc-golang/internal/domain/entity"
	"VincentLimarus/grpc-golang/internal/model/request"
	"VincentLimarus/grpc-golang/internal/model/response"
)

func ToProductResponse(p *entity.Product) *response.ProductResponse {
	if p == nil {
		return nil
	}
	return &response.ProductResponse{
		ID:        p.ID,
		Name:      p.Name,
		Price:     p.Price,
		Stock:     p.Stock,
		CreatedAt: p.CreatedAt,
		UpdatedAt: p.UpdatedAt,
	}
}

func ToProductResponses(products []*entity.Product) []*response.ProductResponse {
	items := make([]*response.ProductResponse, 0, len(products))
	for _, p := range products {
		items = append(items, ToProductResponse(p))
	}
	return items
}

func ToProductPatch(req request.PatchProductRequest) entity.ProductPatch {
	return entity.ProductPatch{
		Name:  req.Name,
		Price: req.Price,
		Stock: req.Stock,
	}
}
