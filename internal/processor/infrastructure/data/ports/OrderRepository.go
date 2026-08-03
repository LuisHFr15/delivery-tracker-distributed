package ports

import (
	"context"

	"github.com/LuisHFr15/delivery-tracker-distributed/internal/processor/domain/models"
)

type OrderRepository interface {
	Add(ctx context.Context, order models.Order) error
	Update(ctx context.Context, order models.Order) error
}
