package ports

import (
	"context"

	"github.com/LuisHFr15/delivery-tracker-distributed/internal/processor/app/dtos"
)

type LocationReader interface {
	Read(ctx context.Context) (dtos.LocationEventDTO, error)
	Close() error
}
