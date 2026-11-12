package postgres

import (
	"context"
	"errors"

	"github.com/jackc/pgx/v5/pgxpool"
)

type UnitOfWork struct {
	pool *pgxpool.Pool
}

func NewUnitOfWork(pool *pgxpool.Pool) *UnitOfWork {
	return &UnitOfWork{pool: pool}
}

func (u *UnitOfWork) Do(ctx context.Context, fn func(context.Context) error) error {
	if _, ok := ExtractTx(ctx); ok {
		return errors.New("nested UnitOfWork is not supported")
	}
	tx, err := u.pool.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)

	ctxWithTx := InjectTx(ctx, tx)
	if err := fn(ctxWithTx); err != nil {
		return err
	}
	return tx.Commit(ctx)
}
