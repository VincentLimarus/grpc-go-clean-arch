package result

import "VincentLimarus/grpc-golang/internal/domain"

func NotNil[T any](value *T, err error) (*T, error) {
	if err != nil {
		return nil, err
	}
	if value == nil {
		return nil, domain.ErrNotFound
	}
	return value, nil
}
