package dtos

import (
	"time"

	"github.com/LuisHFr15/delivery-tracker-distributed/internal/processor/domain/models/order"

	"github.com/google/uuid"
)

type LocationEventDTO struct {
	EventID   uuid.UUID `json:"event_id,omitempty"`
	OrderID   uuid.UUID `json:"order_id"`
	Latitude  float64   `json:"latitude"`
	Longitude float64   `json:"longitude"`
	Timestamp time.Time `json:"timestamp,omitempty"`
}

func (d *LocationEventDTO) ToDomain() (uuid.UUID, uuid.UUID, order.Location, time.Time) {
	timestamp := d.Timestamp
	if timestamp.IsZero() {
		timestamp = time.Now()
	}

	return d.EventID, d.OrderID, order.Location{
		Lat: d.Latitude,
		Lng: d.Longitude,
	}, timestamp
}
