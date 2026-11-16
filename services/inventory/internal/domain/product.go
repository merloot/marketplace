package domain

import "github.com/google/uuid"

type ProductID uuid.UUID

type ProductStock struct {
	productID uuid.UUID
	sku       SKU
	quantity  int
	reserved  int
}

func NewProductStock(productID uuid.UUID, sku SKU) (*ProductStock, error) {
	if productID == uuid.Nil {
		return nil, ErrInvalidProduct
	}
	return &ProductStock{
		productID: productID,
		sku:       sku,
	}, nil
}

func (p *ProductStock) ProductID() uuid.UUID { return p.productID }
func (p *ProductStock) SKU() SKU             { return p.sku }
func (p *ProductStock) Quantity() int        { return p.quantity }
func (p *ProductStock) Reserved() int        { return p.reserved }

func (p *ProductStock) Add(qty int) error {
	if qty <= 0 {
		return ErrNegativeQuantity
	}
	p.quantity += qty
	return nil
}

func (p *ProductStock) Reserve(qty int) error {
	if qty <= 0 {
		return ErrNegativeQuantity
	}
	if qty+p.reserved > p.quantity {
		return ErrInsufficientStock
	}
	p.reserved += qty
	return nil
}
