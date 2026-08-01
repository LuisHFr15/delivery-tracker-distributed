package services

import (
	"context"
	"fmt"

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
	if order.Order.ID == uuid.Nil {
		return fmt.Errorf("order id is required to ingest order")
	}
	return s.publisher.PublishOrder(ctx, order)
}

func (s *IngesterService) IngestLocation(ctx context.Context, dto dtos.LocationEventDTO, orderId uuid.UUID) error {
	if orderId == uuid.Nil {
		return fmt.Errorf("order id is required to ingest location")
	}
	return s.publisher.PublishLocation(ctx, dto, orderId)
}

func (s *IngesterService) Health(ctx context.Context) error {
	return nil
}
