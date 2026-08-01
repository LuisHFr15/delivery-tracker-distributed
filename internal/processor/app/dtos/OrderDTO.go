package dtos

import (
	"github.com/LuisHFr15/delivery-tracker-distributed/internal/ingester/domain/models"

	"github.com/google/uuid"
)

type OrderDTO struct {
	ID         uuid.UUID      `json:"id"`
	ClientID   uuid.UUID      `json:"client_id"`
	Products   []OrderItemDTO `json:"products"`
	DeliveryID *uuid.UUID     `json:"delivery_id,omitempty"`
	Status     string         `json:"status"`
}

func (d *OrderDTO) ToDomain() models.Order {
	order := models.Order{
		Id:         d.ID,
		Client:     models.Client{Id: d.ClientID},
		DeliveryId: d.DeliveryID,
		Status:     d.Status,
	}

	for _, p := range d.Products {
		order.Products = append(order.Products, p.ToDomain())
	}

	return order
}
