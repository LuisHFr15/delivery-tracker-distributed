package services

import (
	"context"

	"github.com/LuisHFr15/delivery-tracker-distributed/internal/ingester/app/dtos"

	"github.com/google/uuid"
)

type EventPublisher interface {
	PublishOrder(ctx context.Context, dto dtos.OrderEventDTO) error
	PublishLocation(ctx context.Context, dto dtos.LocationEventDTO, orderId uuid.UUID) error
}

type IngesterService struct {
	publisher EventPublisher
}

func NewIngesterService(publisher EventPublisher) *IngesterService {
	return &IngesterService{publisher: publisher}
}

func (s *IngesterService) IngestOrder(ctx context.Context, order dtos.OrderEventDTO) error {
	return s.publisher.PublishOrder(ctx, order)
}

func (s *IngesterService) IngestLocation(ctx context.Context, dto dtos.LocationEventDTO, orderId uuid.UUID) error {
	return s.publisher.PublishLocation(ctx, dto, orderId)
}
