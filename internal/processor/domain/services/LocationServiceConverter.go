package services

import (
	"github.com/LuisHFr15/delivery-tracker-distributed/internal/processor/app/dtos"
	"github.com/LuisHFr15/delivery-tracker-distributed/internal/processor/domain/models/data"
	"github.com/LuisHFr15/delivery-tracker-distributed/internal/processor/domain/models/order"
)

type LocationEventConverter struct {
}

func NewLocationEventConverter() LocationEventConverter {
	return LocationEventConverter{}
}

func (ee LocationEventConverter) Convert(dto dtos.LocationEventDTO, status string) data.DynamoEvent {
	return data.DynamoEvent{
		EventType: "location_event",
		Id:        dto.EventID,
		OrderId:   dto.OrderID,
		Status:    status,
		Location:  order.Location{Lat: dto.Latitude, Lng: dto.Longitude},
		Timestamp: dto.Timestamp,
	}
}
