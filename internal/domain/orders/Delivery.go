package orders

import (
	"time"

	"github.com/google/uuid"
)

type Delivery struct {
	EventId          uuid.UUID  `json:"event_id"`
	Id               uuid.UUID  `json:"id"`
	Order            Order      `json:"order"`
	StartingPosition Location   `json:"starting_position"`
	CurrentPosition  Location   `json:"current_location"`
	StartedAt        time.Time  `json:"started_at"`
	Arrived          bool       `json:"arrived"`
	ArrivedAt        *time.Time `json:"arrived_at,omitempty"`
}

func (d *Delivery) Arrive() error {
	err := d.Order.Deliver(d.EventId, d.CurrentPosition)
	if err != nil {
		return err
	}
	now := time.Now()
	d.Arrived = true
	d.ArrivedAt = &now
	return nil
}
