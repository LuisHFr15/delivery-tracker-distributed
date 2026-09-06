package services

import (
	"context"
	"fmt"

	"github.com/LuisHFr15/delivery-tracker-distributed/internal/ingester/app/dtos"
	"github.com/google/uuid"
)

type EventPublisher interface {
	PublishOrder(dto dtos.OrderEventDTO)
	PublishLocation(dto dtos.LocationEventDTO)
}

type IngesterService struct {
	publisher EventPublisher
}

func NewIngesterService(publisher EventPublisher) *IngesterService {
	return &IngesterService{publisher: publisher}
}

func (s *IngesterService) IngestOrder(_ context.Context, order dtos.OrderEventDTO) error {
	if order.Order.ID == uuid.Nil {
		return fmt.Errorf("order id is required to ingest order")
	}
	s.publisher.PublishOrder(order)
	return nil
}

func (s *IngesterService) IngestLocation(_ context.Context, dto dtos.LocationEventDTO) error {
	if dto.OrderID == uuid.Nil {
		return fmt.Errorf("order id is required to ingest location")
	}
	s.publisher.PublishLocation(dto)
	return nil
}

func (s *IngesterService) Health(_ context.Context) error {
	return nil
}
