package grpchandler

import (
	"context"

	"VincentLimarus/grpc-golang/internal/domain/entity"
	"VincentLimarus/grpc-golang/internal/pb"
	"VincentLimarus/grpc-golang/internal/usecase"
)

type ProductHandler struct {
	pb.UnimplementedProductServiceServer
	uc usecase.ProductUseCase
}

func NewProductHandler(uc usecase.ProductUseCase) *ProductHandler {
	return &ProductHandler{uc: uc}
}

var _ pb.ProductServiceServer = (*ProductHandler)(nil)

func (h *ProductHandler) GetProduct(ctx context.Context, req *pb.GetProductRequest) (*pb.Product, error) {
	p, err := h.uc.GetProduct(ctx, req.GetId())
	if err != nil {
		return nil, err
	}
	return toProto(p), nil
}

func (h *ProductHandler) ListProducts(ctx context.Context, req *pb.ListProductsRequest) (*pb.ListProductsResponse, error) {
	products, total, err := h.uc.ListProducts(ctx, req.GetPage(), req.GetLimit())
	if err != nil {
		return nil, err
	}
	resp := &pb.ListProductsResponse{
		Products: make([]*pb.Product, 0, len(products)),
		Total:    total,
	}
	for _, p := range products {
		resp.Products = append(resp.Products, toProto(p))
	}
	return resp, nil
}

func (h *ProductHandler) CreateProduct(ctx context.Context, req *pb.CreateProductRequest) (*pb.Product, error) {
	p, err := h.uc.CreateProduct(ctx, usecase.CreateProductInput{
		Name:  req.GetName(),
		Price: req.GetPrice(),
		Stock: req.GetStock(),
	})
	if err != nil {
		return nil, err
	}
	return toProto(p), nil
}

func (h *ProductHandler) UpdateProduct(ctx context.Context, req *pb.UpdateProductRequest) (*pb.Product, error) {
	p, err := h.uc.UpdateProduct(ctx, usecase.UpdateProductInput{
		ID:    req.GetId(),
		Name:  req.GetName(),
		Price: req.GetPrice(),
		Stock: req.GetStock(),
	})
	if err != nil {
		return nil, err
	}
	return toProto(p), nil
}

func (h *ProductHandler) PatchProduct(ctx context.Context, req *pb.PatchProductRequest) (*pb.Product, error) {
	in := usecase.PatchProductInput{ID: req.GetId()}
	if req.Name != nil {
		name := req.GetName()
		in.Patch.Name = &name
	}
	if req.Price != nil {
		price := req.GetPrice()
		in.Patch.Price = &price
	}
	if req.Stock != nil {
		stock := req.GetStock()
		in.Patch.Stock = &stock
	}

	p, err := h.uc.PatchProduct(ctx, in)
	if err != nil {
		return nil, err
	}
	return toProto(p), nil
}

func (h *ProductHandler) DeleteProduct(ctx context.Context, req *pb.DeleteProductRequest) (*pb.DeleteProductResponse, error) {
	if err := h.uc.DeleteProduct(ctx, req.GetId()); err != nil {
		return nil, err
	}
	return &pb.DeleteProductResponse{Success: true}, nil
}

func toProto(p *entity.Product) *pb.Product {
	return &pb.Product{
		Id:        p.ID,
		Name:      p.Name,
		Price:     p.Price,
		Stock:     p.Stock,
		CreatedAt: p.CreatedAt.Unix(),
		UpdatedAt: p.UpdatedAt.Unix(),
	}
}
