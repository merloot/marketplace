package app

import (
	"context"
	"encoding/json"

	"github.com/google/uuid"
	"github.com/merloot/marketplace/pkg/events"
	"github.com/merloot/marketplace/services/catalog/internal/domain"
	"github.com/merloot/marketplace/services/catalog/internal/storage"
)

type CreateProductInput struct {
	Title, Description, Category, Currency, SKU string
	PriceAmount                                 int64
}

type CreateProductUseCase struct {
	uow      storage.UnitOfWork
	products storage.ProductRepo
	outbox   storage.OutboxRepo
}

func NewCreateProductUseCase(uow storage.UnitOfWork, products storage.ProductRepo, outbox storage.OutboxRepo) *CreateProductUseCase {
	return &CreateProductUseCase{uow: uow, products: products, outbox: outbox}
}

func (uc *CreateProductUseCase) Execute(ctx context.Context, input CreateProductInput) (uuid.UUID, error) {
	price, err := domain.NewMoney(input.PriceAmount, input.Currency)
	if err != nil {
		return uuid.Nil, err
	}
	sku, err := domain.NewSKU(input.SKU)
	if err != nil {
		return uuid.Nil, err
	}

	product, err := domain.NewProduct(input.Title, input.Description, price, sku, input.Category)
	if err != nil {
		return uuid.Nil, err
	}

	event := events.ProductCreated{
		ProductID: product.ID(), Title: product.Title(), Description: product.Description(),
		Price: price.Amount(), SKU: sku.String(),
	}
	payload, err := json.Marshal(event)
	if err != nil {
		return uuid.Nil, err
	}

	err = uc.uow.Do(ctx, func(ctx context.Context) error {
		if err := uc.products.Save(ctx, *product); err != nil {
			return err
		}
		return uc.outbox.Save(ctx, product.ID(), events.ProductCreatedEvent, payload)
	})
	if err != nil {
		return uuid.Nil, err
	}
	return product.ID(), nil
}
