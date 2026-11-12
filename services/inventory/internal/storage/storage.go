package storage

import (
	"context"
)

type UnitOfWork interface {
	Do(ctx context.Context, fn func(context.Context) error) error
}
