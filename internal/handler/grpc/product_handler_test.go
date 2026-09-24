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
	"VincentLimarus/grpc-golang/internal/pb"
	"VincentLimarus/grpc-golang/internal/server"
	"VincentLimarus/grpc-golang/internal/usecase"
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

func sampleEntity(id uint64) *entity.Product {
	ts := time.Date(2026, 9, 24, 12, 0, 0, 0, time.UTC)
	return &entity.Product{ID: id, Name: "Laptop", Price: 1500.5, Stock: 4, CreatedAt: ts, UpdatedAt: ts}
}

func TestCreateProduct(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		client, uc := startServer(t)
		in := usecase.CreateProductInput{Name: "Laptop", Price: 1500.5, Stock: 4}
		uc.On("CreateProduct", mock.Anything, in).Return(sampleEntity(1), nil).Once()

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
		in := usecase.CreateProductInput{Name: "", Price: 10, Stock: 1}
		uc.On("CreateProduct", mock.Anything, in).Return(nil, entity.ErrNameRequired).Once()

		_, err := client.CreateProduct(context.Background(), &pb.CreateProductRequest{Name: "", Price: 10, Stock: 1})
		require.Error(t, err)
		assert.Equal(t, codes.InvalidArgument, status.Code(err))
	})

	t.Run("internal error hides detail", func(t *testing.T) {
		client, uc := startServer(t)
		in := usecase.CreateProductInput{Name: "Laptop", Price: 10, Stock: 1}
		uc.On("CreateProduct", mock.Anything, in).Return(nil, errors.New("secret db detail")).Once()

		_, err := client.CreateProduct(context.Background(), &pb.CreateProductRequest{Name: "Laptop", Price: 10, Stock: 1})
		require.Error(t, err)
		assert.Equal(t, codes.Internal, status.Code(err))
		assert.NotContains(t, err.Error(), "secret")
	})
}

func TestGetProduct(t *testing.T) {
	t.Run("found", func(t *testing.T) {
		client, uc := startServer(t)
		uc.On("GetProduct", mock.Anything, uint64(7)).Return(sampleEntity(7), nil).Once()

		resp, err := client.GetProduct(context.Background(), &pb.GetProductRequest{Id: 7})
		require.NoError(t, err)
		assert.Equal(t, uint64(7), resp.Id)
	})

	t.Run("not found", func(t *testing.T) {
		client, uc := startServer(t)
		uc.On("GetProduct", mock.Anything, uint64(9)).Return(nil, domain.ErrNotFound).Once()

		_, err := client.GetProduct(context.Background(), &pb.GetProductRequest{Id: 9})
		require.Error(t, err)
		assert.Equal(t, codes.NotFound, status.Code(err))
	})
}

func TestListProducts(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		client, uc := startServer(t)
		uc.On("ListProducts", mock.Anything, int32(2), int32(5)).
			Return([]*entity.Product{sampleEntity(1), sampleEntity(2)}, int64(2), nil).Once()

		resp, err := client.ListProducts(context.Background(), &pb.ListProductsRequest{Page: 2, Limit: 5})
		require.NoError(t, err)
		require.Len(t, resp.Products, 2)
		assert.Equal(t, int64(2), resp.Total)
		assert.Equal(t, "Laptop", resp.Products[1].Name)
	})

	t.Run("empty", func(t *testing.T) {
		client, uc := startServer(t)
		uc.On("ListProducts", mock.Anything, int32(0), int32(0)).
			Return([]*entity.Product{}, int64(0), nil).Once()

		resp, err := client.ListProducts(context.Background(), &pb.ListProductsRequest{})
		require.NoError(t, err)
		assert.Empty(t, resp.Products)
		assert.Zero(t, resp.Total)
	})
}

func TestUpdateProduct(t *testing.T) {
	client, uc := startServer(t)
	updated := sampleEntity(7)
	updated.Name = "Laptop Pro"
	updated.Price = 2200
	updated.Stock = 2
	in := usecase.UpdateProductInput{ID: 7, Name: "Laptop Pro", Price: 2200, Stock: 2}
	uc.On("UpdateProduct", mock.Anything, in).Return(updated, nil).Once()

	resp, err := client.UpdateProduct(context.Background(), &pb.UpdateProductRequest{Id: 7, Name: "Laptop Pro", Price: 2200, Stock: 2})
	require.NoError(t, err)
	assert.Equal(t, "Laptop Pro", resp.Name)
	assert.Equal(t, 2200.0, resp.Price)
	assert.Equal(t, int32(2), resp.Stock)
}

func TestPatchProduct(t *testing.T) {
	t.Run("sends only provided fields", func(t *testing.T) {
		client, uc := startServer(t)

		var gotInput usecase.PatchProductInput
		uc.On("PatchProduct", mock.Anything, mock.MatchedBy(func(in usecase.PatchProductInput) bool {
			gotInput = in
			return true
		})).Return(sampleEntity(7), nil).Once()

		resp, err := client.PatchProduct(context.Background(), &pb.PatchProductRequest{Id: 7, Name: protoStr("Gaming")})
		require.NoError(t, err)
		assert.Equal(t, "Laptop", resp.Name)
		require.Equal(t, uint64(7), gotInput.ID)
		require.NotNil(t, gotInput.Patch.Name)
		assert.Equal(t, "Gaming", *gotInput.Patch.Name)
		assert.Nil(t, gotInput.Patch.Price)
		assert.Nil(t, gotInput.Patch.Stock)
	})

	t.Run("price only", func(t *testing.T) {
		client, uc := startServer(t)

		var gotInput usecase.PatchProductInput
		uc.On("PatchProduct", mock.Anything, mock.MatchedBy(func(in usecase.PatchProductInput) bool {
			gotInput = in
			return true
		})).Return(sampleEntity(7), nil).Once()

		_, err := client.PatchProduct(context.Background(), &pb.PatchProductRequest{Id: 7, Price: protoF64(999)})
		require.NoError(t, err)
		require.NotNil(t, gotInput.Patch.Price)
		assert.Equal(t, 999.0, *gotInput.Patch.Price)
		assert.Nil(t, gotInput.Patch.Name)
	})

	t.Run("no fields maps to invalid argument", func(t *testing.T) {
		client, uc := startServer(t)
		in := usecase.PatchProductInput{ID: 7}
		uc.On("PatchProduct", mock.Anything, in).Return(nil, entity.ErrNoPatchFields).Once()

		_, err := client.PatchProduct(context.Background(), &pb.PatchProductRequest{Id: 7})
		require.Error(t, err)
		assert.Equal(t, codes.InvalidArgument, status.Code(err))
	})
}

func TestDeleteProduct(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		client, uc := startServer(t)
		uc.On("DeleteProduct", mock.Anything, uint64(3)).Return(nil).Once()

		resp, err := client.DeleteProduct(context.Background(), &pb.DeleteProductRequest{Id: 3})
		require.NoError(t, err)
		assert.True(t, resp.Success)
	})

	t.Run("not found", func(t *testing.T) {
		client, uc := startServer(t)
		uc.On("DeleteProduct", mock.Anything, uint64(9)).Return(domain.ErrNotFound).Once()

		_, err := client.DeleteProduct(context.Background(), &pb.DeleteProductRequest{Id: 9})
		require.Error(t, err)
		assert.Equal(t, codes.NotFound, status.Code(err))
	})
}

func protoStr(v string) *string   { return &v }
func protoF64(v float64) *float64 { return &v }
