package app_test

import (
	"context"

	"github.com/google/uuid"
	pgbase "github.com/merloot/marketplace/pkg/postgres"
	"github.com/merloot/marketplace/services/catalog/internal/domain"
	"github.com/merloot/marketplace/services/catalog/internal/storage"
)

type MockUnitOfWork struct {
	Err error
}

func (m *MockUnitOfWork) Do(ctx context.Context, fn func(context.Context) error) error {
	if m.Err != nil {
		return m.Err
	}
	return fn(ctx)
}

type MockProductRepo struct {
	SavedProducts []domain.Product
	Err           error
}

func (m *MockProductRepo) Save(ctx context.Context, p domain.Product) error {
	if m.Err != nil {
		return m.Err
	}
	m.SavedProducts = append(m.SavedProducts, p)
	return nil
}

type MockOutboxRepo struct {
	SavedEvents []storage.OutboxMessage
	Err         error
}

func (m *MockOutboxRepo) Save(ctx context.Context, aggregateID uuid.UUID, eventType string, payload []byte) error {
	if m.Err != nil {
		return m.Err
	}
	m.SavedEvents = append(m.SavedEvents, storage.OutboxMessage{
		AggregateID: aggregateID,
		EventType:   eventType,
		Payload:     payload,
	})
	return nil
}

func (m *MockOutboxRepo) ClaimBatch(ctx context.Context, exec pgbase.DBExecutor, limit int) ([]storage.OutboxMessage, error) {
	return nil, nil
}

func (m *MockOutboxRepo) MarkAsPublished(ctx context.Context, exec pgbase.DBExecutor, ids []int) error {
	return nil
}
