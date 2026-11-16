package domain

import "errors"

var (
	ErrInvalidProduct = errors.New("product id must by uuid")
	ErrInvalidSKU     = errors.New("product SKU must be uppercase alphanumeric")

	ErrNegativeQuantity  = errors.New("product quantity cannot by negative")
	ErrInsufficientStock = errors.New("insufficient stock")
)
