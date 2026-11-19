package storage

import (
	"context"

	"github.com/google/uuid"

	"github.com/merloot/marketplace/services/inventory/internal/domain"
)

type ProductStockRepo interface {
	Save(ctx context.Context, productStock domain.ProductStock) error
	GetById(ctx context.Context, id uuid.UUID) (domain.ProductStock, error)
	TryReserve(ctx context.Context, productID uuid.UUID, quantity int) (bool, error)
	AddStock(ctx context.Context, productID uuid.UUID, quantity int) error
}

type UnitOfWork interface {
	Do(ctx context.Context, fn func(context.Context) error) error
}
