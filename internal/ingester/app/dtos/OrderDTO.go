package dtos

import (
	"github.com/google/uuid"
)

type OrderDTO struct {
	ID         uuid.UUID      `json:"id"`
	ClientID   uuid.UUID      `json:"client_id"`
	Products   []OrderItemDTO `json:"products"`
	DeliveryID *uuid.UUID     `json:"delivery_id,omitempty"`
	Status     string         `json:"status"`
}
