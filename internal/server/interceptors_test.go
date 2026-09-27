package server

import (
	"bytes"
	"context"
	"errors"
	"io"
	"log/slog"
	"net"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/status"
	"google.golang.org/grpc/test/bufconn"

	"VincentLimarus/grpc-golang/internal/domain"
	"VincentLimarus/grpc-golang/internal/domain/entity"
	"VincentLimarus/grpc-golang/internal/pb"
)

type stubProductServer struct {
	pb.UnimplementedProductServiceServer
	getErr error
}

func (s *stubProductServer) GetProduct(ctx context.Context, req *pb.GetProductRequest) (*pb.Product, error) {
	return nil, s.getErr
}

func (s *stubProductServer) CreateProduct(ctx context.Context, req *pb.CreateProductRequest) (*pb.Product, error) {
	panic("boom")
}

func TestMapDomainError(t *testing.T) {
	table := []struct {
		name string
		err  error
		code codes.Code
	}{
		{name: "not found", err: domain.ErrNotFound, code: codes.NotFound},
		{name: "name required", err: entity.ErrNameRequired, code: codes.InvalidArgument},
		{name: "price invalid", err: entity.ErrPriceInvalid, code: codes.InvalidArgument},
		{name: "stock invalid", err: entity.ErrStockInvalid, code: codes.InvalidArgument},
		{name: "no patch fields", err: entity.ErrNoPatchFields, code: codes.InvalidArgument},
		{name: "context canceled", err: context.Canceled, code: codes.Canceled},
		{name: "deadline exceeded", err: context.DeadlineExceeded, code: codes.DeadlineExceeded},
		{name: "wrapped not found", err: errors.Join(domain.ErrNotFound), code: codes.NotFound},
		{name: "unknown", err: errors.New("db down"), code: codes.Internal},
	}
	for _, tc := range table {
		t.Run(tc.name, func(t *testing.T) {
			got := mapDomainError(tc.err, productErrorMappings)
			assert.Equal(t, tc.code, status.Code(got))
		})
	}
}

func TestErrorMappingInterceptor(t *testing.T) {
	interceptor := errorMappingInterceptor(productErrorMappings)

	table := []struct {
		name     string
		stubErr  error
		wantCode codes.Code
		wantMsg  string
	}{
		{name: "maps domain not found", stubErr: domain.ErrNotFound, wantCode: codes.NotFound, wantMsg: domain.ErrNotFound.Error()},
		{name: "maps domain validation", stubErr: entity.ErrNameRequired, wantCode: codes.InvalidArgument, wantMsg: entity.ErrNameRequired.Error()},
		{name: "maps unknown to internal", stubErr: errors.New("secret"), wantCode: codes.Internal, wantMsg: "internal server error"},
		{name: "passes through grpc status", stubErr: status.Error(codes.AlreadyExists, "exists"), wantCode: codes.AlreadyExists, wantMsg: "exists"},
	}
	for _, tc := range table {
		t.Run(tc.name, func(t *testing.T) {
			err := invokeGet(t, interceptor, tc.stubErr)
			require.Error(t, err)
			assert.Equal(t, tc.wantCode, status.Code(err))
			assert.Contains(t, status.Convert(err).Message(), tc.wantMsg)
		})
	}
}

func TestRecoveryInterceptor(t *testing.T) {
	logger := slog.New(slog.NewTextHandler(io.Discard, nil))
	interceptors := UnaryInterceptors(logger)

	srv := grpc.NewServer(grpc.ChainUnaryInterceptor(interceptors...))
	stub := &stubProductServer{}
	pb.RegisterProductServiceServer(srv, stub)

	client := dialBufconn(t, srv)
	_, err := client.CreateProduct(context.Background(), &pb.CreateProductRequest{Name: "X"})
	require.Error(t, err)
	assert.Equal(t, codes.Internal, status.Code(err))
	assert.NotContains(t, err.Error(), "boom")
}

func TestLoggingInterceptor(t *testing.T) {
	buf := &bytes.Buffer{}
	logger := slog.New(slog.NewTextHandler(buf, nil))
	interceptor := loggingInterceptor(logger)

	info := &grpc.UnaryServerInfo{FullMethod: "/pb.ProductService/GetProduct"}
	handler := func(ctx context.Context, req any) (any, error) {
		return nil, status.Error(codes.NotFound, "record not found")
	}

	resp, err := interceptor(context.Background(), &pb.GetProductRequest{}, info, handler)
	assert.Nil(t, resp)
	require.Error(t, err)

	output := buf.String()
	assert.Contains(t, output, "/pb.ProductService/GetProduct")
	assert.Contains(t, output, `code=NotFound`)
}

func TestUnaryInterceptors_RecoveryWrapsChain(t *testing.T) {
	logger := slog.New(slog.NewTextHandler(io.Discard, nil))
	srv := grpc.NewServer(grpc.ChainUnaryInterceptor(UnaryInterceptors(logger)...))
	pb.RegisterProductServiceServer(srv, &stubProductServer{})

	client := dialBufconn(t, srv)
	_, err := client.CreateProduct(context.Background(), &pb.CreateProductRequest{Name: "X"})
	require.Error(t, err)
	assert.Equal(t, codes.Internal, status.Code(err))
}

func invokeGet(t *testing.T, interceptor grpc.UnaryServerInterceptor, stubErr error) error {
	t.Helper()
	srv := grpc.NewServer(grpc.ChainUnaryInterceptor(interceptor))
	pb.RegisterProductServiceServer(srv, &stubProductServer{getErr: stubErr})

	client := dialBufconn(t, srv)
	_, err := client.GetProduct(context.Background(), &pb.GetProductRequest{Id: 1})
	return err
}

func dialBufconn(t *testing.T, srv *grpc.Server) pb.ProductServiceClient {
	t.Helper()
	lis := bufconn.Listen(1024 * 1024)
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

	return pb.NewProductServiceClient(conn)
}
