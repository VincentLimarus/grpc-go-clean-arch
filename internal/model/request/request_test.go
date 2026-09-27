package request_test

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"VincentLimarus/grpc-golang/internal/model/request"
	"VincentLimarus/grpc-golang/internal/shared/pagination"
)

func TestListProductsRequest_Pagination(t *testing.T) {
	table := []struct {
		name string
		req  request.ListProductsRequest
		want pagination.Params
	}{
		{name: "zero values are normalized", req: request.ListProductsRequest{}, want: pagination.New(1, 10)},
		{name: "values are kept", req: request.ListProductsRequest{Page: 2, Limit: 5}, want: pagination.New(2, 5)},
		{name: "limit is capped", req: request.ListProductsRequest{Page: 3, Limit: 500}, want: pagination.New(3, 100)},
	}
	for _, tc := range table {
		t.Run(tc.name, func(t *testing.T) {
			assert.Equal(t, tc.want, tc.req.Pagination())
		})
	}
}
