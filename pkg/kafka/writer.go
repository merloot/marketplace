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

func toKafkaMessages(msg []Message) []kafka.Message {
	kafkaMsgs := make([]kafka.Message, 0, len(msg))

	for _, m := range msg {
		kafkaHeaders := make([]kafka.Header, 0, len(m.Headers))

		for _, h := range m.Headers {
			kafkaHeaders = append(kafkaHeaders, kafka.Header{Key: h.Key, Value: h.Value})
		}

		kafkaMsgs = append(kafkaMsgs, kafka.Message{
			Key:     m.Key,
			Value:   m.Value,
			Headers: kafkaHeaders,
		})
	}
	return kafkaMsgs
}

func (w *Writer) WriteMessages(ctx context.Context, msgs ...Message) error {
	kafkaMsgs := toKafkaMessages(msgs)
	if err := w.writer.WriteMessages(ctx, kafkaMsgs...); err != nil {
		w.log.Error("failed to write messages", "error", err)
		return err
	}
	return nil
}

func (w *Writer) Close() error {
	return w.writer.Close()
}
