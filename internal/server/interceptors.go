package server

import (
	"context"
	"errors"
	"log/slog"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"VincentLimarus/grpc-golang/internal/domain"
	"VincentLimarus/grpc-golang/internal/domain/entity"
)

type ErrorMapping struct {
	sentinel error
	code     codes.Code
}

var productErrorMappings = []ErrorMapping{
	{sentinel: domain.ErrNotFound, code: codes.NotFound},
	{sentinel: entity.ErrNameRequired, code: codes.InvalidArgument},
	{sentinel: entity.ErrPriceInvalid, code: codes.InvalidArgument},
	{sentinel: entity.ErrStockInvalid, code: codes.InvalidArgument},
	{sentinel: entity.ErrNoPatchFields, code: codes.InvalidArgument},
}

func UnaryInterceptors(logger *slog.Logger) []grpc.UnaryServerInterceptor {
	return []grpc.UnaryServerInterceptor{
		recoveryInterceptor(logger),
		loggingInterceptor(logger),
		errorMappingInterceptor(productErrorMappings),
	}
}

func recoveryInterceptor(logger *slog.Logger) grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req any, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (resp any, err error) {
		defer func() {
			if r := recover(); r != nil {
				logger.Error("panic recovered",
					"method", info.FullMethod,
					"panic", r,
				)
				resp = nil
				err = status.Error(codes.Internal, "internal server error")
			}
		}()
		return handler(ctx, req)
	}
}

func loggingInterceptor(logger *slog.Logger) grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req any, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (any, error) {
		start := time.Now()
		resp, err := handler(ctx, req)
		logger.Info("grpc call",
			"method", info.FullMethod,
			"code", status.Code(err).String(),
			"duration", time.Since(start).String(),
		)
		return resp, err
	}
}

func errorMappingInterceptor(mappings []ErrorMapping) grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req any, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (any, error) {
		resp, err := handler(ctx, req)
		if err == nil {
			return resp, nil
		}
		if s, ok := status.FromError(err); ok && s.Code() != codes.Unknown {
			return nil, err
		}
		return nil, mapDomainError(err, mappings)
	}
}

func mapDomainError(err error, mappings []ErrorMapping) error {
	for _, m := range mappings {
		if errors.Is(err, m.sentinel) {
			return status.Error(m.code, err.Error())
		}
	}
	if errors.Is(err, context.Canceled) || errors.Is(err, context.DeadlineExceeded) {
		return status.FromContextError(err).Err()
	}
	return status.Error(codes.Internal, "internal server error")
}
