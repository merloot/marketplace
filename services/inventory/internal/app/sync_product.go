package app

import (
	"context"
	"errors"
	"log/slog"

	"github.com/merloot/marketplace/pkg/events"
	"github.com/merloot/marketplace/services/inventory/internal/domain"
	"github.com/merloot/marketplace/services/inventory/internal/storage"
)

type SyncProductUseCase struct {
	repo storage.ProductStockRepo
	log  *slog.Logger
}

func NewSyncProductUseCase(repo storage.ProductStockRepo, log *slog.Logger) *SyncProductUseCase {
	return &SyncProductUseCase{repo: repo, log: log}
}

func (uc *SyncProductUseCase) Execute(ctx context.Context, event events.ProductCreated) error {
	sku, err := domain.NewSKU(event.SKU)
	if err != nil {
		uc.log.Error("invalid SKU in event", "sku", event.SKU, "err", err)
		return err
	}
	stock, err := domain.NewProductStock(event.ProductID, sku)
	if err != nil {
		return err
	}

	err = uc.repo.Save(ctx, *stock)
	if err != nil {
		if errors.Is(err, domain.ErrDuplicateSKU) {
			uc.log.Warn("product stock already exists, skipping", "product_id", event.ProductID)
			return nil
		}
		return err
	}
	uc.log.Info("product stock created", "product_id", event.ProductID)
	return nil
}
