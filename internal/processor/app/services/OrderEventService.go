package services

import (
	"context"

	"github.com/LuisHFr15/delivery-tracker-distributed/internal/processor/app/dtos"
	"github.com/LuisHFr15/delivery-tracker-distributed/internal/processor/domain/services"
	repo "github.com/LuisHFr15/delivery-tracker-distributed/internal/processor/infrastructure/data/ports"
)

type OrderEventService struct {
	dto       dtos.OrderEventDTO
	auditRepo repo.AuditingEventRepository
	converter services.ConvertOrderEvent
}

func NewOrderEventService(dto dtos.OrderEventDTO, auditRepo repo.AuditingEventRepository,
	converter services.ConvertOrderEvent) OrderEventService {
	return OrderEventService{
		dto:       dto,
		auditRepo: auditRepo,
		converter: converter,
	}
}

func (ee OrderEventService) ConvertEvent(ctx context.Context) {
	converter := ee.converter
	originalEvent := ee.dto
	convertedEvent := converter.Convert(originalEvent)
	ee.auditRepo.Add(convertedEvent)
}
