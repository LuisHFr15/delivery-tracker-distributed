package services

import (
	"github.com/LuisHFr15/delivery-tracker-distributed/internal/processor/app/dtos"
	"github.com/LuisHFr15/delivery-tracker-distributed/internal/processor/domain/services"
	repo "github.com/LuisHFr15/delivery-tracker-distributed/internal/processor/infrastructure/data/ports"
)

type OrderEventService struct {
	dto       dtos.OrderEventDTO
	auditRepo repo.AuditingEventRepository
	orderRepo repo.OrderRepository
	converter services.OrderEventConverter
}

func NewOrderEventService(dto dtos.OrderEventDTO, auditRepo repo.AuditingEventRepository, orderRepo repo.OrderRepository) OrderEventService {
	return OrderEventService{
		dto:       dto,
		auditRepo: auditRepo,
		orderRepo: orderRepo,
		converter: services.NewOrderEventConverter(),
	}
}

func (ee OrderEventService) ConvertEvent() {
	converter := ee.converter
	originalEvent := ee.dto
	convertedEvent := converter.Convert(originalEvent)
	ee.auditRepo.Add(convertedEvent)

	ee.orderRepo.Add(ee.dto.ToDomain())
}
