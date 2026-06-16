package dtos

import (
	"time"

	"github.com/LuisHFr15/delivery-tracker-distributed/internal/ingester/domain/models"

	"github.com/google/uuid"
)

type OrderEventDTO struct {
	EventID         uuid.UUID `json:"event_id,omitempty"`
	Order           OrderDTO  `json:"order"`
	TransactionType string    `json:"transaction_type"` // CREATE, CANCEL, UPDATE
	Timestamp       time.Time `json:"timestamp,omitempty"`
}

func (d *OrderEventDTO) ToDomain() models.Order {
	order := d.Order.ToDomain()
	order.EventId = d.EventID
	order.CreatedAt = d.Timestamp

	if order.EventId == uuid.Nil {
		order.EventId = uuid.New()
	}
	if order.CreatedAt.IsZero() {
		order.CreatedAt = time.Now()
	}

	return order
}
