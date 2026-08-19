package services

import (
	"context"
	"log"

	"github.com/LuisHFr15/delivery-tracker-distributed/internal/processor/app/dtos"
	"github.com/LuisHFr15/delivery-tracker-distributed/internal/processor/domain/models/data"
	"github.com/LuisHFr15/delivery-tracker-distributed/internal/processor/domain/models/order"
	"github.com/LuisHFr15/delivery-tracker-distributed/internal/processor/domain/services"
	repo "github.com/LuisHFr15/delivery-tracker-distributed/internal/processor/infrastructure/data/ports"
	"github.com/LuisHFr15/delivery-tracker-distributed/internal/processor/infrastructure/messaging/ports"
)

type LocationEventService struct {
	dto          dtos.LocationEventDTO
	lastLocation order.Location
	calculator   services.DeliveryCalculator
	delUpdater   services.DeliveryUpdater
	auditRepo    repo.AuditingEventRepository
	procRepo     repo.ProcessedOrderRepository
	orderRepo    repo.OrderRepository
	writer       ports.NotificationWriter
}

func NewLocationEventService(dto dtos.LocationEventDTO, auditRepo repo.AuditingEventRepository,
	procRepo repo.ProcessedOrderRepository, orderRepo repo.OrderRepository,
	writer ports.NotificationWriter) *LocationEventService {
	return &LocationEventService{
		dto:        dto,
		calculator: *services.NewDeliveryCalculator(),
		delUpdater: *services.NewDeliveryUpdater(),
		auditRepo:  auditRepo,
		procRepo:   procRepo,
		orderRepo:  orderRepo,
		writer:     writer,
	}
}

func (l *LocationEventService) ProcessEvent() error {
	locConverter := services.NewLocationEventConverter()
	factory := data.NewProcessedOrderFactory()

	eventId, _, _, ts := l.dto.ToDomain()
	dto := l.dto
	dto.EventID = eventId
	dto.Timestamp = ts
	actualLoc := order.Location{
		Lat: dto.Latitude,
		Lng: dto.Longitude,
	}

	oId := dto.OrderID.String()
	ord, err := l.orderRepo.GetByOrderId(context.Background(), oId)
	if err != nil {
		log.Printf("could not get order %s: %s", oId, err)
		return err
	}
	if ord == nil {
		log.Printf("order %s not found, skipping location event", oId)
		return nil
	}

	if err := l.delUpdater.ProcessDelivery(ord, actualLoc, eventId); err != nil {
		log.Printf("could not process delivery for order %s: %s", oId, err)
		return err
	}
	timeToDelivery := l.calculator.CalculateTime(actualLoc, *ord.Destination)

	event := locConverter.Convert(dto, ord.Status)
	l.auditRepo.Add(event)

	l.orderRepo.Add(*ord)

	proc := factory.CreateProcessedOrder(*ord, eventId, actualLoc, ts)
	l.procRepo.Add(proc)

	l.writer.Write(dtos.FromDomain(proc, timeToDelivery))
	return nil
}
