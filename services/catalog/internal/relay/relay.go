package relay

import (
	"context"
	"log/slog"
	"time"

	pgbase "github.com/merloot/marketplace/pkg/postgres"
	"github.com/merloot/marketplace/services/catalog/internal/storage"
	"github.com/segmentio/kafka-go"
)

type KafkaWriter interface {
	WriteMessages(context.Context, ...kafka.Message) error
}

type TransactionManager interface {
	WithTransaction(ctx context.Context, fn func(ctx context.Context, exec pgbase.DBExecutor) error) error
}

type Relay struct {
	db         TransactionManager
	outboxRepo storage.OutboxRepo
	writer     KafkaWriter
	log        *slog.Logger
}

func New(db TransactionManager, outboxRepo storage.OutboxRepo, writer KafkaWriter, log *slog.Logger) *Relay {
	return &Relay{
		db:         db,
		outboxRepo: outboxRepo,
		writer:     writer,
		log:        log,
	}
}

func (r *Relay) Run(ctx context.Context) error {
	ticker := time.NewTicker(1 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return nil
		case <-ticker.C:
			if err := r.processBatch(ctx); err != nil {
				r.log.Error("failed to process outbox batch", "err", err)
			}
		}
	}
}

func (r *Relay) processBatch(ctx context.Context) error {
	return r.db.WithTransaction(ctx, func(ctx context.Context, exec pgbase.DBExecutor) error {
		msgs, err := r.outboxRepo.ClaimBatch(ctx, exec, 100)
		if err != nil || len(msgs) == 0 {
			return err
		}

		kafkaMsgs := mapToKafka(msgs)
		if err := r.writer.WriteMessages(ctx, kafkaMsgs...); err != nil {
			return err
		}
		ids := extractIDs(msgs)
		return r.outboxRepo.MarkAsPublished(ctx, exec, ids)
	})
}

func mapToKafka(msgs []storage.OutboxMessage) []kafka.Message {
	kafkaMsgs := make([]kafka.Message, 0, len(msgs))

	for _, m := range msgs {
		kafkaMsgs = append(kafkaMsgs, kafka.Message{
			Key:   []byte(m.AggregateID.String()),
			Value: m.Payload,
			Headers: []kafka.Header{
				{Key: "event_type", Value: []byte(m.EventType)},
			},
		})
	}
	return kafkaMsgs
}

func extractIDs(msgs []storage.OutboxMessage) []int {
	ids := make([]int, 0, len(msgs))
	for _, m := range msgs {
		ids = append(ids, m.ID)
	}
	return ids
}
