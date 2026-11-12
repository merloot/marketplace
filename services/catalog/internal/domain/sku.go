package domain

import (
	"regexp"
)

var (
	skuPattern = regexp.MustCompile(`^[A-Z0-9-]{3,32}$`)
)

type SKU struct {
	value string
}

func NewSKU(s string) (SKU, error) {
	if !skuPattern.MatchString(s) {
		return SKU{}, ErrInvalidSKU
	}
	return SKU{value: s}, nil
}

func (s *SKU) String() string { return s.value }
