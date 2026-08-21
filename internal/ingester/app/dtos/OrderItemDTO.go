package dtos

import (
	"github.com/google/uuid"
)

type OrderItemDTO struct {
	ProductID uuid.UUID `json:"product_id"`
	Name      string    `json:"name"`
	Price     int32     `json:"price"`
	Quantity  int32     `json:"quantity"`
}
