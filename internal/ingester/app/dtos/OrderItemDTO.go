package dtos

import (
	m "main/internal/ingester/domain/models"

	"github.com/google/uuid"
)

type OrderItemDTO struct {
	ProductID uuid.UUID `json:"product_id"`
	Quantity  int32     `json:"quantity"`
}

func (d *OrderItemDTO) ToDomain() m.OrderItem {
	return m.OrderItem{
		Product: m.Product{
			Id: d.ProductID,
		},
		Quantity: d.Quantity,
	}
}
