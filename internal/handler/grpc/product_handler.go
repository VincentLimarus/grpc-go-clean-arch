package grpchandler

import (
	"context"

	"VincentLimarus/grpc-golang/internal/model/request"
	"VincentLimarus/grpc-golang/internal/model/response"
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
	res, err := h.uc.GetProduct(ctx, toGetProductRequest(req))
	if err != nil {
		return nil, err
	}
	return toProductMessage(res), nil
}

func (h *ProductHandler) ListProducts(ctx context.Context, req *pb.ListProductsRequest) (*pb.ListProductsResponse, error) {
	res, err := h.uc.ListProducts(ctx, toListProductsRequest(req))
	if err != nil {
		return nil, err
	}
	return toListProductsMessage(res), nil
}

func (h *ProductHandler) CreateProduct(ctx context.Context, req *pb.CreateProductRequest) (*pb.Product, error) {
	res, err := h.uc.CreateProduct(ctx, toCreateProductRequest(req))
	if err != nil {
		return nil, err
	}
	return toProductMessage(res), nil
}

func (h *ProductHandler) UpdateProduct(ctx context.Context, req *pb.UpdateProductRequest) (*pb.Product, error) {
	res, err := h.uc.UpdateProduct(ctx, toUpdateProductRequest(req))
	if err != nil {
		return nil, err
	}
	return toProductMessage(res), nil
}

func (h *ProductHandler) PatchProduct(ctx context.Context, req *pb.PatchProductRequest) (*pb.Product, error) {
	res, err := h.uc.PatchProduct(ctx, toPatchProductRequest(req))
	if err != nil {
		return nil, err
	}
	return toProductMessage(res), nil
}

func (h *ProductHandler) DeleteProduct(ctx context.Context, req *pb.DeleteProductRequest) (*pb.DeleteProductResponse, error) {
	if err := h.uc.DeleteProduct(ctx, toDeleteProductRequest(req)); err != nil {
		return nil, err
	}
	return &pb.DeleteProductResponse{Success: true}, nil
}

func toGetProductRequest(req *pb.GetProductRequest) request.GetProductRequest {
	return request.GetProductRequest{ID: req.GetId()}
}

func toListProductsRequest(req *pb.ListProductsRequest) request.ListProductsRequest {
	return request.ListProductsRequest{Page: req.GetPage(), Limit: req.GetLimit()}
}

func toCreateProductRequest(req *pb.CreateProductRequest) request.CreateProductRequest {
	return request.CreateProductRequest{
		Name:  req.GetName(),
		Price: req.GetPrice(),
		Stock: req.GetStock(),
	}
}

func toUpdateProductRequest(req *pb.UpdateProductRequest) request.UpdateProductRequest {
	return request.UpdateProductRequest{
		ID:    req.GetId(),
		Name:  req.GetName(),
		Price: req.GetPrice(),
		Stock: req.GetStock(),
	}
}

func toPatchProductRequest(req *pb.PatchProductRequest) request.PatchProductRequest {
	return request.PatchProductRequest{
		ID:    req.GetId(),
		Name:  req.Name,
		Price: req.Price,
		Stock: req.Stock,
	}
}

func toDeleteProductRequest(req *pb.DeleteProductRequest) request.DeleteProductRequest {
	return request.DeleteProductRequest{ID: req.GetId()}
}

func toProductMessage(res *response.ProductResponse) *pb.Product {
	if res == nil {
		return nil
	}
	return &pb.Product{
		Id:        res.ID,
		Name:      res.Name,
		Price:     res.Price,
		Stock:     res.Stock,
		CreatedAt: res.CreatedAt.Unix(),
		UpdatedAt: res.UpdatedAt.Unix(),
	}
}

func toListProductsMessage(res *response.ProductListResponse) *pb.ListProductsResponse {
	if res == nil {
		return &pb.ListProductsResponse{Products: make([]*pb.Product, 0)}
	}
	out := &pb.ListProductsResponse{
		Products: make([]*pb.Product, 0, len(res.Items)),
		Total:    res.Total,
	}
	for _, item := range res.Items {
		out.Products = append(out.Products, toProductMessage(item))
	}
	return out
}
