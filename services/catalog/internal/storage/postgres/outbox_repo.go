package postgres

import (
	"context"

	"github.com/google/uuid"
	pgbase "github.com/merloot/marketplace/pkg/postgres"
	"github.com/merloot/marketplace/services/catalog/internal/storage"
)

type OutboxRepo struct {
	provider ExecutorProvider
}

func NewOutboxRepo(p ExecutorProvider) *OutboxRepo {
	return &OutboxRepo{provider: p}
}

func (r *OutboxRepo) Save(ctx context.Context, aggregateID uuid.UUID, eventType string, payload []byte) error {
	executor := r.provider.Executor(ctx)
	_, err := executor.Exec(ctx,
		` INSERT INTO outbox (aggregate_id, event_type, payload)
			VALUES ($1, $2, $3)`,
		aggregateID, eventType, payload,
	)
	return err
}

func (r *OutboxRepo) ClaimBatch(ctx context.Context, exec pgbase.DBExecutor, limit int) ([]storage.OutboxMessage, error) {
	rows, err := exec.Query(ctx, `
		SELECT 
			id,
			aggregate_id,
			event_type,
			payload
		FROM
			outbox
		WHERE 
			processed_at IS NULL
		LIMIT 
			$1
		FOR UPDATE SKIP LOCKED
	`, limit)
	if err != nil {
		return nil, err
	}

	defer rows.Close()

	var msgs []storage.OutboxMessage
	for rows.Next() {
		var m storage.OutboxMessage
		if err := rows.Scan(&m.ID, &m.AggregateID, &m.EventType, &m.Payload); err != nil {
			return nil, err
		}
		msgs = append(msgs, m)
	}

	return msgs, rows.Err()
}

func (r *OutboxRepo) MarkAsPublished(ctx context.Context, exec pgbase.DBExecutor, ids []int) error {
	_, err := exec.Exec(ctx, `UPDATE outbox SET processed_at = NOW() where ID = ANY($1)`, ids)
	return err
}
