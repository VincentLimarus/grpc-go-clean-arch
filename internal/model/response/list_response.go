package response

type ListResponse[T any] struct {
	Items []*T
	Total int64
}

func NewListResponse[T any](items []*T, total int64) *ListResponse[T] {
	if items == nil {
		items = make([]*T, 0)
	}
	return &ListResponse[T]{Items: items, Total: total}
}
