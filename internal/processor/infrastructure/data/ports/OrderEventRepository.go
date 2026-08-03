package ports

import (
	"context"

	"github.com/LuisHFr15/delivery-tracker-distributed/internal/processor/app/dtos"
)

// Auditing purpose only
type OrderEventRepository interface {
	Add(ctx context.Context, dto dtos.OrderEventDTO) error
}
