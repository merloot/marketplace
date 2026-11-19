package kafka

import "errors"

var (
	ErrNoBrokers = errors.New("kafka: brokers are required")
	ErrNoTopic   = errors.New("kafka: topic is required")
)
