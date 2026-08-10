package dtos

import (
	"time"

	"github.com/LuisHFr15/delivery-tracker-distributed/internal/processor/domain/models/order"
	"github.com/google/uuid"
)

type ProcessedOrderDTO struct {
	OrderId          uuid.UUID      `json:"orderId"`
	ClientId         uuid.UUID      `json:"clientId"`
	TimeToDelivery   time.Time      `json:"timeToDelivery"`
	Timestamp        time.Time      `json:"timestamp"`
	ActualLocation   order.Location `json:"actualLocation"`
	FinalDestination order.Location `json:"finalDestination"`
}
