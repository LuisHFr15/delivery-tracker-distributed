package dtos

import (
	m "main/internal/ingester/domain/models"
	t "time"

	"github.com/google/uuid"
)

type OrderEventDTO struct {
	EventID         uuid.UUID `json:"event_id,omitempty"`
	Order           OrderDTO  `json:"order"`
	TransactionType string    `json:"transaction_type"` // CREATE, CANCEL, UPDATE
	Timestamp       t.Time    `json:"timestamp,omitempty"`
}

func (d *OrderEventDTO) ToDomain() m.Order {
	order := d.Order.ToDomain()
	order.EventId = d.EventID
	order.CreatedAt = d.Timestamp

	if order.EventId == uuid.Nil {
		order.EventId = uuid.New()
	}
	if order.CreatedAt.IsZero() {
		order.CreatedAt = t.Now()
	}

	return order
}
