package config

import "time"

type KafkaConfig struct {
	Brokers []string `env:"KAFKA_BROKERS" envSeparator:"," envRequired:"true"`
	Topic   string   `env:"KAFKA_TOPIC" envDefault:"catalog.events"`
	GroupID string   `env:"KAFKA_GROUP_ID" envDefault:"inventory-service"`
	CommitInterval time.Duration `env:"KAFKA_COMMIT_INTERVAL" envDefault:"0"`
}
