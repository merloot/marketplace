package app_test

import (
	"context"
	"encoding/json"
	"testing"

	"github.com/merloot/marketplace/pkg/events"
	"github.com/merloot/marketplace/services/catalog/internal/app"
	"github.com/merloot/marketplace/services/catalog/internal/domain"
	"github.com/stretchr/testify/assert"
)

func TestCreateProductUseCase_Execute(t *testing.T) {
	validInput := app.CreateProductInput{
		Title:       "Test Product",
		Description: "Desc",
		Category:    "Cat",
		PriceAmount: 1000,
		Currency:    "RUB",
		SKU:         "SKU-123",
	}

	t.Run("success", func(t *testing.T) {
		uow := &MockUnitOfWork{}
		productRepo := &MockProductRepo{}
		outboxRepo := &MockOutboxRepo{}

		uc := app.NewCreateProductUseCase(uow, productRepo, outboxRepo)
		id, err := uc.Execute(context.Background(), validInput)

		assert.NoError(t, err)
		assert.NotNil(t, id)
		assert.Len(t, productRepo.SavedProducts, 1)
		assert.Len(t, outboxRepo.SavedEvents, 1)

		var evt events.ProductCreated
		err = json.Unmarshal(outboxRepo.SavedEvents[0].Payload, &evt)
		assert.NoError(t, err)
		assert.Equal(t, "Test Product", evt.Title)
		assert.Equal(t, int64(1000), evt.Price)
		assert.Equal(t, "SKU-123", evt.SKU)
	})

	t.Run("domain validation error", func(t *testing.T) {
		invalidInput := validInput
		invalidInput.Title = ""

		uow := &MockUnitOfWork{}
		productRepo := &MockProductRepo{}
		outboxRepo := &MockOutboxRepo{}
		uc := app.NewCreateProductUseCase(uow, productRepo, outboxRepo)

		_, err := uc.Execute(context.Background(), invalidInput)

		assert.ErrorIs(t, err, domain.ErrEmptyTitle)
		assert.Empty(t, productRepo.SavedProducts)
		assert.Empty(t, outboxRepo.SavedEvents)
	})

	t.Run("product repo error", func(t *testing.T) {
		uow := &MockUnitOfWork{}
		productRepo := &MockProductRepo{Err: domain.ErrDuplicateSKU}
		outboxRepo := &MockOutboxRepo{}
		uc := app.NewCreateProductUseCase(uow, productRepo, outboxRepo)

		_, err := uc.Execute(context.Background(), validInput)

		assert.ErrorIs(t, err, domain.ErrDuplicateSKU)
		assert.Empty(t, outboxRepo.SavedEvents)
	})
}
