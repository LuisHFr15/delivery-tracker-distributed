package dtos

import (
	"time"

	"github.com/LuisHFr15/delivery-tracker-distributed/internal/processor/domain/models/order"

	"github.com/google/uuid"
)

type OrderEventDTO struct {
	EventID         uuid.UUID `json:"event_id,omitempty"`
	Order           OrderDTO  `json:"order"`
	TransactionType string    `json:"transaction_type"`
	Timestamp       time.Time `json:"timestamp,omitempty"`
}

func (d *OrderEventDTO) ToDomain() (order.Order, error) {
	ord, err := d.Order.ToDomain()
	if err != nil {
		return order.Order{}, err
	}
	ord.EventId = d.EventID
	ord.CreatedAt = d.Timestamp

	if ord.EventId == uuid.Nil {
		ord.EventId = uuid.New()
	}
	if ord.CreatedAt.IsZero() {
		ord.CreatedAt = time.Now()
	}

	return ord, nil
}
