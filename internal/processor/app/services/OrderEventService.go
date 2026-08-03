package services

import (
	"github.com/LuisHFr15/delivery-tracker-distributed/internal/processor/app/dtos"
	repo "github.com/LuisHFr15/delivery-tracker-distributed/internal/processor/infrastructure/data/ports"
	queue "github.com/LuisHFr15/delivery-tracker-distributed/internal/processor/infrastructure/queue"
)

// TODO: create the usage to save both order event and order inside the dynamo
type OrderEventService struct {
	dto            dtos.OrderEventDTO
	writer         queue.KafkaWriter
	orderEventRepo repo.OrderEventRepository
	orderRepo      repo.OrderRepository
}
