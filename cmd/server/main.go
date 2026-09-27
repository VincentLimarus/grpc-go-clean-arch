package main

import (
	"context"
	"log/slog"
	"os"
	"os/signal"
	"syscall"

	"VincentLimarus/grpc-golang/internal/config"
	"VincentLimarus/grpc-golang/internal/server"
)

func main() {
	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))
	cfg := config.Load()

	srv, err := server.New(cfg, logger)
	if err != nil {
		logger.Error("failed to initialize server", "error", err)
		os.Exit(1)
	}

	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	go func() {
		<-ctx.Done()
		logger.Info("shutting down gracefully")
		srv.GracefulStop()
	}()

	logger.Info("product gRPC server listening", "addr", srv.Addr())
	if err := srv.Serve(); err != nil {
		logger.Error("failed to serve", "error", err)
		os.Exit(1)
	}
}
