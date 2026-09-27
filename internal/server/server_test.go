package server

import (
	"context"
	"io"
	"log/slog"
	"net"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"

	"VincentLimarus/grpc-golang/internal/pb"
)

type stubService struct {
	pb.UnimplementedProductServiceServer
}

func (stubService) GetProduct(ctx context.Context, req *pb.GetProductRequest) (*pb.Product, error) {
	return &pb.Product{Id: req.GetId(), Name: "Laptop", Price: 10, Stock: 1}, nil
}

func TestServerServeAndGracefulStop(t *testing.T) {
	lis, err := net.Listen("tcp", "127.0.0.1:0")
	require.NoError(t, err)
	logger := slog.New(slog.NewTextHandler(io.Discard, nil))
	srv := &Server{
		logger:   logger,
		grpc:     grpc.NewServer(grpc.ChainUnaryInterceptor(UnaryInterceptors(logger)...)),
		listener: lis,
	}
	pb.RegisterProductServiceServer(srv.grpc, stubService{})

	errCh := make(chan error, 1)
	go func() { errCh <- srv.Serve() }()

	assert.Contains(t, srv.Addr(), "127.0.0.1:")

	conn, err := grpc.NewClient(srv.Addr(),
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	)
	require.NoError(t, err)
	t.Cleanup(func() { _ = conn.Close() })

	client := pb.NewProductServiceClient(conn)
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()
	resp, err := client.GetProduct(ctx, &pb.GetProductRequest{Id: 1})
	require.NoError(t, err)
	assert.Equal(t, "Laptop", resp.Name)

	srv.GracefulStop()
	assert.NoError(t, <-errCh)
}

func TestRegisterServices(t *testing.T) {
	s := grpc.NewServer()
	registerServices(s, nil)

	assert.Contains(t, s.GetServiceInfo(), "pb.ProductService")
}
