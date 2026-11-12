package domain_test

import (
	"testing"

	"github.com/merloot/marketplace/services/catalog/internal/domain"
	"github.com/stretchr/testify/assert"
)

func TestNewProduct(t *testing.T) {
	validPrice, _ := domain.NewMoney(1000, "RUB")
	validSKU, _ := domain.NewSKU("SKU-123")

	t.Run("successful creation", func(t *testing.T) {
		p, err := domain.NewProduct("Title", "Desc", validPrice, validSKU, "Category")
		assert.NoError(t, err)
		assert.NotNil(t, p.ID())
		assert.Equal(t, "Title", p.Title())
	})

	t.Run("empty title", func(t *testing.T) {
		_, err := domain.NewProduct("", "Desc", validPrice, validSKU, "Category")
		assert.ErrorIs(t, err, domain.ErrEmptyTitle)
	})
}

func TestProduct_ChangePrice(t *testing.T) {
	validPrice, _ := domain.NewMoney(1000, "RUB")
	validSKU, _ := domain.NewSKU("SKU-123")
	p, _ := domain.NewProduct("Title", "Desc", validPrice, validSKU, "Category")

	t.Run("change to valid price", func(t *testing.T) {
		newPrice, _ := domain.NewMoney(2000, "RUB")
		err := p.ChangePrice(newPrice)
		assert.NoError(t, err)
		assert.Equal(t, int64(2000), p.Price().Amount())
	})

	t.Run("change to negative price", func(t *testing.T) {
		_, err := domain.NewMoney(-100, "RUB")
		assert.ErrorIs(t, err, domain.ErrNegativePrice)
	})
}
