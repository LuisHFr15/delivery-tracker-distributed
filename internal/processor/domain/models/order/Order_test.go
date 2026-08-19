package order

import (
	"testing"
	"time"

	"github.com/google/uuid"
)

func timePtr(t time.Time) *time.Time {
	return &t
}

func TestOrder_CalculateTimeToDelivery(t *testing.T) {
	tests := []struct {
		name    string
		wantErr bool
		order   *Order
	}{
		{
			name:    "success",
			wantErr: false,
			order: &Order{
				EventId:     uuid.Nil,
				Id:          uuid.Nil,
				CreatedAt:   time.Now().Add(-(time.Minute) * 5),
				Status:      "DELIVERED",
				DeliveredAt: timePtr(time.Now()),
			},
		},
		{
			name:    "missing creation date",
			wantErr: true,
			order: &Order{
				EventId:     uuid.Nil,
				Id:          uuid.Nil,
				Status:      "DELIVERED",
				DeliveredAt: timePtr(time.Now()),
			},
		},
		{
			name:    "not delivered",
			wantErr: true,
			order: &Order{
				EventId:   uuid.Nil,
				Id:        uuid.Nil,
				CreatedAt: time.Now().Add(-(time.Minute) * 5),
				Status:    "ONGOING",
			},
		},
		{
			name:    "cancelled order",
			wantErr: true,
			order: &Order{
				EventId:   uuid.Nil,
				Id:        uuid.Nil,
				CreatedAt: time.Now().Add(-(time.Minute) * 5),
				Status:    "CANCELLED",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			wantErr := tt.wantErr
			order := tt.order

			_, err := order.CalculateTimeToDelivery()

			if (err != nil) != wantErr {
				t.Errorf("CalculateTimeToDelivery() error = %v, wantErr %v", err, wantErr)
			}
		})
	}
}

func TestOrder_CalculateTimeToCancel(t *testing.T) {
	tests := []struct {
		name    string
		wantErr bool
		order   *Order
	}{
		{
			name:    "success",
			wantErr: false,
			order: &Order{
				EventId:     uuid.Nil,
				Id:          uuid.Nil,
				CreatedAt:   time.Now().Add(-(time.Minute) * 5),
				Status:      "CANCELLED",
				CancelledAt: timePtr(time.Now()),
			},
		},
		{
			name:    "missing creation date",
			wantErr: true,
			order: &Order{
				EventId:     uuid.Nil,
				Id:          uuid.Nil,
				Status:      "CANCELLED",
				CancelledAt: timePtr(time.Now()),
			},
		},
		{
			name:    "not cancelled",
			wantErr: true,
			order: &Order{
				EventId:     uuid.Nil,
				Id:          uuid.Nil,
				CreatedAt:   time.Now().Add(-(time.Minute) * 5),
				Status:      "ONGOING",
				CancelledAt: timePtr(time.Now()),
			},
		},
		{
			name:    "delivered order",
			wantErr: true,
			order: &Order{
				EventId:     uuid.Nil,
				Id:          uuid.Nil,
				CreatedAt:   time.Now().Add(-(time.Minute) * 5),
				Status:      "DELIVERED",
				DeliveredAt: timePtr(time.Now()),
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			wantErr := tt.wantErr
			order := tt.order
			_, err := order.CalculateTimeToCancel()
			if (err != nil) != wantErr {
				t.Errorf("CalculateTimeToCancel() error = %v, wantErr %v", err, wantErr)
			}
		})
	}
}

func TestOrder_Deliver(t *testing.T) {
	tests := []struct {
		name     string
		wantErr  bool
		location Location
		order    *Order
	}{
		{
			name:     "success",
			wantErr:  false,
			location: Location{Lat: 45, Lng: -45},
			order: &Order{
				EventId:     uuid.Nil,
				Id:          uuid.Nil,
				CreatedAt:   time.Now().Add(-(time.Minute) * 5),
				Status:      "ONGOING",
				Destination: &Location{Lat: 45, Lng: -45},
			},
		},
		{
			name:     "already delivered",
			wantErr:  true,
			location: Location{Lat: 45, Lng: -45},
			order: &Order{
				EventId:     uuid.Nil,
				Id:          uuid.Nil,
				CreatedAt:   time.Now().Add(-(time.Minute) * 5),
				Status:      "DELIVERED",
				DeliveredAt: timePtr(time.Now()),
				Destination: &Location{Lat: 45, Lng: -45},
			},
		},
		{
			name:     "already cancelled",
			wantErr:  true,
			location: Location{Lat: 45, Lng: -45},
			order: &Order{
				EventId:     uuid.Nil,
				Id:          uuid.Nil,
				CreatedAt:   time.Now().Add(-(time.Minute) * 5),
				Status:      "CANCELLED",
				CancelledAt: timePtr(time.Now()),
				Destination: &Location{Lat: 45, Lng: -45},
			},
		},
		{
			name:     "wrong location",
			wantErr:  true,
			location: Location{Lat: 45, Lng: 45},
			order: &Order{
				EventId:     uuid.Nil,
				Id:          uuid.Nil,
				CreatedAt:   time.Now().Add(-(time.Minute) * 5),
				Status:      "ONGOING",
				Destination: &Location{Lat: 45, Lng: -45},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			wantErr := tt.wantErr
			order := tt.order
			location := tt.location

			err := order.Deliver(uuid.New(), location)

			if (err != nil) != wantErr {
				t.Errorf("Deliver() error = %v, wantErr %v", err, wantErr)
			}
		})
	}
}
