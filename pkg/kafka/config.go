package kafka

import "time"

type WriterConfig struct {
	Brokers      []string
	Topic        string
	RequiredAcks int
	Async        bool
	BatchSize    int
	BatchTimeout time.Duration
}

type ReaderConfig struct {
	Brokers        []string
	Topic          string
	GroupID        string
	MinBytes       int
	MaxBytes       int
	CommitInterval time.Duration
}
