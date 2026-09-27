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

const errInternalMessage = "internal server error"

type ErrorMapping struct {
	Sentinel error
	Code     codes.Code
}

type ErrorRegistry struct {
	mappings []ErrorMapping
}

func NewErrorRegistry(mappings ...ErrorMapping) *ErrorRegistry {
	return &ErrorRegistry{mappings: append(make([]ErrorMapping, 0, len(mappings)), mappings...)}
}

func (r *ErrorRegistry) Register(sentinel error, code codes.Code) *ErrorRegistry {
	r.mappings = append(r.mappings, ErrorMapping{Sentinel: sentinel, Code: code})
	return r
}

func (r *ErrorRegistry) Status(err error) error {
	for _, m := range r.mappings {
		if errors.Is(err, m.Sentinel) {
			return status.Error(m.Code, err.Error())
		}
	}
	if errors.Is(err, context.Canceled) || errors.Is(err, context.DeadlineExceeded) {
		return status.FromContextError(err).Err()
	}
	return status.Error(codes.Internal, errInternalMessage)
}

var productErrorMappings = []ErrorMapping{
	{Sentinel: domain.ErrNotFound, Code: codes.NotFound},
	{Sentinel: entity.ErrNameRequired, Code: codes.InvalidArgument},
	{Sentinel: entity.ErrPriceInvalid, Code: codes.InvalidArgument},
	{Sentinel: entity.ErrStockInvalid, Code: codes.InvalidArgument},
	{Sentinel: entity.ErrNoPatchFields, Code: codes.InvalidArgument},
}

func DefaultErrorRegistry() *ErrorRegistry {
	return NewErrorRegistry(productErrorMappings...)
}

func UnaryInterceptors(logger *slog.Logger) []grpc.UnaryServerInterceptor {
	return []grpc.UnaryServerInterceptor{
		recoveryInterceptor(logger),
		loggingInterceptor(logger),
		errorMappingInterceptor(DefaultErrorRegistry()),
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
				err = status.Error(codes.Internal, errInternalMessage)
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

func errorMappingInterceptor(registry *ErrorRegistry) grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req any, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (any, error) {
		resp, err := handler(ctx, req)
		if err == nil {
			return resp, nil
		}
		if s, ok := status.FromError(err); ok && s.Code() != codes.Unknown {
			return nil, err
		}
		return nil, registry.Status(err)
	}
}
