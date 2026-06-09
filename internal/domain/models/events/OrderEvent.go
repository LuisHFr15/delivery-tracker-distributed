package events

import (
	o "main/internal/domain/models/orders"
	t "time"

	"github.com/google/uuid"
)

type OrderEvent struct {
	Id              uuid.UUID `json:"event_id"`
	RelatedOrder    o.Order   `json:"order"`
	TransactionType string    `json:"transaction_type"`
	Timestamp       t.Time    `json:"timestamp"`
}
