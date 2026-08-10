package ports

import "github.com/LuisHFr15/delivery-tracker-distributed/internal/processor/app/dtos"

type KafkaWriter interface {
	Write(dto dtos.ProcessedOrderDTO)
}
