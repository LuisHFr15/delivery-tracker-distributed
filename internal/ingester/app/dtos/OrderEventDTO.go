package dtos

import (
	"time"

	"github.com/google/uuid"
)

type OrderEventDTO struct {
	EventID         uuid.UUID `json:"event_id,omitempty"`
	Order           OrderDTO  `json:"order"`
	TransactionType string    `json:"transaction_type"`
	Timestamp       time.Time `json:"timestamp,omitempty"`
}
