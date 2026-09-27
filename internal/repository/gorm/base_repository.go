package gormrepo

import (
	"context"
	"errors"

	"gorm.io/gorm"

	"VincentLimarus/grpc-golang/internal/domain"
	"VincentLimarus/grpc-golang/internal/shared/pagination"
)

type BaseRepository[T any] struct {
	db *gorm.DB
}

func NewBaseRepository[T any](db *gorm.DB) *BaseRepository[T] {
	return &BaseRepository[T]{db: db}
}

func (r *BaseRepository[T]) Create(ctx context.Context, m *T) error {
	return r.db.WithContext(ctx).Create(m).Error
}

func (r *BaseRepository[T]) GetByID(ctx context.Context, id uint64) (*T, error) {
	var m T
	if err := r.db.WithContext(ctx).First(&m, "id = ?", id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, domain.ErrNotFound
		}
		return nil, err
	}
	return &m, nil
}

func (r *BaseRepository[T]) List(ctx context.Context, params pagination.Params, orderBy string) ([]*T, int64, error) {
	var total int64
	if err := r.db.WithContext(ctx).Model(new(T)).Count(&total).Error; err != nil {
		return nil, 0, err
	}

	q := r.db.WithContext(ctx).Model(new(T))
	if orderBy != "" {
		q = q.Order(orderBy)
	}
	if params.IsPaged() {
		q = q.Offset(params.Offset()).Limit(params.Size())
	}

	items := make([]*T, 0)
	if err := q.Find(&items).Error; err != nil {
		return nil, 0, err
	}
	return items, total, nil
}

func (r *BaseRepository[T]) UpdateColumns(ctx context.Context, id uint64, values map[string]any) error {
	res := r.db.WithContext(ctx).Model(new(T)).Where("id = ?", id).Updates(values)
	if res.Error != nil {
		return res.Error
	}
	if res.RowsAffected == 0 {
		return domain.ErrNotFound
	}
	return nil
}

func (r *BaseRepository[T]) Delete(ctx context.Context, id uint64) error {
	res := r.db.WithContext(ctx).Delete(new(T), "id = ?", id)
	if res.Error != nil {
		return res.Error
	}
	if res.RowsAffected == 0 {
		return domain.ErrNotFound
	}
	return nil
}
