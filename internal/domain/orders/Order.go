package orders

import (
	"errors"
	"fmt"
	"strings"
	t "time"

	"github.com/google/uuid"
)

var (
	ErrOrderNotCancelled        = errors.New("order is not cancelled")
	ErrOrderNotDelivered        = errors.New("order has not been delivered yet")
	ErrDeliveryLocationMismatch = errors.New("delivery location does not match destination")
	ErrInvalidStatus            = errors.New("invalid status: cannot be empty")
)

type Order struct {
	EventId     uuid.UUID         `json:"event_id"`
	Id          uuid.UUID         `json:"id"`
	Products    map[Product]int32 `json:"products"`
	Client      Client            `json:"client"`
	Destination Location          `json:"destination"`
	Status      string            `json:"status"`
	CreatedAt   t.Time            `json:"timestamp"`
	IsCancelled bool              `json:"is_cancelled"`
	CancelledAt *t.Time           `json:"cancelled_at,omitempty"`
	DeliveryAt  *t.Time           `json:"delivery_at,omitempty"`
}

func (o *Order) CalculateTimeToDelivery() (t.Duration, error) {
	if o.DeliveryAt == nil {
		return 0, fmt.Errorf("%w: order %s", ErrOrderNotDelivered, o.Id)
	}

	return o.DeliveryAt.Sub(o.CreatedAt), nil
}

func (o *Order) CalculateTimeToCancel() (t.Duration, error) {
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
	now := t.Now()
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

	now := t.Now()
	o.DeliveryAt = &now
	return nil
}
