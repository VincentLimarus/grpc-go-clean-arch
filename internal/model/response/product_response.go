package response

import "time"

type ProductResponse struct {
	ID        uint64
	Name      string
	Price     float64
	Stock     int32
	CreatedAt time.Time
	UpdatedAt time.Time
}

type ProductListResponse = ListResponse[ProductResponse]
