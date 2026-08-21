package dynamo

import (
	"testing"
	"time"

	"github.com/LuisHFr15/delivery-tracker-distributed/internal/processor/domain/models/data"
	"github.com/LuisHFr15/delivery-tracker-distributed/internal/processor/domain/models/order"
	"github.com/google/uuid"
)

func TestDynamoProcessedOrderRepository_GetLatestByOrderId(t *testing.T) {
	ordId := uuid.New()
	destination := order.Location{Lat: 40, Lng: 20}
	orders := []data.ProcessedOrder{
		{
			OrderId:           ordId,
			EventId:           uuid.New(),
			OrderStatus:       "DELIVERING",
			TimestampLocation: order.Location{Lat: 45, Lng: 45},
			FinalLocation:     destination,
			Timestamp:         time.Date(2026, 8, 21, 8, 45, 0, 0, time.UTC),
		},
		{
			OrderId:           ordId,
			EventId:           uuid.New(),
			OrderStatus:       "DELIVERING",
			TimestampLocation: order.Location{Lat: 44.912, Lng: 45},
			Timestamp:         time.Date(2026, 8, 21, 8, 45, 2, 0, time.UTC),
		},
		{
			OrderId:           ordId,
			EventId:           uuid.New(),
			OrderStatus:       "DELIVERING",
			TimestampLocation: order.Location{Lat: 44.9, Lng: 45},
			Timestamp:         time.Date(2026, 8, 21, 8, 45, 4, 0, time.UTC),
		},
	}
	var latest *data.ProcessedOrder

	for i := range orders {
		if latest == nil || orders[i].Timestamp.After(latest.Timestamp) {
			latest = &orders[i]
		}
	}

	expected := time.Date(2026, 8, 21, 8, 45, 4, 0, time.UTC)
	if latest == nil || !latest.Timestamp.Equal(expected) {
		t.Fatalf("Latest time was %v; expected %v", latest, expected)
	}
}
