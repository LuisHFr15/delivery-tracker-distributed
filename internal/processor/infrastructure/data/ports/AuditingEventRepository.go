package ports

import (
	"context"

	"github.com/LuisHFr15/delivery-tracker-distributed/internal/processor/domain/models/data"
)

type OrderEventRepository interface {
	Add(ctx context.Context, cls data.DynamoEvent) error
}
