package ports

import (
	"context"

	"github.com/LuisHFr15/delivery-tracker-distributed/internal/processor/domain/models"
)

type LocationRepository interface {
	Add(ctx context.Context, loc models.Location) error
}
