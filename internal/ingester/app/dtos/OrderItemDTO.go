package dtos

import (
	"github.com/LuisHFr15/delivery-tracker-distributed/internal/ingester/domain/models"

	"github.com/google/uuid"
)

type OrderItemDTO struct {
	ProductID uuid.UUID `json:"product_id"`
	Quantity  int32     `json:"quantity"`
}

func (d *OrderItemDTO) ToDomain() models.OrderItem {
	return models.OrderItem{
		Product: models.Product{
			Id: d.ProductID,
		},
		Quantity: d.Quantity,
	}
}
