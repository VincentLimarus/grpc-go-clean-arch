package result_test

import (
	"errors"
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"VincentLimarus/grpc-golang/internal/domain"
	"VincentLimarus/grpc-golang/internal/shared/result"
)

func TestNotNil(t *testing.T) {
	t.Run("returns value", func(t *testing.T) {
		value := 42
		got, err := result.NotNil(&value, nil)
		require.NoError(t, err)
		assert.Equal(t, 42, *got)
	})

	t.Run("nil value without error becomes not found", func(t *testing.T) {
		got, err := result.NotNil[int](nil, nil)
		require.ErrorIs(t, err, domain.ErrNotFound)
		assert.Nil(t, got)
	})

	t.Run("propagates error", func(t *testing.T) {
		value := 1
		want := errors.New("db down")
		got, err := result.NotNil(&value, want)
		require.ErrorIs(t, err, want)
		assert.Nil(t, got)
	})

	t.Run("error wins over value", func(t *testing.T) {
		value := 1
		got, err := result.NotNil(&value, fmt.Errorf("wrapped: %w", domain.ErrNotFound))
		require.ErrorIs(t, err, domain.ErrNotFound)
		assert.Nil(t, got)
	})
}
