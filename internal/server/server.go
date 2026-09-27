package server

import (
	"log/slog"
	"net"

	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"

	"VincentLimarus/grpc-golang/internal/config"
	grpchandler "VincentLimarus/grpc-golang/internal/handler/grpc"
	"VincentLimarus/grpc-golang/internal/pb"
	gormrepo "VincentLimarus/grpc-golang/internal/repository/gorm"
	"VincentLimarus/grpc-golang/internal/usecase"
)

type Server struct {
	cfg      *config.Config
	logger   *slog.Logger
	grpc     *grpc.Server
	listener net.Listener
}

func New(cfg *config.Config, logger *slog.Logger) (*Server, error) {
	db, err := connect(cfg.DSN())
	if err != nil {
		return nil, err
	}
	if err := gormrepo.AutoMigrate(db); err != nil {
		return nil, err
	}

	lis, err := net.Listen("tcp", cfg.GRPCAddr)
	if err != nil {
		return nil, err
	}

	srv := grpc.NewServer(grpc.ChainUnaryInterceptor(UnaryInterceptors(logger)...))
	registerServices(srv, db)
	reflection.Register(srv)

	return &Server{
		cfg:      cfg,
		logger:   logger,
		grpc:     srv,
		listener: lis,
	}, nil
}

func (s *Server) Serve() error {
	return s.grpc.Serve(s.listener)
}

func (s *Server) GracefulStop() {
	s.grpc.GracefulStop()
}

func (s *Server) Addr() string {
	return s.listener.Addr().String()
}

func connect(dsn string) (*gorm.DB, error) {
	return gorm.Open(postgres.Open(dsn), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Warn),
	})
}

func registerServices(s *grpc.Server, db *gorm.DB) {
	productRepo := gormrepo.NewProductRepository(db)
	productUC := usecase.NewProductUseCase(productRepo)
	pb.RegisterProductServiceServer(s, grpchandler.NewProductHandler(productUC))
}
