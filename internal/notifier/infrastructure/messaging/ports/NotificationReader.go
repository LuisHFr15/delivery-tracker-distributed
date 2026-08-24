package ports

import (
	"context"

	"github.com/LuisHFr15/delivery-tracker-distributed/internal/notifier/infrastructure/messaging/dtos"
)

type NotificationReader interface {
	Read(ctx context.Context) (dtos.ProcessedOrderNotificationDTO, error)
	Close() error
}
