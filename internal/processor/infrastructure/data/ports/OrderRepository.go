package ports

import (
	"context"

	"github.com/LuisHFr15/delivery-tracker-distributed/internal/processor/domain/models/order"
)

type OrderRepository interface {
	Add(o order.Order)
	RunWorker()
	StopWorker() error
	GetByOrderId(ctx context.Context, orderId string) (*order.Order, error)
}
