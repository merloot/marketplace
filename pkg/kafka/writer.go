package kafka

import (
	"context"
	"log/slog"

	"github.com/segmentio/kafka-go"
)

type Writer struct {
	writer *kafka.Writer
	log    *slog.Logger
}

func NewWriter(cfg WriterConfig, log *slog.Logger) (*Writer, error) {
	if len(cfg.Brokers) == 0 {
		return nil, ErrNoBrokers
	}
	if cfg.Topic == "" {
		return nil, ErrNoTopic
	}
	if log == nil {
		log = slog.Default()
	}
	w := kafka.Writer{
		Addr:         kafka.TCP(cfg.Brokers...),
		Topic:        cfg.Topic,
		Balancer:     &kafka.LeastBytes{},
		RequiredAcks: kafka.RequiredAcks(cfg.RequiredAcks),
		Async:        cfg.Async,
		BatchSize:    cfg.BatchSize,
		BatchTimeout: cfg.BatchTimeout,
	}

	return &Writer{writer: &w, log: log}, nil
}

func (w *Writer) WriteMessages(ctx context.Context, msgs ...kafka.Message) error {
	if err := w.writer.WriteMessages(ctx, msgs...); err != nil {
		w.log.Error("failed to write messages", "error", err)
		return err
	}
	return nil
}

func (w *Writer) Close() error {
	return w.writer.Close()
}
