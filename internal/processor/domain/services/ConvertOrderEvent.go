package services

import (
	"time"

	"github.com/LuisHFr15/delivery-tracker-distributed/internal/processor/app/dtos"
	"github.com/LuisHFr15/delivery-tracker-distributed/internal/processor/domain/models/data"
)

type ConvertOrderEvent struct {
}

func NewConvertOrderEvent() ConvertOrderEvent {
	return ConvertOrderEvent{}
}

func (ee ConvertOrderEvent) Convert(dto dtos.OrderEventDTO) data.DynamoEvent {
	return data.DynamoEvent{
		EventType: "order_event",
		Id:        dto.EventID,
		OrderId:   dto.Order.ID,
		Status:    dto.TransactionType,
		Location:  dto.Order.Destination,
		Timestamp: time.Now(),
	}
}
