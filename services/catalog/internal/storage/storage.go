package storage

import (
	"context"

	"github.com/google/uuid"

	pgbase "github.com/merloot/marketplace/pkg/postgres"
	"github.com/merloot/marketplace/services/catalog/internal/domain"
)

type OutboxMessage struct {
	ID          int
	AggregateID uuid.UUID
	EventType   string
	Payload     []byte
}

type ProductRepo interface {
	Save(ctx context.Context, product domain.Product) error
}

type OutboxRepo interface {
	Save(ctx context.Context, aggregateID uuid.UUID, eventType string, payload []byte) error
	ClaimBatch(ctx context.Context, exec pgbase.DBExecutor, limit int) ([]OutboxMessage, error)
	MarkAsPublished(ctx context.Context, exec pgbase.DBExecutor, ids []int) error
}

type UnitOfWork interface {
	Do(ctx context.Context, fn func(context.Context) error) error
}
