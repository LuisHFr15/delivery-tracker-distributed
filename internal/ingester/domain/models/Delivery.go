package models

import (
	"time"

	"github.com/google/uuid"
)

type Delivery struct {
	EventId          uuid.UUID            `json:"event_id"`
	Id               uuid.UUID            `json:"id"`
	OrderIds         map[uuid.UUID]string `json:"order_ids"`
	StartingPosition Location             `json:"starting_position"`
	CurrentPosition  Location             `json:"current_location"`
	StartedAt        time.Time            `json:"started_at"`
	FinishedAt       time.Time            `json:"finished_at"`
}

func (d *Delivery) Arrive(id uuid.UUID) error {
	d.OrderIds[id] = "DELIVERED"
	if d.hasPendindOrders() {
		return nil
	}
	d.FinishedAt = time.Now()
	return nil
}

func (d *Delivery) hasPendindOrders() bool {
	for _, status := range d.OrderIds {
		if status != "DELIVERED" {
			return true
		}
	}
	return false
}
