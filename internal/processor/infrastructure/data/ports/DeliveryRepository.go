package ports

import (
	"context"

	"github.com/LuisHFr15/delivery-tracker-distributed/internal/processor/domain/models"
)

type DeliveryRepository interface {
	Add(ctx context.Context, delivery models.Delivery) error
	Update(ctx context.Context, delivery models.Delivery) error
}
