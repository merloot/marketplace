// package domain_test

// import (
// 	"testing"
// 	"github.com/google/uuid"

// 	"github.com/merloot/marketplace/services/inventory/internal/domain"
// 	"github.com/stretchr/testify/assert"
// )

// func TestNewProduct(t *testing.T) {
// 	t.Run("successful creation", func(t *testing.T) {
// 		uuid := uuid.New()
// 		p := domain.NewProductStock(uuid, "SKU-123")
// 		assert.Equal(t, uuid, p.ProductID())
// 		assert.Equal(t, "SKU-123", p.SKU())
// 	})

// 	t.Run("empty title", func(t *testing.T) {
// 		_, err := domain.NewProduct("", "Desc", validPrice, validSKU, "Category")
// 		assert.ErrorIs(t, err, domain.ErrEmptyTitle)
// 	})
// }

// func TestProduct_ChangePrice(t *testing.T) {
// 	validPrice, _ := domain.NewMoney(1000, "RUB")
// 	validSKU, _ := domain.NewSKU("SKU-123")
// 	p, _ := domain.NewProduct("Title", "Desc", validPrice, validSKU, "Category")

// 	t.Run("change to valid price", func(t *testing.T) {
// 		newPrice, _ := domain.NewMoney(2000, "RUB")
// 		err := p.ChangePrice(newPrice)
// 		assert.NoError(t, err)
// 		assert.Equal(t, int64(2000), p.Price().Amount())
// 	})

//		t.Run("change to negative price", func(t *testing.T) {
//			_, err := domain.NewMoney(-100, "RUB")
//			assert.ErrorIs(t, err, domain.ErrNegativePrice)
//		})
//	}
package domain_test

import (
	"testing"

	"github.com/google/uuid"

	"github.com/merloot/marketplace/services/inventory/internal/domain"
	"github.com/stretchr/testify/assert"
)

func TestNewProductStock(t *testing.T) {
	validSKU, _ := domain.NewSKU("SKU-123")
	t.Run("successful creation", func(t *testing.T) {
		uuid := uuid.New()
		p, err := domain.NewProductStock(uuid, validSKU)
		assert.NoError(t, err)
		assert.Equal(t, uuid, p.ProductID())
		assert.Equal(t, "SKU-123", p.SKU().String())
	})

	t.Run("empty product id", func(t *testing.T) {
		_, err := domain.NewProductStock(uuid.Nil, validSKU)
		assert.ErrorIs(t, err, domain.ErrInvalidProduct)

	})
}

func TestProductStock_Add(t *testing.T) {
	validSKU, _ := domain.NewSKU("SKU-123")
	validUUID := uuid.New()

	t.Run("successful add", func(t *testing.T) {
		p, _ := domain.NewProductStock(validUUID, validSKU)
		err := p.Add(10)
		assert.NoError(t, err)
		assert.Equal(t, 10, p.Quantity())
	})

	t.Run("negative quantity", func(t *testing.T) {
		p, _ := domain.NewProductStock(validUUID, validSKU)
		err := p.Add(-10)
		assert.ErrorIs(t, err, domain.ErrInvalidQuantity)
		assert.Equal(t, 0, p.Quantity())
	})

	t.Run("zero quantity", func(t *testing.T) {
		p, _ := domain.NewProductStock(validUUID, validSKU)
		err := p.Add(0)
		assert.ErrorIs(t, err, domain.ErrInvalidQuantity)
		assert.Equal(t, 0, p.Quantity())
	})
}

func TestProductStock_Reserve(t *testing.T) {
	validUUID := uuid.New()
	validSKU, _ := domain.NewSKU("SKU-123")
	t.Run("successful quantity", func(t *testing.T) {
		p, _ := domain.NewProductStock(validUUID, validSKU)
		_ = p.Add(10)
		err := p.Reserve(10)
		assert.NoError(t, err)
		assert.Equal(t, 10, p.Quantity())
		assert.Equal(t, 10, p.Reserved())
	})

	t.Run("negative quantity", func(t *testing.T) {
		p, _ := domain.NewProductStock(validUUID, validSKU)
		_ = p.Add(10)
		err := p.Reserve(-10)
		assert.ErrorIs(t, err, domain.ErrInvalidQuantity)
		assert.Equal(t, 10, p.Quantity())
		assert.Equal(t, 0, p.Reserved())
	})

	t.Run("insufficient stock", func(t *testing.T) {
		p, _ := domain.NewProductStock(validUUID, validSKU)
		_ = p.Add(10)
		err := p.Reserve(20)
		assert.ErrorIs(t, err, domain.ErrInsufficientStock)
		assert.Equal(t, 10, p.Quantity())
		assert.Equal(t, 0, p.Reserved())
	})
}
