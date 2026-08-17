package data

import (
	"time"

	"github.com/LuisHFr15/delivery-tracker-distributed/internal/processor/domain/models/order"
	"github.com/google/uuid"
)

type ProcessedOrderFactory struct{}

func NewProcessedOrderFactory() *ProcessedOrderFactory {
	return &ProcessedOrderFactory{}
}

func (f *ProcessedOrderFactory) CreateProcessedOrder(o order.Order, eventId uuid.UUID, actualLoc order.Location, ts time.Time) ProcessedOrder {
	products := make([]order.Product, 0, len(o.Products))
	for _, item := range o.Products {
		products = append(products, item.Product)
	}

	return ProcessedOrder{
		OrderId:           o.Id,
		EventId:           eventId,
		ClientId:          o.Client.Id,
		OrderStatus:       o.Status,
		Products:          products,
		TimestampLocation: actualLoc,
		FinalLocation:     o.Destination,
		Timestamp:         ts,
	}
}
