package queues

import (
	"context"
	"encoding/json"
	"fmt"
	"main/internal/ingester/app/dtos"
	"os"

	"github.com/google/uuid"
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

func (p *KafkaPublisher) PublishOrder(ctx context.Context, dto dtos.OrderEventDTO) error {
	value, err := json.Marshal(dto)
	if err != nil {
		return fmt.Errorf("failed to marshal order: %w", err)
	}

	return p.orderWriter.WriteMessages(ctx, kafka.Message{
		Key:   []byte(dto.Order.ID.String()),
		Value: value,
	})
}

// TODO: unitary tests and batch tests
// TODO: verificar como ficou a o application service e o handler, se aplicou corretamente os middlewares e routes
func (p *KafkaPublisher) PublishLocation(ctx context.Context, dto dtos.LocationEventDTO, orderId uuid.UUID) error {
	value, err := json.Marshal(dto)
	if err != nil {
		return fmt.Errorf("failed to marshal location event: %w", err)
	}

	return p.locationWriter.WriteMessages(ctx, kafka.Message{
		Key:   []byte(orderId.String()),
		Value: value,
	})
}

func (p *KafkaPublisher) Close() {
	p.orderWriter.Close()
	p.locationWriter.Close()
}
