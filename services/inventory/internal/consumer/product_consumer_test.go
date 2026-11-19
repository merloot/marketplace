package consumer_test

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"log/slog"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"

	"github.com/merloot/marketplace/pkg/events"
	"github.com/merloot/marketplace/pkg/kafka"
	"github.com/merloot/marketplace/services/inventory/internal/consumer"
)


type MockReader struct {
	Messages  []kafka.Message
	FetchErr  error
	CommitErr error
	Committed []kafka.Message
}

func (m *MockReader) FetchMessage(ctx context.Context) (kafka.Message, error) {
	if m.FetchErr != nil {
		return kafka.Message{}, m.FetchErr
	}
	if len(m.Messages) == 0 {
		<-ctx.Done()
		return kafka.Message{}, ctx.Err()
	}
	msg := m.Messages[0]
	m.Messages = m.Messages[1:]
	return msg, nil
}

func (m *MockReader) CommitMessages(ctx context.Context, msgs ...kafka.Message) error {
	if m.CommitErr != nil {
		return m.CommitErr
	}
	m.Committed = append(m.Committed, msgs...)
	return nil
}

type MockUseCase struct {
	Err error
}

func (m *MockUseCase) Execute(ctx context.Context, event events.ProductCreated) error {
	return m.Err
}


func TestProductConsumer_Run(t *testing.T) {
	silentLog := slog.New(slog.NewTextHandler(io.Discard, nil))

	validEvent := events.ProductCreated{ProductID: uuid.New(), SKU: "SKU-123"}
	validPayload, _ := json.Marshal(validEvent)
	validMsg := kafka.Message{Key: []byte("1"), Value: validPayload}

	t.Run("successful processing and commit", func(t *testing.T) {
		reader := &MockReader{Messages: []kafka.Message{validMsg}}
		uc := &MockUseCase{}
		c := consumer.NewProductConsumer(reader, uc, silentLog)

		ctx, cancel := context.WithTimeout(context.Background(), 50*time.Millisecond)
		defer cancel()
		_ = c.Run(ctx)

		assert.Len(t, reader.Committed, 1)
	})

	t.Run("invalid JSON should be committed to avoid poison pill", func(t *testing.T) {
		badMsg := kafka.Message{Value: []byte(`{invalid json`)}
		reader := &MockReader{Messages: []kafka.Message{badMsg}}
		uc := &MockUseCase{}
		c := consumer.NewProductConsumer(reader, uc, silentLog)

		ctx, cancel := context.WithTimeout(context.Background(), 50*time.Millisecond)
		defer cancel()
		_ = c.Run(ctx)

		assert.Len(t, reader.Committed, 1)
		assert.Equal(t, badMsg.Value, reader.Committed[0].Value)
	})

	t.Run("use case error should not commit", func(t *testing.T) {
		reader := &MockReader{Messages: []kafka.Message{validMsg}}
		uc := &MockUseCase{Err: errors.New("db error")}
		c := consumer.NewProductConsumer(reader, uc, silentLog)

		ctx, cancel := context.WithTimeout(context.Background(), 50*time.Millisecond)
		defer cancel()
		_ = c.Run(ctx)

		assert.Empty(t, reader.Committed)
	})

	t.Run("fetch error should not crash loop", func(t *testing.T) {
		reader := &MockReader{FetchErr: errors.New("kafka down")}
		uc := &MockUseCase{}
		c := consumer.NewProductConsumer(reader, uc, silentLog)

		ctx, cancel := context.WithTimeout(context.Background(), 50*time.Millisecond)
		defer cancel()
		_ = c.Run(ctx)

		assert.Empty(t, reader.Committed)
	})

	t.Run("commit error should not crash loop", func(t *testing.T) {
		reader := &MockReader{Messages: []kafka.Message{validMsg}, CommitErr: errors.New("commit failed")}
		uc := &MockUseCase{}
		c := consumer.NewProductConsumer(reader, uc, silentLog)

		ctx, cancel := context.WithTimeout(context.Background(), 50*time.Millisecond)
		defer cancel()
		_ = c.Run(ctx)

		assert.Empty(t, reader.Committed)
	})
}
