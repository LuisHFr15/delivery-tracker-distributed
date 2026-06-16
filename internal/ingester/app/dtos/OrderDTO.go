package dtos

import (
	m "main/internal/ingester/domain/models"

	"github.com/google/uuid"
)

type OrderDTO struct {
	ID         uuid.UUID      `json:"id"`
	ClientID   uuid.UUID      `json:"client_id"`
	Products   []OrderItemDTO `json:"products"`
	DeliveryID *uuid.UUID     `json:"delivery_id,omitempty"`
	Status     string         `json:"status"`
}

func (d *OrderDTO) ToDomain() m.Order {
	order := m.Order{
		Id:         d.ID,
		Client:     m.Client{Id: d.ClientID},
		DeliveryId: d.DeliveryID,
		Status:     d.Status,
	}

	for _, p := range d.Products {
		order.Products = append(order.Products, p.ToDomain())
	}

	return order
}
