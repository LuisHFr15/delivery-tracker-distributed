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
	Destination order.Location `json:"location"`
	Status      string         `json:"status"`
}

func (d *OrderDTO) ToDomain() order.Order {
	order := order.Order{
		Id:         d.ID,
		Client:     order.Client{Id: d.ClientID},
		DeliveryId: d.DeliveryID,
		Status:     d.Status,
	}

	for _, p := range d.Products {
		order.Products = append(order.Products, p.ToDomain())
	}

	return order
}
