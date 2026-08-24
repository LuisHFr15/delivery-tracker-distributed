package messaging

import (
	"context"
	"encoding/json"
	"errors"
	"log"
	"os"
	"time"

	"github.com/LuisHFr15/delivery-tracker-distributed/internal/notifier/infrastructure/messaging/dtos"
	"github.com/segmentio/kafka-go"
)

type KafkaReadNotificationTopic struct {
	reader *kafka.Reader
}

func NewKafkaReadNotificationTopic() *KafkaReadNotificationTopic {
	broker := os.Getenv("KAFKA_BROKER")
	if broker == "" {
		broker = "localhost:9092"
	}
	return &KafkaReadNotificationTopic{
		reader: kafka.NewReader(kafka.ReaderConfig{
			Brokers:        []string{broker},
			Topic:          "processed-orders",
			CommitInterval: time.Second * 2,
		}),
	}
}

func (k *KafkaReadNotificationTopic) Read(ctx context.Context) (dtos.ProcessedOrderNotificationDTO, error) {
	msg, err := k.reader.ReadMessage(ctx)
	if err != nil {
		if errors.Is(err, context.Canceled) {
			log.Println("Shut down command received, closing kafka read notification client")
			return dtos.ProcessedOrderNotificationDTO{}, err
		}
		log.Printf("Error reading message from processed-orders topic in kafka: %v\n", err)
		return dtos.ProcessedOrderNotificationDTO{}, err
	}
	var dto dtos.ProcessedOrderNotificationDTO
	err = json.Unmarshal(msg.Value, &dto)
	if err != nil {
		log.Printf("Bad message to processed order notification: %v\n", err)
		return dtos.ProcessedOrderNotificationDTO{}, err
	}
	return dto, nil
}

func (k *KafkaReadNotificationTopic) Close() error {
	return k.reader.Close()
}
