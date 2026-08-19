package services

import (
	"time"

	"github.com/LuisHFr15/delivery-tracker-distributed/internal/processor/app/dtos"
	"github.com/LuisHFr15/delivery-tracker-distributed/internal/processor/domain/models/data"
	"github.com/LuisHFr15/delivery-tracker-distributed/internal/processor/domain/models/order"
)

type OrderEventConverter struct {
}

func NewOrderEventConverter() OrderEventConverter {
	return OrderEventConverter{}
}

func (ee OrderEventConverter) Convert(dto dtos.OrderEventDTO) data.DynamoEvent {
	var location order.Location
	if dto.Order.Destination != nil {
		location = *dto.Order.Destination
	}

	return data.DynamoEvent{
		EventType: "order_event",
		Id:        dto.EventID,
		OrderId:   dto.Order.ID,
		Status:    dto.TransactionType,
		Location:  location,
		Timestamp: time.Now(),
	}
}
