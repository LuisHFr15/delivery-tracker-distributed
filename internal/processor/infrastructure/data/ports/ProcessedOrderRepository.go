package ports

import (
	"context"

	"github.com/LuisHFr15/delivery-tracker-distributed/internal/processor/domain/models/data"
)

type ProcessedOrderRepository interface {
	Add(ctx context.Context, dto data.ProcessedOrder) error
	RunWorker()
	StopWorker()
}
