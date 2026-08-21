package ports

import (
	"context"

	"github.com/LuisHFr15/delivery-tracker-distributed/internal/processor/domain/models/data"
)

type ProcessedOrderRepository interface {
	Add(dto data.ProcessedOrder)
	RunWorker()
	StopWorker() error
	GetLatestByOrderId(ctx context.Context, orderId string) (*data.ProcessedOrder, error)
}
