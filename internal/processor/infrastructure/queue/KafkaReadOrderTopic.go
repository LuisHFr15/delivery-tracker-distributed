package infrastructure

import (
	"context"
	"encoding/json"
	"errors"
	"log"
	"os"
	"time"

	"github.com/LuisHFr15/delivery-tracker-distributed/internal/processor/app/dtos"
	"github.com/segmentio/kafka-go"
)

type KafkaReadOrderTopic struct {
	reader *kafka.Reader
}

func NewKafkaReadOrderTopic() *KafkaReadOrderTopic {
	broker := os.Getenv("KAFKA_BROKER")
	if broker == "" {
		broker = "localhost:9092"
	}
	return &KafkaReadOrderTopic{
		reader: kafka.NewReader(kafka.ReaderConfig{
			Brokers:        []string{broker},
			Topic:          "location-events",
			CommitInterval: time.Second * 2,
		}),
	}
}

func (k *KafkaReadOrderTopic) Read(ctx context.Context) (dtos.OrderEventDTO, error) {
	msg, err := k.reader.ReadMessage(ctx)
	if err != nil {
		if errors.Is(err, context.Canceled) {
			log.Println("Shut down command received, closing kafka read order client")
			return dtos.OrderEventDTO{}, err
		}
		log.Printf("Error reading message from order topic in kafka: %v\n", err)
		return dtos.OrderEventDTO{}, err
	}
	var dto dtos.OrderEventDTO
	err = json.Unmarshal(msg.Value, &dto)
	if err != nil {
		log.Printf("Bad message to order event: %v\n", err)
		return dtos.OrderEventDTO{}, err
	}

	return dto, nil
}

func (k *KafkaReadOrderTopic) Close() error {
	return k.reader.Close()
}
