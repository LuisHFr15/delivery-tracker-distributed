package dtos

import (
	m "main/internal/ingester/domain/models"
	t "time"

	"github.com/google/uuid"
)

type LocationEventDTO struct {
	EventID   uuid.UUID `json:"event_id,omitempty"`
	OrderID   uuid.UUID `json:"order_id"`
	Latitude  float64   `json:"latitude"`
	Longitude float64   `json:"longitude"`
	Timestamp t.Time    `json:"timestamp,omitempty"`
}

func (d *LocationEventDTO) ToDomain() (uuid.UUID, uuid.UUID, m.Location, t.Time) {
	eventId := d.EventID
	if eventId == uuid.Nil {
		eventId = uuid.New()
	}

	timestamp := d.Timestamp
	if timestamp.IsZero() {
		timestamp = t.Now()
	}

	return eventId, d.OrderID, m.Location{
		Lat: d.Latitude,
		Lng: d.Longitude,
	}, timestamp
}
