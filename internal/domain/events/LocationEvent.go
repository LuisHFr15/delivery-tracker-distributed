package events

import (
	o "main/internal/domain/orders"
	t "time"

	"github.com/google/uuid"
)

type LocationEvent struct {
	Id        uuid.UUID  `json:"event_id"`
	Delivery  o.Delivery `json:"delivery"`
	TImestamp t.Time     `json:"timestamp"`
}
