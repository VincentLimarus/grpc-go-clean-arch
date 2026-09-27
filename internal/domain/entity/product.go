package entity

import (
	"errors"
	"strings"
	"time"
)

var (
	ErrNameRequired  = errors.New("product name is required")
	ErrPriceInvalid  = errors.New("product price must be greater than zero")
	ErrStockInvalid  = errors.New("product stock cannot be negative")
	ErrNoPatchFields = errors.New("patch must contain at least one field")
)

type Product struct {
	ID        uint64
	Name      string
	Price     float64
	Stock     int32
	CreatedAt time.Time
	UpdatedAt time.Time
}

type ProductPatch struct {
	Name  *string
	Price *float64
	Stock *int32
}

func NewProduct(name string, price float64, stock int32) (*Product, error) {
	now := time.Now().UTC()
	p := &Product{
		Name:      strings.TrimSpace(name),
		Price:     price,
		Stock:     stock,
		CreatedAt: now,
		UpdatedAt: now,
	}
	if err := p.Validate(); err != nil {
		return nil, err
	}
	return p, nil
}

func (p *Product) Validate() error {
	if strings.TrimSpace(p.Name) == "" {
		return ErrNameRequired
	}
	if p.Price <= 0 {
		return ErrPriceInvalid
	}
	if p.Stock < 0 {
		return ErrStockInvalid
	}
	return nil
}

func (p *Product) ApplyUpdate(name string, price float64, stock int32) error {
	p.Name = strings.TrimSpace(name)
	p.Price = price
	p.Stock = stock
	p.touch()
	return p.Validate()
}

func (p *Product) ApplyPatch(patch ProductPatch) error {
	if patch.Name == nil && patch.Price == nil && patch.Stock == nil {
		return ErrNoPatchFields
	}
	if patch.Name != nil {
		p.Name = strings.TrimSpace(*patch.Name)
	}
	if patch.Price != nil {
		p.Price = *patch.Price
	}
	if patch.Stock != nil {
		p.Stock = *patch.Stock
	}
	p.touch()
	return p.Validate()
}

func (p *Product) touch() {
	p.UpdatedAt = time.Now().UTC()
}
