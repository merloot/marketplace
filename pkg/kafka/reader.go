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

func (r *Reader) FetchMessage(ctx context.Context) (Message, error) {
	kafkaMsg, err := r.reader.FetchMessage(ctx)
	if err != nil {
		return Message{}, err
	}
	headers := make([]Header, 0, len(kafkaMsg.Headers))

	for _, h := range kafkaMsg.Headers {
		headers = append(headers, Header{Key: h.Key, Value: h.Value})
	}
	return Message{
		Key:       kafkaMsg.Key,
		Value:     kafkaMsg.Value,
		Headers:   headers,
		partition: kafkaMsg.Partition,
		offset:    kafkaMsg.Offset,
	}, nil
}

func (r *Reader) CommitMessages(ctx context.Context, msgs ...Message) error {
	kafkaMsg := make([]kafka.Message, 0, len(msgs))
	for _, m := range msgs {
		kafkaMsg = append(kafkaMsg, kafka.Message{
			Topic:     r.reader.Config().Topic,
			Partition: m.partition,
			Offset:    m.offset,
			Key:       m.Key,
			Value:     m.Value,
		})
	}
	return r.reader.CommitMessages(ctx, kafkaMsg...)
}

func (r *Reader) Close() error {
	return r.reader.Close()
}
