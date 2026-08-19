package services

import (
	"testing"
	"time"

	"github.com/LuisHFr15/delivery-tracker-distributed/internal/processor/domain/models/order"
	"github.com/google/uuid"
)

func ptr[T any](v T) *T { return &v }

func TestDeliveryUpdater_ProcessDelivery(t *testing.T) {
	tests := []struct {
		name     string
		wantErr  bool
		order    *order.Order
		location order.Location
		eventId  uuid.UUID
	}{
		{
			name:    "success_delivery",
			wantErr: false,
			order: &order.Order{
				Id:          uuid.New(),
				EventId:     uuid.Nil,
				Status:      "DELIVERING",
				Destination: &order.Location{Lat: 42.0, Lng: -72.0},
			},
			location: order.Location{Lat: 42.0, Lng: -72.0},
			eventId:  uuid.New(),
		},
		{
			name:    "success_start_delviering",
			wantErr: false,
			order: &order.Order{
				Id:          uuid.New(),
				EventId:     uuid.Nil,
				Status:      "NEW",
				Destination: &order.Location{Lat: 42.0, Lng: -72.0},
			},
			location: order.Location{Lat: 42.0, Lng: -70.0},
			eventId:  uuid.New(),
		},
		{
			name:    "error_already_delivered",
			wantErr: true,
			order: &order.Order{
				Id:          uuid.New(),
				EventId:     uuid.New(),
				Status:      "DELIVERED",
				Destination: &order.Location{Lat: 42.0, Lng: -72.0},
				DeliveredAt: ptr(time.Now()),
			},
			location: order.Location{Lat: 42.0, Lng: -72.0},
			eventId:  uuid.New(),
		},
		{
			name:    "error_already_cancelled",
			wantErr: true,
			order: &order.Order{
				Id:          uuid.New(),
				EventId:     uuid.Nil,
				Status:      "CANCELLED",
				Destination: &order.Location{Lat: 42.0, Lng: -72.0},
				CancelledAt: ptr(time.Now()),
			},
			location: order.Location{Lat: 42.0, Lng: -72.0},
			eventId:  uuid.New(),
		},
		{
			name:    "error_missing_destination",
			wantErr: true,
			order: &order.Order{
				Id:          uuid.New(),
				EventId:     uuid.Nil,
				Status:      "NEW",
				Destination: nil,
			},
			location: order.Location{Lat: 42.0, Lng: -72.0},
			eventId:  uuid.New(),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			wantErr := tt.wantErr
			ord := tt.order
			location := tt.location
			eventId := tt.eventId
			delUpdater := NewDeliveryUpdater()
			ordPastLoc := ord.Destination
			ordPastStatus := ord.Status

			err := delUpdater.ProcessDelivery(ord, location, eventId)
			if (err != nil) != wantErr {
				t.Errorf("DeliveryUpdater.ProcessDelivery() error = %v, wantErr %v", err, wantErr)
			}
			if ord.Destination != ordPastLoc {
				t.Errorf("DeliveryUpdater.ProcessDelivery() changed ord.Destination = %v, want %v", ordPastLoc, ord)
			}
			if (ordPastStatus == "CANCELLED" || ordPastStatus == "DELIVERED") && ord.Status == "DELIVERING" {
				t.Errorf("DeliveryUpdater.ProcessDelivery() should not change status for cancelled or delivered orders")
			}
		})
	}
}
