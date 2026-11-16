package config

type KafkaConfig struct {
	Brokers []string `env:"KAFKA_BROKERS" envSeparator:"," envRequired:"true"`
	Topic   string   `env:"KAFKA_TOPIC" envDefault:"catalog.events"`
}
