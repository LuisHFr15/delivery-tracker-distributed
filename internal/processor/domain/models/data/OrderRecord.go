package data

import (
	"time"

	"github.com/LuisHFr15/delivery-tracker-distributed/internal/processor/domain/models/order"
	"github.com/google/uuid"
)

type OrderRecord struct {
	OrderId     uuid.UUID         `dynamodbav:"OrderId"`
	EventId     uuid.UUID         `dynamodbav:"EventId"`
	ClientId    uuid.UUID         `dynamodbav:"ClientId"`
	Products    []order.OrderItem `dynamodbav:"Products"`
	DeliveryId  *uuid.UUID        `dynamodbav:"DeliveryId,omitempty"`
	Destination order.Location    `dynamodbav:"Destination"`
	Status      string            `dynamodbav:"Status"`
	CreatedAt   time.Time         `dynamodbav:"CreatedAt"`
	CancelledAt *time.Time        `dynamodbav:"CancelledAt,omitempty"`
	DeliveredAt *time.Time        `dynamodbav:"DeliveredAt,omitempty"`
}

func FromOrder(o order.Order) OrderRecord {
	return OrderRecord{
		OrderId:     o.Id,
		EventId:     o.EventId,
		ClientId:    o.Client.Id,
		Products:    o.Products,
		DeliveryId:  o.DeliveryId,
		Destination: *o.Destination,
		Status:      o.Status,
		CreatedAt:   o.CreatedAt,
		CancelledAt: o.CancelledAt,
		DeliveredAt: o.DeliveredAt,
	}
}

func (r OrderRecord) ToDomain() order.Order {
	return order.Order{
		EventId:     r.EventId,
		Id:          r.OrderId,
		Products:    r.Products,
		Client:      order.Client{Id: r.ClientId},
		Destination: &r.Destination,
		DeliveryId:  r.DeliveryId,
		Status:      r.Status,
		CreatedAt:   r.CreatedAt,
		CancelledAt: r.CancelledAt,
		DeliveredAt: r.DeliveredAt,
	}
}
