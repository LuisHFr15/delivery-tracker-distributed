package services

import (
	"github.com/LuisHFr15/delivery-tracker-distributed/internal/processor/app/dtos"
	"github.com/LuisHFr15/delivery-tracker-distributed/internal/processor/domain/services"
	repo "github.com/LuisHFr15/delivery-tracker-distributed/internal/processor/infrastructure/data/ports"

	"github.com/google/uuid"
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

func (ee OrderEventService) ConvertEvent() error {
	converter := ee.converter
	originalEvent := ee.dto
	convertedEvent := converter.Convert(originalEvent)
	ee.auditRepo.Add(convertedEvent)

	if ee.dto.EventID == uuid.Nil {
		return ErrMissingEventID
	}

	ord, err := ee.dto.ToDomain()
	if err != nil {
		return err
	}
	ee.orderRepo.Add(ord)
	return nil
}
