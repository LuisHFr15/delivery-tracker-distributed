package ports

import (
	"context"

	"github.com/LuisHFr15/delivery-tracker-distributed/internal/processor/app/dtos"
)

type ProcessedOrderReader interface {
	Read(ctx context.Context) (dtos.ProcessedOrderDTO, error)
	Close() error
}
