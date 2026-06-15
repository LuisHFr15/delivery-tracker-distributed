package queues

import (
	"context"
	"encoding/json"
	"fmt"
	"main/internal/domain/models/events"
	"os"

	"github.com/segmentio/kafka-go"
)

type KafkaPublisher struct {
	orderWriter    *kafka.Writer
	locationWriter *kafka.Writer
}

func NewKafkaPublisher() *KafkaPublisher {
	broker := os.Getenv("KAFKA_BROKER")
	if broker == "" {
		broker = "localhost:9092"
	}

	return &KafkaPublisher{
		orderWriter: &kafka.Writer{
			Addr:     kafka.TCP(broker),
			Topic:    "order-events",
			Balancer: &kafka.LeastBytes{},
		},
		locationWriter: &kafka.Writer{
			Addr:     kafka.TCP(broker),
			Topic:    "location-events",
			Balancer: &kafka.LeastBytes{},
		},
	}
}

func (p *KafkaPublisher) PublishOrderEvent(ctx context.Context, event events.OrderEvent) error {
	value, err := json.Marshal(event)
	if err != nil {
		return fmt.Errorf("failed to marshal order event: %w", err)
	}

	return p.orderWriter.WriteMessages(ctx, kafka.Message{
		Key:   []byte(event.Id.String()),
		Value: value,
	})
}

func (p *KafkaPublisher) PublishLocationEvent(ctx context.Context, event events.LocationEvent) error {
	value, err := json.Marshal(event)
	if err != nil {
		return fmt.Errorf("failed to marshal location event: %w", err)
	}

	// context lib -> propagate cancellation signals, signal lifecycles
	return p.locationWriter.WriteMessages(ctx, kafka.Message{
		Key:   []byte(event.OrderId.String()),
		Value: value,
	})
}

func (p *KafkaPublisher) Close() {
	p.orderWriter.Close()
	p.locationWriter.Close()
}
