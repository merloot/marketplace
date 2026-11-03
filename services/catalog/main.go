package main

import (
	"context"
	"fmt"
	"log"

	"github.com/segmentio/kafka-go"
)

func main() {
	w := &kafka.Writer{
		Addr:  kafka.TCP("localhost:9093"),
		Topic: "catalog.events",
	}
	defer w.Close()

	err := w.WriteMessages(context.Background(),
		kafka.Message{
			Key:   []byte("test"),
			Value: []byte(`{"hello":"world"}`)},
	)

	if err != nil {
		log.Fatal("failed to write:", err)
	}

	fmt.Println("message written to catalog.events")
}
