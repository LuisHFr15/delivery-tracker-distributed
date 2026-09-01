package messaging

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

type KafkaReadProcessedOrders struct {
	reader *kafka.Reader
}

func NewKafkaReadProcessedOrders() *KafkaReadProcessedOrders {
	broker := os.Getenv("KAFKA_BROKER")
	if broker == "" {
		broker = "localhost:9092"
	}
	return &KafkaReadProcessedOrders{
		reader: kafka.NewReader(kafka.ReaderConfig{
			Brokers:        []string{broker},
			Topic:          "processed-orders",
			CommitInterval: time.Second * 2,
		}),
	}
}

func (k *KafkaReadProcessedOrders) Read(ctx context.Context) (dtos.ProcessedOrderDTO, error) {
	msg, err := k.reader.ReadMessage(ctx)
	if err != nil {
		if errors.Is(err, context.Canceled) {
			log.Println("Shut down command received, closing kafka processed-orders client")
			return dtos.ProcessedOrderDTO{}, err
		}
		log.Printf("Error reading message from processed-orders topic in kafka: %v\n", err)
		return dtos.ProcessedOrderDTO{}, err
	}
	var dto dtos.ProcessedOrderDTO
	if err := json.Unmarshal(msg.Value, &dto); err != nil {
		log.Printf("Bad processed order message: %v\n", err)
		return dtos.ProcessedOrderDTO{}, err
	}
	return dto, nil
}

func (k *KafkaReadProcessedOrders) Close() error {
	return k.reader.Close()
}
