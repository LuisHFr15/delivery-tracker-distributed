package ports

import "github.com/segmentio/kafka-go"

type KafkaWriter interface {
	// TODO: create the notification DTO
	Write(partition int32, offset int64, message *kafka.Message) error
}
