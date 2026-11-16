package consumer

import (
	"context"
	"encoding/json"
	"log/slog"

	"github.com/merloot/marketplace/pkg/events"
	"github.com/merloot/marketplace/pkg/kafka"
)

type MessageReader interface {
	FetchMessage(ctx context.Context) (kafka.Message, error)
	CommitMessages(ctx context.Context, msg ...kafka.Message) error
}

type UseCase interface {
	Execute(ctx context.Context, event events.ProductCreated) error
}

type ProductConsumer struct {
	reader MessageReader
	syncUC UseCase
	log    *slog.Logger
}

func NewProductConsumer(reader MessageReader, syncUC UseCase, log *slog.Logger) *ProductConsumer {
	return &ProductConsumer{reader: reader, syncUC: syncUC, log: log}
}

func (c *ProductConsumer) Run(ctx context.Context) error {
	c.log.Info("staring product consumer")
	for {
		msg, err := c.reader.FetchMessage(ctx)
		if err != nil {
			if ctx.Err() != nil {
				c.log.Info("consumer context canceled, shutting down")
				return nil
			}
			c.log.Error("failed to fetch message", "err", err)
			continue
		}
		var event events.ProductCreated
		if err := json.Unmarshal(msg.Value, &event); err != nil {
			c.log.Error("failed to unmarshal event, skipping", "err", err)
			_ = c.reader.CommitMessages(ctx, msg)
			continue
		}

		if err := c.syncUC.Execute(ctx, event); err != nil {
			c.log.Error("failed to sync product", "product_id", event.ProductID, "err", err)
			continue
		}

		if err := c.reader.CommitMessages(ctx, msg); err != nil {
			c.log.Error("failed to commit message", "err", err)
		}
	}
}
