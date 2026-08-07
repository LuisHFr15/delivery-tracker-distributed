package queues

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"sync"
	"time"

	"github.com/LuisHFr15/delivery-tracker-distributed/internal/ingester/app/dtos"

	"github.com/segmentio/kafka-go"
)

type KafkaPublisher struct {
	orderWriter    *kafka.Writer
	locationWriter *kafka.Writer
	orderBuffer    chan *dtos.OrderEventDTO
	locationBuffer chan *dtos.LocationEventDTO
	wg             sync.WaitGroup
}

func NewKafkaPublisher() *KafkaPublisher {
	broker := os.Getenv("KAFKA_BROKER")
	if broker == "" {
		broker = "localhost:9092"
	}

	return &KafkaPublisher{
		orderWriter: &kafka.Writer{
			Addr:                   kafka.TCP(broker),
			Topic:                  "order-events",
			Balancer:               &kafka.LeastBytes{},
			AllowAutoTopicCreation: true,
		},
		locationWriter: &kafka.Writer{
			Addr:                   kafka.TCP(broker),
			Topic:                  "location-events",
			Balancer:               &kafka.LeastBytes{},
			AllowAutoTopicCreation: true,
		},
		orderBuffer:    make(chan *dtos.OrderEventDTO, 100),
		locationBuffer: make(chan *dtos.LocationEventDTO, 100),
	}
}

func (p *KafkaPublisher) PublishOrder(dto dtos.OrderEventDTO) {
	p.orderBuffer <- &dto
}

func (p *KafkaPublisher) PublishLocation(dto dtos.LocationEventDTO) {
	p.locationBuffer <- &dto
}

func (p *KafkaPublisher) Start(errCh chan error) {
	p.wg.Add(2)
	go func() {
		defer p.wg.Done()
		for o := range p.orderBuffer {
			value, err := json.Marshal(*o)
			if err != nil {
				errCh <- fmt.Errorf("failed to marshal order: %w", err)
			}

			reqCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)

			err = p.orderWriter.WriteMessages(reqCtx, kafka.Message{
				Key:   []byte(o.Order.ID.String()),
				Value: value,
			})

			if err != nil {
				errCh <- fmt.Errorf("failed to publish order: %w", err)
			}

			cancel()
		}
	}()
	go func() {
		defer p.wg.Done()
		for l := range p.locationBuffer {
			value, err := json.Marshal(*l)
			if err != nil {
				errCh <- fmt.Errorf("failed to marshal location: %w", err)
			}

			reqCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)

			err = p.locationWriter.WriteMessages(reqCtx, kafka.Message{
				Key:   []byte(l.OrderID.String()),
				Value: value,
			})

			if err != nil {
				errCh <- fmt.Errorf("failed to publish location: %w", err)
			}

			cancel()
		}
	}()
}

func (p *KafkaPublisher) Close() error {
	close(p.orderBuffer)
	close(p.locationBuffer)

	p.wg.Wait()

	err := p.orderWriter.Close()
	if err != nil {
		return fmt.Errorf("failed to close Kafka writer: %w", err)
	}
	err = p.locationWriter.Close()
	if err != nil {

		return fmt.Errorf("failed to close Kafka writer: %w", err)
	}
	return nil
}
