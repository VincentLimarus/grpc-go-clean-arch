package grpchandler_test

import (
	"context"
	"errors"
	"io"
	"log/slog"
	"net"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/status"
	"google.golang.org/grpc/test/bufconn"

	"VincentLimarus/grpc-golang/internal/domain"
	"VincentLimarus/grpc-golang/internal/domain/entity"
	grpchandler "VincentLimarus/grpc-golang/internal/handler/grpc"
	"VincentLimarus/grpc-golang/internal/mocks"
	"VincentLimarus/grpc-golang/internal/model/request"
	"VincentLimarus/grpc-golang/internal/model/response"
	"VincentLimarus/grpc-golang/internal/pb"
	"VincentLimarus/grpc-golang/internal/server"
)

func startServer(t *testing.T) (pb.ProductServiceClient, *mocks.ProductUseCase) {
	t.Helper()
	uc := mocks.NewProductUseCase()
	logger := slog.New(slog.NewTextHandler(io.Discard, nil))

	lis := bufconn.Listen(1024 * 1024)
	srv := grpc.NewServer(grpc.ChainUnaryInterceptor(server.UnaryInterceptors(logger)...))
	pb.RegisterProductServiceServer(srv, grpchandler.NewProductHandler(uc))
	go func() { _ = srv.Serve(lis) }()
	t.Cleanup(srv.Stop)

	conn, err := grpc.NewClient("passthrough:///bufnet",
		grpc.WithContextDialer(func(ctx context.Context, _ string) (net.Conn, error) {
			return lis.DialContext(ctx)
		}),
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	)
	require.NoError(t, err)
	t.Cleanup(func() { _ = conn.Close() })

	return pb.NewProductServiceClient(conn), uc
}

func sampleResponse(id uint64) *response.ProductResponse {
	ts := time.Date(2026, 9, 24, 12, 0, 0, 0, time.UTC)
	return &response.ProductResponse{ID: id, Name: "Laptop", Price: 1500.5, Stock: 4, CreatedAt: ts, UpdatedAt: ts}
}

func TestCreateProduct(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		client, uc := startServer(t)
		req := request.CreateProductRequest{Name: "Laptop", Price: 1500.5, Stock: 4}
		uc.On("CreateProduct", mock.Anything, req).Return(sampleResponse(1), nil).Once()

		resp, err := client.CreateProduct(context.Background(), &pb.CreateProductRequest{Name: "Laptop", Price: 1500.5, Stock: 4})
		require.NoError(t, err)
		assert.Equal(t, uint64(1), resp.Id)
		assert.Equal(t, "Laptop", resp.Name)
		assert.Equal(t, 1500.5, resp.Price)
		assert.Equal(t, int32(4), resp.Stock)
		assert.Equal(t, int64(1790251200), resp.CreatedAt)
		assert.Equal(t, int64(1790251200), resp.UpdatedAt)
	})

	t.Run("invalid argument", func(t *testing.T) {
		client, uc := startServer(t)
		req := request.CreateProductRequest{Name: "", Price: 10, Stock: 1}
		uc.On("CreateProduct", mock.Anything, req).Return(nil, entity.ErrNameRequired).Once()

		_, err := client.CreateProduct(context.Background(), &pb.CreateProductRequest{Name: "", Price: 10, Stock: 1})
		require.Error(t, err)
		assert.Equal(t, codes.InvalidArgument, status.Code(err))
	})

	t.Run("internal error hides detail", func(t *testing.T) {
		client, uc := startServer(t)
		req := request.CreateProductRequest{Name: "Laptop", Price: 10, Stock: 1}
		uc.On("CreateProduct", mock.Anything, req).Return(nil, errors.New("secret db detail")).Once()

		_, err := client.CreateProduct(context.Background(), &pb.CreateProductRequest{Name: "Laptop", Price: 10, Stock: 1})
		require.Error(t, err)
		assert.Equal(t, codes.Internal, status.Code(err))
		assert.NotContains(t, err.Error(), "secret")
	})
}

func TestGetProduct(t *testing.T) {
	t.Run("found", func(t *testing.T) {
		client, uc := startServer(t)
		uc.On("GetProduct", mock.Anything, request.GetProductRequest{ID: 7}).Return(sampleResponse(7), nil).Once()

		resp, err := client.GetProduct(context.Background(), &pb.GetProductRequest{Id: 7})
		require.NoError(t, err)
		assert.Equal(t, uint64(7), resp.Id)
	})

	t.Run("not found", func(t *testing.T) {
		client, uc := startServer(t)
		uc.On("GetProduct", mock.Anything, request.GetProductRequest{ID: 9}).Return(nil, domain.ErrNotFound).Once()

		_, err := client.GetProduct(context.Background(), &pb.GetProductRequest{Id: 9})
		require.Error(t, err)
		assert.Equal(t, codes.NotFound, status.Code(err))
	})

	t.Run("nil response does not panic", func(t *testing.T) {
		client, uc := startServer(t)
		uc.On("GetProduct", mock.Anything, request.GetProductRequest{ID: 5}).Return(nil, nil).Once()

		resp, err := client.GetProduct(context.Background(), &pb.GetProductRequest{Id: 5})
		if err == nil {
			assert.Equal(t, uint64(0), resp.GetId())
		}
	})
}

func TestListProducts(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		client, uc := startServer(t)
		uc.On("ListProducts", mock.Anything, request.ListProductsRequest{Page: 2, Limit: 5}).
			Return(response.NewListResponse([]*response.ProductResponse{sampleResponse(1), sampleResponse(2)}, 2), nil).Once()

		resp, err := client.ListProducts(context.Background(), &pb.ListProductsRequest{Page: 2, Limit: 5})
		require.NoError(t, err)
		require.Len(t, resp.Products, 2)
		assert.Equal(t, int64(2), resp.Total)
		assert.Equal(t, "Laptop", resp.Products[1].Name)
	})

	t.Run("empty", func(t *testing.T) {
		client, uc := startServer(t)
		uc.On("ListProducts", mock.Anything, request.ListProductsRequest{}).
			Return(response.NewListResponse[*response.ProductResponse](nil, 0), nil).Once()

		resp, err := client.ListProducts(context.Background(), &pb.ListProductsRequest{})
		require.NoError(t, err)
		assert.Empty(t, resp.Products)
		assert.Zero(t, resp.Total)
	})

	t.Run("repo error", func(t *testing.T) {
		client, uc := startServer(t)
		uc.On("ListProducts", mock.Anything, request.ListProductsRequest{Page: 1, Limit: 10}).
			Return(nil, errors.New("db down")).Once()

		_, err := client.ListProducts(context.Background(), &pb.ListProductsRequest{Page: 1, Limit: 10})
		require.Error(t, err)
		assert.Equal(t, codes.Internal, status.Code(err))
	})
}

func TestUpdateProduct(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		client, uc := startServer(t)
		updated := sampleResponse(7)
		updated.Name = "Laptop Pro"
		updated.Price = 2200
		updated.Stock = 2
		req := request.UpdateProductRequest{ID: 7, Name: "Laptop Pro", Price: 2200, Stock: 2}
		uc.On("UpdateProduct", mock.Anything, req).Return(updated, nil).Once()

		resp, err := client.UpdateProduct(context.Background(), &pb.UpdateProductRequest{Id: 7, Name: "Laptop Pro", Price: 2200, Stock: 2})
		require.NoError(t, err)
		assert.Equal(t, "Laptop Pro", resp.Name)
		assert.Equal(t, 2200.0, resp.Price)
		assert.Equal(t, int32(2), resp.Stock)
	})

	t.Run("not found", func(t *testing.T) {
		client, uc := startServer(t)
		req := request.UpdateProductRequest{ID: 9, Name: "Laptop Pro", Price: 2200, Stock: 2}
		uc.On("UpdateProduct", mock.Anything, req).Return(nil, domain.ErrNotFound).Once()

		_, err := client.UpdateProduct(context.Background(), &pb.UpdateProductRequest{Id: 9, Name: "Laptop Pro", Price: 2200, Stock: 2})
		require.Error(t, err)
		assert.Equal(t, codes.NotFound, status.Code(err))
	})
}

func TestPatchProduct(t *testing.T) {
	t.Run("sends only provided fields", func(t *testing.T) {
		client, uc := startServer(t)

		var gotReq request.PatchProductRequest
		uc.On("PatchProduct", mock.Anything, mock.MatchedBy(func(in request.PatchProductRequest) bool {
			gotReq = in
			return true
		})).Return(sampleResponse(7), nil).Once()

		resp, err := client.PatchProduct(context.Background(), &pb.PatchProductRequest{Id: 7, Name: protoStr("Gaming")})
		require.NoError(t, err)
		assert.Equal(t, "Laptop", resp.Name)
		require.Equal(t, uint64(7), gotReq.ID)
		require.NotNil(t, gotReq.Name)
		assert.Equal(t, "Gaming", *gotReq.Name)
		assert.Nil(t, gotReq.Price)
		assert.Nil(t, gotReq.Stock)
	})

	t.Run("price only", func(t *testing.T) {
		client, uc := startServer(t)

		var gotReq request.PatchProductRequest
		uc.On("PatchProduct", mock.Anything, mock.MatchedBy(func(in request.PatchProductRequest) bool {
			gotReq = in
			return true
		})).Return(sampleResponse(7), nil).Once()

		_, err := client.PatchProduct(context.Background(), &pb.PatchProductRequest{Id: 7, Price: protoF64(999)})
		require.NoError(t, err)
		require.NotNil(t, gotReq.Price)
		assert.Equal(t, 999.0, *gotReq.Price)
		assert.Nil(t, gotReq.Name)
	})

	t.Run("no fields maps to invalid argument", func(t *testing.T) {
		client, uc := startServer(t)
		uc.On("PatchProduct", mock.Anything, request.PatchProductRequest{ID: 7}).Return(nil, entity.ErrNoPatchFields).Once()

		_, err := client.PatchProduct(context.Background(), &pb.PatchProductRequest{Id: 7})
		require.Error(t, err)
		assert.Equal(t, codes.InvalidArgument, status.Code(err))
	})
}

func TestDeleteProduct(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		client, uc := startServer(t)
		uc.On("DeleteProduct", mock.Anything, request.DeleteProductRequest{ID: 3}).Return(nil).Once()

		resp, err := client.DeleteProduct(context.Background(), &pb.DeleteProductRequest{Id: 3})
		require.NoError(t, err)
		assert.True(t, resp.Success)
	})

	t.Run("not found", func(t *testing.T) {
		client, uc := startServer(t)
		uc.On("DeleteProduct", mock.Anything, request.DeleteProductRequest{ID: 9}).Return(domain.ErrNotFound).Once()

		_, err := client.DeleteProduct(context.Background(), &pb.DeleteProductRequest{Id: 9})
		require.Error(t, err)
		assert.Equal(t, codes.NotFound, status.Code(err))
	})
}

func protoStr(v string) *string   { return &v }
func protoF64(v float64) *float64 { return &v }
