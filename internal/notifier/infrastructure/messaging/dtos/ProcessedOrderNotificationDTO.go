package dtos

import (
	"time"

	"github.com/google/uuid"
)

type Location struct {
	Lat float64 `json:"lat"`
	Lng float64 `json:"lng"`
}

type ProcessedOrderNotificationDTO struct {
	OrderId          uuid.UUID     `json:"orderId"`
	ClientId         uuid.UUID     `json:"clientId"`
	TimeToDelivery   time.Duration `json:"timeToDelivery"`
	Timestamp        time.Time     `json:"timestamp"`
	ActualLocation   Location      `json:"actualLocation"`
	FinalDestination Location      `json:"finalDestination"`
}
