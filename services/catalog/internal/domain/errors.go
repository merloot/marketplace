package domain

import "errors"

var (
	ErrEmptyTitle      = errors.New("product title cannot be empty")
	ErrNegativePrice   = errors.New("product price cannot be negative")
	ErrInvalidSKU      = errors.New("product SKU must be uppercase alphanumeric")
	ErrInvalidCurrency = errors.New("product currency cannot be empty")
	ErrDuplicateSKU    = errors.New("product with this SKU already exists")
)
