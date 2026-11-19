package app_test

import (
	"context"
	"errors"
	"io"
	"log/slog"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"

	"github.com/merloot/marketplace/pkg/events"
	"github.com/merloot/marketplace/services/inventory/internal/app"
	"github.com/merloot/marketplace/services/inventory/internal/domain"
)

type MockProductStockRepo struct {
	SavedStocks []domain.ProductStock
	Err         error
}

func (m *MockProductStockRepo) Save(ctx context.Context, stock domain.ProductStock) error {
	if m.Err != nil {
		return m.Err
	}
	m.SavedStocks = append(m.SavedStocks, stock)
	return nil
}

func (m *MockProductStockRepo) GetById(ctx context.Context, id uuid.UUID) (domain.ProductStock, error) {
	return domain.ProductStock{}, nil
}
func (m *MockProductStockRepo) TryReserve(ctx context.Context, productID uuid.UUID, quantity int) (bool, error) {
	return false, nil
}
func (m *MockProductStockRepo) AddStock(ctx context.Context, productID uuid.UUID, quantity int) error {
	return nil
}

func TestSyncProductUseCase_Execute(t *testing.T) {
	validEvent := events.ProductCreated{
		ProductID: uuid.New(),
		Title:     "Test Product",
		SKU:       "SKU-123",
		Price:     1000,
	}

	silentLog := slog.New(slog.NewTextHandler(io.Discard, nil))

	t.Run("successful sync", func(t *testing.T) {
		repo := &MockProductStockRepo{}
		uc := app.NewSyncProductUseCase(repo, silentLog)

		err := uc.Execute(context.Background(), validEvent)

		assert.NoError(t, err)
		assert.Len(t, repo.SavedStocks, 1)
		assert.Equal(t, validEvent.ProductID, repo.SavedStocks[0].ProductID())
		assert.Equal(t, validEvent.SKU, repo.SavedStocks[0].SKU().String())
		assert.Equal(t, 0, repo.SavedStocks[0].Quantity())
	})

	t.Run("invalid SKU in event", func(t *testing.T) {
		repo := &MockProductStockRepo{}
		uc := app.NewSyncProductUseCase(repo, silentLog)

		invalidEvent := validEvent
		invalidEvent.SKU = "invalid sku with spaces"

		err := uc.Execute(context.Background(), invalidEvent)

		assert.ErrorIs(t, err, domain.ErrInvalidSKU)
		assert.Empty(t, repo.SavedStocks)
	})

	t.Run("idempotency: duplicate SKU is ignored", func(t *testing.T) {
		repo := &MockProductStockRepo{Err: domain.ErrDuplicateSKU}
		uc := app.NewSyncProductUseCase(repo, silentLog)

		err := uc.Execute(context.Background(), validEvent)

		assert.NoError(t, err)
		assert.Empty(t, repo.SavedStocks)
	})

	t.Run("db error is propagated", func(t *testing.T) {
		dbErr := errors.New("connection refused")
		repo := &MockProductStockRepo{Err: dbErr}
		uc := app.NewSyncProductUseCase(repo, silentLog)

		err := uc.Execute(context.Background(), validEvent)

		assert.ErrorIs(t, err, dbErr)
		assert.Empty(t, repo.SavedStocks)
	})
}
