package ports

import (
	"context"

	"github.com/LuisHFr15/delivery-tracker-distributed/internal/processor/domain/models"
)

type ClientRepository interface {
	Add(ctx context.Context, client models.Client) error
}
