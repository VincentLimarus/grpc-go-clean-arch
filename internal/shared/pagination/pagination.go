package pagination

const (
	DefaultPage  int32 = 1
	DefaultLimit int32 = 10
	MaxLimit     int32 = 100
)

type Params struct {
	Page  int32
	Limit int32
}

func New(page, limit int32) Params {
	return Params{
		Page:  normalizePage(page),
		Limit: normalizeLimit(limit),
	}
}

func (p Params) Offset() int {
	return int((p.Page - 1) * p.Limit)
}

func (p Params) Size() int {
	return int(p.Limit)
}

func (p Params) IsPaged() bool {
	return p.Page > 0 && p.Limit > 0
}

func normalizePage(page int32) int32 {
	if page < DefaultPage {
		return DefaultPage
	}
	return page
}

func normalizeLimit(limit int32) int32 {
	if limit < 1 {
		return DefaultLimit
	}
	if limit > MaxLimit {
		return MaxLimit
	}
	return limit
}
