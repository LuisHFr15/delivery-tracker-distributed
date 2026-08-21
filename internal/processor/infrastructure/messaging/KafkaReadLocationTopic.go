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

type KafkaReadLocationTopic struct {
	reader *kafka.Reader
}

func NewKafkaReadLocationTopic() *KafkaReadLocationTopic {
	broker := os.Getenv("KAFKA_BROKER")
	if broker == "" {
		broker = "localhost:9092"
	}
	return &KafkaReadLocationTopic{
		reader: kafka.NewReader(kafka.ReaderConfig{
			Brokers:        []string{broker},
			Topic:          "location-events",
			CommitInterval: time.Second * 2,
		}),
	}
}

func (k *KafkaReadLocationTopic) Read(ctx context.Context) (dtos.LocationEventDTO, error) {
	msg, err := k.reader.ReadMessage(ctx)
	if err != nil {
		if errors.Is(err, context.Canceled) {
			log.Println("Shut down command received, closing kafka read location client")
			return dtos.LocationEventDTO{}, err
		}
		log.Printf("Error reading message from location topic kafka: %v\n", err)
		return dtos.LocationEventDTO{}, err
	}
	var dto dtos.LocationEventDTO
	err = json.Unmarshal(msg.Value, &dto)
	if err != nil {
		log.Printf("Bad message to location event: %v\n", err)
		return dtos.LocationEventDTO{}, err
	}
	return dto, nil
}

func (k *KafkaReadLocationTopic) Close() error {
	return k.reader.Close()
}
