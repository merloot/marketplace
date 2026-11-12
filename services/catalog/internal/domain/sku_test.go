package domain_test

import (
	"testing"

	"github.com/merloot/marketplace/services/catalog/internal/domain"
	"github.com/stretchr/testify/assert"
)

func TestNewSKU(t *testing.T) {
	tests := []struct {
		name        string
		input       string
		expectedErr error
	}{
		{"valid SKU", "BOOK-123", nil},
		{"valid with numbers", "A1B2C3", nil},
		{"too short", "AB", domain.ErrInvalidSKU},
		{"lowercase letters", "book-123", domain.ErrInvalidSKU},
		{"spaces", "BOOK 123", domain.ErrInvalidSKU},
		{"empty", "", domain.ErrInvalidSKU},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			sku, err := domain.NewSKU(tt.input)
			if tt.expectedErr != nil {
				assert.ErrorIs(t, err, tt.expectedErr)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.input, sku.String())
			}
		})
	}
}