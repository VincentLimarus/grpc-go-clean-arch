package gormrepo

import (
	"context"
	"time"

	"gorm.io/gorm"

	"VincentLimarus/grpc-golang/internal/domain/entity"
	"VincentLimarus/grpc-golang/internal/domain/repository"
)

type ProductModel struct {
	ID        uint64    `gorm:"column:id;primaryKey;autoIncrement"`
	Name      string    `gorm:"column:name;type:varchar(255);not null"`
	Price     float64   `gorm:"column:price;not null"`
	Stock     int32     `gorm:"column:stock;not null"`
	CreatedAt time.Time `gorm:"column:created_at;not null"`
	UpdatedAt time.Time `gorm:"column:updated_at;not null"`
}

func (ProductModel) TableName() string {
	return "products"
}

func AutoMigrate(db *gorm.DB) error {
	return db.AutoMigrate(&ProductModel{})
}

type ProductRepository struct {
	base *BaseRepository[ProductModel]
}

func NewProductRepository(db *gorm.DB) *ProductRepository {
	return &ProductRepository{base: NewBaseRepository[ProductModel](db)}
}

var _ repository.ProductRepository = (*ProductRepository)(nil)

func (r *ProductRepository) Create(ctx context.Context, p *entity.Product) error {
	m := toModel(p)
	if err := r.base.Create(ctx, &m); err != nil {
		return err
	}
	p.ID = m.ID
	p.CreatedAt = m.CreatedAt
	p.UpdatedAt = m.UpdatedAt
	return nil
}

func (r *ProductRepository) GetByID(ctx context.Context, id uint64) (*entity.Product, error) {
	m, err := r.base.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}
	return toEntity(m), nil
}

func (r *ProductRepository) List(ctx context.Context, page, limit int32) ([]*entity.Product, int64, error) {
	models, total, err := r.base.List(ctx, page, limit, "id ASC")
	if err != nil {
		return nil, 0, err
	}

	products := make([]*entity.Product, 0, len(models))
	for _, m := range models {
		products = append(products, toEntity(m))
	}
	return products, total, nil
}

func (r *ProductRepository) Update(ctx context.Context, p *entity.Product) error {
	values := map[string]any{
		"name":       p.Name,
		"price":      p.Price,
		"stock":      p.Stock,
		"updated_at": p.UpdatedAt,
	}
	return r.base.UpdateColumns(ctx, p.ID, values)
}

func (r *ProductRepository) Delete(ctx context.Context, id uint64) error {
	return r.base.Delete(ctx, id)
}

func toModel(p *entity.Product) ProductModel {
	return ProductModel{
		ID:        p.ID,
		Name:      p.Name,
		Price:     p.Price,
		Stock:     p.Stock,
		CreatedAt: p.CreatedAt,
		UpdatedAt: p.UpdatedAt,
	}
}

func toEntity(m *ProductModel) *entity.Product {
	return &entity.Product{
		ID:        m.ID,
		Name:      m.Name,
		Price:     m.Price,
		Stock:     m.Stock,
		CreatedAt: m.CreatedAt,
		UpdatedAt: m.UpdatedAt,
	}
}
