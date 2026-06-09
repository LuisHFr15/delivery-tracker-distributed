package events

import (
	o "main/internal/domain/models/orders"
	t "time"

	"github.com/google/uuid"
)

type LocationEvent struct {
	Id              uuid.UUID  `json:"event_id"`
	OrderId         uuid.UUID  `json:"order"`
	CurrentPosition o.Location `json:"current_position"`
	Timestamp       t.Time     `json:"timestamp"`
}
