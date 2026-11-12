package postgres

import (
	"context"
	"errors"

	"github.com/jackc/pgx/v5/pgconn"
	"github.com/merloot/marketplace/services/catalog/internal/domain"
)

type ProductRepo struct {
	provider ExecutorProvider
}

func NewProductRepo(p ExecutorProvider) *ProductRepo {
	return &ProductRepo{provider: p}
}

func (r *ProductRepo) Save(ctx context.Context, product domain.Product) error {
	price := product.Price()
	sku := product.SKU()
	executor := r.provider.Executor(ctx)
	_, err := executor.Exec(ctx,
		` INSERT INTO products (id, title, description, price_amount, currency, sku, category)
			VALUES ($1, $2, $3, $4, $5, $6, $7)`,
		product.ID(), product.Title(), product.Description(), price.Amount(), price.Currency(), sku.String(), product.Category(),
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
