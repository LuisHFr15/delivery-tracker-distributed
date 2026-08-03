package ports

import (
	"context"

	"github.com/LuisHFr15/delivery-tracker-distributed/internal/processor/app/dtos"
)

// Auditing purpose only
type LocationEventRepository interface {
	Add(ctx context.Context, dto dtos.LocationEventDTO) error
}
