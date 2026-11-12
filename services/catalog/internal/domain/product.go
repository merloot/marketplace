package domain

import (
	"strings"

	"github.com/google/uuid"
)


type Product struct {
	id          uuid.UUID
	title       string
	description string
	price       Money
	sku         SKU
	category    string
}

func NewProduct(title string, description string, price Money, sku SKU, category string) (*Product, error) {
	if strings.TrimSpace(title) == "" {
		return nil, ErrEmptyTitle
	}
	if price.IsNegative() {
		return nil, ErrNegativePrice
	}
	return &Product{
		id:          uuid.New(),
		title:       title,
		description: description,
		price:       price,
		sku:         sku,
		category:    category,
	}, nil
}

func (p *Product) ID() uuid.UUID       { return p.id }
func (p *Product) Category() string    { return p.category }
func (p *Product) Title() string       { return p.title }
func (p *Product) SKU() SKU            { return p.sku }
func (p *Product) Price() Money        { return p.price }
func (p *Product) Description() string { return p.description }

func (p *Product) ChangePrice(newPrice Money) error {
	if newPrice.IsNegative() {
		return ErrNegativePrice
	}
	p.price = newPrice
	return nil
}
