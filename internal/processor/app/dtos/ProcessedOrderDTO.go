package dtos

import (
	"time"

	"github.com/LuisHFr15/delivery-tracker-distributed/internal/processor/domain/models/data"
	"github.com/LuisHFr15/delivery-tracker-distributed/internal/processor/domain/models/order"
	"github.com/google/uuid"
)

type ProcessedOrderDTO struct {
	OrderId          uuid.UUID      `json:"orderId"`
	ClientId         uuid.UUID      `json:"clientId"`
	TimeToDelivery   time.Duration  `json:"timeToDelivery"`
	Timestamp        time.Time      `json:"timestamp"`
	ActualLocation   order.Location `json:"actualLocation"`
	FinalDestination order.Location `json:"finalDestination"`
}

func FromDomain(p data.ProcessedOrder, ttd time.Duration) ProcessedOrderDTO {
	return ProcessedOrderDTO{
		OrderId:          p.OrderId,
		ClientId:         p.ClientId,
		TimeToDelivery:   ttd,
		Timestamp:        p.Timestamp,
		ActualLocation:   p.TimestampLocation,
		FinalDestination: p.FinalLocation,
	}
}
