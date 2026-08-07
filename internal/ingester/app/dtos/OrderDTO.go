package dtos

import (
	"github.com/LuisHFr15/delivery-tracker-distributed/internal/processor/domain/models/order"
	"github.com/google/uuid"
)

type OrderDTO struct {
	ID          uuid.UUID      `json:"id"`
	ClientID    uuid.UUID      `json:"client_id"`
	Products    []OrderItemDTO `json:"products"`
	DeliveryID  *uuid.UUID     `json:"delivery_id,omitempty"`
	Destination order.Location `json:"destination"`
	Status      string         `json:"status"`
}
