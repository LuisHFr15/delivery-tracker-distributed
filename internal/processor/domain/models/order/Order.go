package order

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
	ErrMissingCrucialData       = errors.New("missing data required")
	ErrInvalidOperation         = errors.New("this operation can't be performed to this order")
)

type OrderItem struct {
	Product  Product `json:"product" dynamodbav:"Product"`
	Quantity int32   `json:"quantity" dynamodbav:"Quantity"`
}

type Order struct {
	EventId     uuid.UUID
	Id          uuid.UUID
	Products    []OrderItem
	Client      Client
	Destination Location
	DeliveryId  *uuid.UUID
	Status      string
	CreatedAt   time.Time
	CancelledAt *time.Time
	DeliveredAt *time.Time
}

func (o *Order) CalculateTimeToDelivery() (time.Duration, error) {
	if !o.isDelivered() {
		return 0, fmt.Errorf("%w: order %s", ErrOrderNotDelivered, o.Id)
	}

	if o.CreatedAt.IsZero() {
		return 0, fmt.Errorf("%w: %s order %s", ErrMissingCrucialData, "createdAt", o.Id)
	}

	return o.DeliveredAt.Sub(o.CreatedAt), nil
}

func (o *Order) CalculateTimeToCancel() (time.Duration, error) {
	if !o.isCanceled() {
		return 0, ErrOrderNotCancelled
	}

	if o.CreatedAt.IsZero() {
		return 0, fmt.Errorf("%w: %s to order %s", ErrMissingCrucialData, "createdAt", o.Id)
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

func (o *Order) StartDelivering(eventId uuid.UUID) error {
	if o.isDelivered() {
		return fmt.Errorf("%w: event %s tried to start delivering an already delivered order %s", ErrInvalidOperation, eventId, o.Id)
	}
	if o.isCanceled() {
		return fmt.Errorf("%w: event %s tried to start delivering a canceled order %s", ErrInvalidOperation, eventId, o.Id)
	}
	if o.Status == "DELIVERING" {
		return nil
	}
	return o.UpdateOrderStatus(eventId, "DELIVERING")
}

func (o *Order) CancelOrder() {
	if o.CancelledAt != nil {
		return
	}
	now := time.Now()
	o.CancelledAt = &now
	o.Status = "CANCELLED"
}

func (o *Order) isCanceled() bool {
	return o.Status == "CANCELLED" && o.CancelledAt != nil && !o.CancelledAt.IsZero()
}

func (o *Order) isDelivered() bool {
	return o.Status == "DELIVERED" && o.DeliveredAt != nil && !o.DeliveredAt.IsZero()
}

func (o *Order) Deliver(eventId uuid.UUID, actualLocation Location) error {
	if o.isDelivered() {
		return fmt.Errorf("%w: event %s tried to deliver order already delivered %s", ErrInvalidOperation, eventId, o.Id)
	}
	if o.isCanceled() {
		return fmt.Errorf("%w: event %s tried to deliver order already canceled", ErrInvalidOperation, eventId)
	}

	if !o.Destination.SameLocation(actualLocation) {
		return fmt.Errorf("%w: expected %v, got %v", ErrDeliveryLocationMismatch, o.Destination, actualLocation)
	}

	err := o.UpdateOrderStatus(eventId, "DELIVERED")
	if err != nil {
		return err
	}

	now := time.Now()
	o.DeliveredAt = &now
	return nil
}
