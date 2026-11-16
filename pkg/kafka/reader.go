package kafka

import (
	"context"
	"log/slog"

	"github.com/segmentio/kafka-go"
)

type Reader struct {
	reader *kafka.Reader
	log    *slog.Logger
}

func NewReader(cfg ReaderConfig, log *slog.Logger) (*Reader, error) {
	if len(cfg.Brokers) == 0 {
		return nil, ErrNoBrokers
	}
	if cfg.Topic == "" {
		return nil, ErrNoTopic
	}
	if log == nil {
		log = slog.Default()
	}
	r := kafka.NewReader(kafka.ReaderConfig{
		Brokers:        cfg.Brokers,
		Topic:          cfg.Topic,
		GroupID:        cfg.GroupID,
		MinBytes:       cfg.MinBytes,
		MaxBytes:       cfg.MaxBytes,
		CommitInterval: cfg.CommitInterval,
	})
	return &Reader{reader: r, log: log}, nil
}

func (r *Reader) FetchMessage(ctx context.Context) (kafka.Message, error) {
	return r.reader.FetchMessage(ctx)
}

func (r *Reader) CommitMessages(ctx context.Context, msgs ...kafka.Message) error {
	return r.reader.CommitMessages(ctx, msgs...)
}

func (r *Reader) Close() error {
	return r.reader.Close()
}
