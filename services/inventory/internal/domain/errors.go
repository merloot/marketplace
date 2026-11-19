package domain

import "errors"

var (
	ErrInvalidProduct = errors.New("product id must by uuid")
	ErrInvalidSKU     = errors.New("product SKU must be uppercase alphanumeric")

	ErrInvalidQuantity  = errors.New("product quantity must be greater than zero")
	ErrInsufficientStock = errors.New("insufficient stock")

	ErrProductNotFound = errors.New("product not found")
	ErrDuplicateSKU = errors.New("product with this SKU already exists")
)
