package models

import (
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
)

var (
	ErrOrderNotCancelled        = errors.New("order is not cancelled")
	ErrOrderNotDelivered        = errors.New("order has not been delivered yet")
	ErrDeliveryLocationMismatch = errors.New("delivery location does not match destination")
	ErrInvalidStatus            = errors.New("invalid status: cannot be empty")
)

type OrderItem struct {
	Product  Product `json:"product"`
	Quantity int32   `json:"quantity"`
}

type Order struct {
	EventId     uuid.UUID   `json:"event_id,omitempty"`
	Id          uuid.UUID   `json:"id,omitempty"`
	Products    []OrderItem `json:"products"`
	Client      Client      `json:"client"`
	Destination Location    `json:"destination"`
	DeliveryId  *uuid.UUID  `json:"delivery_id,omitempty"`
	Status      string      `json:"status"`
	CreatedAt   time.Time   `json:"timestamp"`
	IsCancelled bool        `json:"is_cancelled"`
	CancelledAt *time.Time  `json:"cancelled_at,omitempty"`
	DeliveredAt *time.Time  `json:"delivered_at,omitempty"`
}

func (o *Order) CalculateTimeToDelivery() (time.Duration, error) {
	if o.DeliveredAt == nil {
		return 0, fmt.Errorf("%w: order %s", ErrOrderNotDelivered, o.Id)
	}

	return o.DeliveredAt.Sub(o.CreatedAt), nil
}

func (o *Order) CalculateTimeToCancel() (time.Duration, error) {
	if !o.IsCancelled || o.CancelledAt == nil {
		return 0, ErrOrderNotCancelled
	}

	return o.CancelledAt.Sub(o.CreatedAt), nil
}

func (o *Order) UpdateOrderStatus(eventId uuid.UUID, status string) error {
	trimString := strings.TrimSpace(status)
	if len(trimString) == 0 {
		return fmt.Errorf("%w: event %s tried to set empty status for order %s", ErrInvalidStatus, eventId, o.Id)
	}
	o.Status = trimString
	return nil
}

func (o *Order) CancelOrder() {
	now := time.Now()
	o.IsCancelled = true
	o.CancelledAt = &now
	o.Status = "CANCELLED"
}

func (o *Order) Deliver(eventId uuid.UUID, actualLocation Location) error {
	if !o.Destination.SameLocation(actualLocation) {
		return fmt.Errorf("%w: expected %v, got %v", ErrDeliveryLocationMismatch, o.Destination, actualLocation)
	}

	err := o.UpdateOrderStatus(eventId, "COMPLETED")
	if err != nil {
		return err
	}

	now := time.Now()
	o.DeliveredAt = &now
	return nil
}
