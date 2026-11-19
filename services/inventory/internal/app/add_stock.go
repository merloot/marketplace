package app

import (
	"context"

	"github.com/google/uuid"
	"github.com/merloot/marketplace/services/inventory/internal/domain"
	"github.com/merloot/marketplace/services/inventory/internal/storage"
)

type AddStockInput struct {
	ProductID uuid.UUID
	Quantity  int
}
type AddStockUseCase struct {
	product storage.ProductStockRepo
}

func NewAddStockUseCase(product storage.ProductStockRepo) *AddStockUseCase {
	return &AddStockUseCase{
		product: product,
	}
}

func (uc *AddStockUseCase) Execute(ctx context.Context, input AddStockInput) (uuid.UUID, error) {
	productID, err := uuid.Parse(input.ProductID.String())
	if err != nil || productID == uuid.Nil {
		return uuid.Nil, domain.ErrInvalidProduct
	}

	if input.Quantity < 0 {
		return uuid.Nil, domain.ErrInvalidQuantity
	}

	if err := uc.product.AddStock(ctx, productID, input.Quantity); err != nil {
		return uuid.Nil, err
	}
	return productID, nil
}
