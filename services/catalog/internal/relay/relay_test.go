package relay

import (
	"context"
	"errors"
	"io"
	"log/slog"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"

	"github.com/merloot/marketplace/pkg/postgres"
	"github.com/merloot/marketplace/services/catalog/internal/storage"
	"github.com/segmentio/kafka-go"
)

type MockTransactionManager struct {
	Err error
}

func (m *MockTransactionManager) WithTransaction(ctx context.Context, fn func(ctx context.Context, exec postgres.DBExecutor) error) error {
	if m.Err != nil {
		return m.Err
	}
	return fn(ctx, nil)
}

type MockKafkaWriter struct {
	WrittenMessages []kafka.Message
	Err             error
}

func (m *MockKafkaWriter) WriteMessages(ctx context.Context, msgs ...kafka.Message) error {
	if m.Err != nil {
		return m.Err
	}
	m.WrittenMessages = append(m.WrittenMessages, msgs...)
	return nil
}

type MockOutboxRepo struct {
	Messages          []storage.OutboxMessage
	ClaimErr          error
	MarkErr           error
	MarkedAsPublished []int
}

func (m *MockOutboxRepo) ClaimBatch(ctx context.Context, exec postgres.DBExecutor, limit int) ([]storage.OutboxMessage, error) {
	if m.ClaimErr != nil {
		return nil, m.ClaimErr
	}
	return m.Messages, nil
}

func (m *MockOutboxRepo) MarkAsPublished(ctx context.Context, exec postgres.DBExecutor, ids []int) error {
	if m.MarkErr != nil {
		return m.MarkErr
	}
	m.MarkedAsPublished = append(m.MarkedAsPublished, ids...)
	return nil
}

func (m *MockOutboxRepo) Save(ctx context.Context, aggregateID uuid.UUID, eventType string, payload []byte) error {
	return nil
}

func TestRelay_processBatch(t *testing.T) {
	validUUID := uuid.New()
	validMsg := storage.OutboxMessage{
		ID:          1,
		AggregateID: validUUID,
		EventType:   "ProductCreated",
		Payload:     []byte(`{"test":"data"}`),
	}

	createTestRelay := func(repo *MockOutboxRepo, writer *MockKafkaWriter, txManager *MockTransactionManager) *Relay {
		log := slog.New(slog.NewTextHandler(io.Discard, nil))
		return New(txManager, repo, writer, log)
	}

	t.Run("successful batch processing", func(t *testing.T) {
		repo := &MockOutboxRepo{Messages: []storage.OutboxMessage{validMsg}}
		writer := &MockKafkaWriter{}
		txManager := &MockTransactionManager{}
		r := createTestRelay(repo, writer, txManager)

		err := r.processBatch(context.Background())
		assert.NoError(t, err)

		assert.Len(t, writer.WrittenMessages, 1)
		assert.Equal(t, string(validMsg.Payload), string(writer.WrittenMessages[0].Value))
		assert.Equal(t, validUUID.String(), string(writer.WrittenMessages[0].Key))

		assert.Len(t, repo.MarkedAsPublished, 1)
		assert.Equal(t, 1, repo.MarkedAsPublished[0])
	})

	t.Run("kafka failure should not mark as published", func(t *testing.T) {
		repo := &MockOutboxRepo{Messages: []storage.OutboxMessage{validMsg}}
		writer := &MockKafkaWriter{Err: errors.New("kafka is down")}
		txManager := &MockTransactionManager{}
		r := createTestRelay(repo, writer, txManager)

		err := r.processBatch(context.Background())
		assert.Error(t, err)

		assert.Empty(t, writer.WrittenMessages)
		assert.Empty(t, repo.MarkedAsPublished)
	})

	t.Run("empty batch", func(t *testing.T) {
		repo := &MockOutboxRepo{Messages: []storage.OutboxMessage{}} // пусто
		writer := &MockKafkaWriter{}
		txManager := &MockTransactionManager{}
		r := createTestRelay(repo, writer, txManager)

		err := r.processBatch(context.Background())
		assert.NoError(t, err)

		assert.Empty(t, writer.WrittenMessages)
		assert.Empty(t, repo.MarkedAsPublished)
	})

	t.Run("claim batch error", func(t *testing.T) {
		repo := &MockOutboxRepo{ClaimErr: errors.New("db is down")}
		writer := &MockKafkaWriter{}
		txManager := &MockTransactionManager{}
		r := createTestRelay(repo, writer, txManager)

		err := r.processBatch(context.Background())
		assert.Error(t, err)
		assert.Empty(t, writer.WrittenMessages)
		assert.Empty(t, repo.MarkedAsPublished)
	})

	t.Run("transaction begin error", func(t *testing.T) {
		repo := &MockOutboxRepo{Messages: []storage.OutboxMessage{validMsg}}
		writer := &MockKafkaWriter{}
		txManager := &MockTransactionManager{Err: errors.New("cannot begin tx")}
		r := createTestRelay(repo, writer, txManager)

		err := r.processBatch(context.Background())
		assert.Error(t, err)
		assert.Empty(t, writer.WrittenMessages)
		assert.Empty(t, repo.MarkedAsPublished)
	})
}
