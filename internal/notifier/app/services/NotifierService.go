package services

import (
	"log"
	"time"

	notifierdata "github.com/LuisHFr15/delivery-tracker-distributed/internal/notifier/domain/models/data"
	"github.com/LuisHFr15/delivery-tracker-distributed/internal/notifier/infrastructure/data/ports"
	"github.com/LuisHFr15/delivery-tracker-distributed/internal/processor/app/dtos"
)

// NotifierService turns a processed-order event into a customer notification.
// For this MVP "notifying" means logging a friendly message and persisting a
// NotifiedMessage record stamped with the moment it was acknowledged.
type NotifierService struct {
	dto  dtos.ProcessedOrderDTO
	repo ports.NotifiedMessageRepository
}

func NewNotifierService(dto dtos.ProcessedOrderDTO, repo ports.NotifiedMessageRepository) *NotifierService {
	return &NotifierService{dto: dto, repo: repo}
}

func (s *NotifierService) Notify() error {
	log.Printf("notifying client %s about order %s (status %s, ETA %s)",
		s.dto.ClientId, s.dto.OrderId, s.dto.OrderStatus, s.dto.TimeToDelivery)

	s.repo.Add(notifierdata.NotifiedMessage{
		OrderId:        s.dto.OrderId,
		ClientId:       s.dto.ClientId,
		OrderStatus:    s.dto.OrderStatus,
		ActualLocation: s.dto.ActualLocation,
		FinalLocation:  s.dto.FinalDestination,
		Timestamp:      time.Now(),
	})
	return nil
}
