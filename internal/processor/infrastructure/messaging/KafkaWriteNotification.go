package messaging

import (
	"context"
	"encoding/json"
	"log"
	"os"

	"github.com/LuisHFr15/delivery-tracker-distributed/internal/processor/app/dtos"
	"github.com/segmentio/kafka-go"
)

type KafkaWriteNotification struct {
	writer *kafka.Writer
	buffer chan *dtos.ProcessedOrderDTO
	done   chan struct{}
}

func NewKafkaWriteNotification() *KafkaWriteNotification {
	broker := os.Getenv("KAFKA_BROKER")
	if broker == "" {
		broker = "localhost:9092"
	}
	return &KafkaWriteNotification{
		writer: &kafka.Writer{
			Addr:                   kafka.TCP(broker),
			Topic:                  "processed-orders",
			Balancer:               &kafka.LeastBytes{},
			AllowAutoTopicCreation: true,
		},
		buffer: make(chan *dtos.ProcessedOrderDTO, 100),
		done:   make(chan struct{}),
	}
}

func (kw *KafkaWriteNotification) Write(dto dtos.ProcessedOrderDTO) {
	kw.buffer <- &dto
}

func (kw *KafkaWriteNotification) RunWorker() {
	defer close(kw.done)
	for cls := range kw.buffer {
		msg, err := json.Marshal(cls)
		if err != nil {
			log.Printf("error marshalling json: %v", err)
			continue
		}
		err = kw.writer.WriteMessages(context.Background(), kafka.Message{
			Key:   []byte(cls.ClientId.String()),
			Value: []byte(msg),
		})
		if err != nil {
			log.Printf("error writing message: %v", err)
		} else {
			log.Printf("message sent to processed-orders")
		}
	}
}

func (kw *KafkaWriteNotification) StopWorker() error {
	close(kw.buffer)
	<-kw.done
	return kw.writer.Close()
}
