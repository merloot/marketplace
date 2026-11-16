package postgres

import (
	"context"
	"errors"

	"github.com/google/uuid"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/merloot/marketplace/services/inventory/internal/domain"
	"github.com/merloot/marketplace/services/inventory/internal/storage"
)

type ProductStockRepo struct {
	provider ExecutorProvider
}

func NewProductStockRepo(p ExecutorProvider) *ProductStockRepo {
	return &ProductStockRepo{provider: p}
}

func (r *ProductStockRepo) Save(ctx context.Context, productStock domain.ProductStock) error {
	executor := r.provider.Executor(ctx)
	_, err := executor.Exec(ctx,
		` INSERT INTO product_stocks (product_id, sku, quantity, reserved)
			VALUES ($1, $2, $3, $4)`,
		productStock.ProductID(), productStock.SKU(), productStock.Quantity(), productStock.Reserved(),
	)
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == "23505" {
			return domain.ErrDuplicateSKU
		}
		return err
	}
	return err
}

func (r *ProductStockRepo) GetById(ctx context.Context, id uuid.UUID) (domain.ProductStock, error) {
	var productID uuid.UUID
	var sku string
	var quantity, reserved int
	row := r.provider.Executor(ctx).QueryRow(ctx, `
		SELECT 
			product_id, 
			sku, 
			quantity, 
			reserved 
		FROM 
			product_stocks 
		WHERE 
			product_id = $1 
		FOR UPDATE`, id)

	if err := row.Scan(&productID, &sku, &quantity, &reserved); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return domain.ProductStock{}, domain.ErrProductNotFound
		}
		return domain.ProductStock{}, err
	}
	skuVO, _ := domain.NewSKU(sku)
	return *domain.RehydrateProductStock(productID, skuVO, quantity, reserved), nil
}

func (r *ProductStockRepo) TryReserve(ctx context.Context, productID uuid.UUID, quantity int) (bool, error) {
	var id uuid.UUID
	err := r.provider.Executor(ctx).QueryRow(ctx, `
	UPDATE 
		product_stocks 
	SET 
		reserved = reserved + $1
	WHERE
		product_id = $2
		AND quantity - reserved >= $1
	RETURNING product_id
	`, quantity, productID).Scan(&id)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return false, nil
		}
		return false, err
	}
	return true, nil
}

func (r *ProductStockRepo) AddStock(ctx context.Context, productID uuid.UUID, quantity int) error {
	if quantity <= 0 {
		return domain.ErrInvalidQuantity
	}

	tag, err := r.provider.Executor(ctx).Exec(ctx, `
		UPDATE product_stocks
		SET quantity = quantity + $1
		WHERE product_id = $2
	`, quantity, productID)
	if err != nil {
		return err
	}

	if tag.RowsAffected() == 0 {
		return domain.ErrProductNotFound
	}
	return nil
}

var _ storage.ProductStockRepo = (*ProductStockRepo)(nil)
