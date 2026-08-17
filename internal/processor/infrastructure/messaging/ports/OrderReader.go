package ports

import (
	"context"

	"github.com/LuisHFr15/delivery-tracker-distributed/internal/processor/app/dtos"
)

type OrderReader interface {
	Read(ctx context.Context) (dtos.OrderEventDTO, error)
	Close() error
}
