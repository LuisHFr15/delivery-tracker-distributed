package dtos

import (
	"time"

	"github.com/google/uuid"
)

type LocationEventDTO struct {
	EventID   uuid.UUID `json:"event_id,omitempty"`
	OrderID   uuid.UUID `json:"order_id"`
	Latitude  float64   `json:"latitude"`
	Longitude float64   `json:"longitude"`
	Timestamp time.Time `json:"timestamp,omitempty"`
}
