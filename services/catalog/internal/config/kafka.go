package config

type KafkaConfig struct {
	Brokers []string `env:"KAFKA_BROKERS required"`
	Topic   string   `env:"KAFKA_TOPIC envDefault:"catalog.events"`
}
