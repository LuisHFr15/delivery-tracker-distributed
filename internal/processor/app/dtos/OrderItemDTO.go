package dtos

import (
	"github.com/LuisHFr15/delivery-tracker-distributed/internal/processor/domain/models/order"

	"github.com/google/uuid"
)

type OrderItemDTO struct {
	ProductID uuid.UUID `json:"product_id"`
	Name      string    `json:"name"`
	Price     int32     `json:"price"`
	Quantity  int32     `json:"quantity"`
}

func (d *OrderItemDTO) ToDomain() order.OrderItem {
	return order.OrderItem{
		Product: order.Product{
			Id:           d.ProductID,
			Name:         d.Name,
			ProductPrice: d.Price,
		},
		Quantity: d.Quantity,
	}
}
