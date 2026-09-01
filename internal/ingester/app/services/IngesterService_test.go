package services

import (
	"context"
	"testing"

	"github.com/LuisHFr15/delivery-tracker-distributed/internal/ingester/app/dtos"

	"github.com/google/uuid"
)

type FakePublisher struct {
	OrdersSent    []dtos.OrderEventDTO
	LocationsSent []dtos.LocationEventDTO
}

func (p *FakePublisher) PublishOrder(dto dtos.OrderEventDTO) {
	p.OrdersSent = append(p.OrdersSent, dto)
}

func (p *FakePublisher) PublishLocation(dto dtos.LocationEventDTO) {
	p.LocationsSent = append(p.LocationsSent, dto)
}

func TestIngesterService_IngestOrder(t *testing.T) {
	tests := []struct {
		name     string
		order    dtos.OrderEventDTO
		wantErr  bool
		wantSent int
	}{
		{
			name:     "valid order is published",
			order:    dtos.OrderEventDTO{Order: dtos.OrderDTO{ID: uuid.New()}},
			wantErr:  false,
			wantSent: 1,
		},
		{
			name:     "order without id is rejected before publishing",
			order:    dtos.OrderEventDTO{Order: dtos.OrderDTO{ID: uuid.Nil}},
			wantErr:  true,
			wantSent: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			fake := &FakePublisher{}
			service := NewIngesterService(fake)

			err := service.IngestOrder(context.Background(), tt.order)

			if (err != nil) != tt.wantErr {
				t.Errorf("IngestOrder() error = %v, wantErr %v", err, tt.wantErr)
			}
			if len(fake.OrdersSent) != tt.wantSent {
				t.Errorf("orders sent = %d, want %d", len(fake.OrdersSent), tt.wantSent)
			}
		})
	}
}

func TestIngesterService_IngestLocation(t *testing.T) {
	tests := []struct {
		name     string
		orderId  uuid.UUID
		wantErr  bool
		wantSent int
	}{
		{
			name:     "valid location is published",
			orderId:  uuid.New(),
			wantErr:  false,
			wantSent: 1,
		},
		{
			name:     "location without order id is rejected before publishing",
			orderId:  uuid.Nil,
			wantErr:  true,
			wantSent: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			fake := &FakePublisher{}
			service := NewIngesterService(fake)

			err := service.IngestLocation(context.Background(), dtos.LocationEventDTO{OrderID: tt.orderId})

			if (err != nil) != tt.wantErr {
				t.Errorf("IngestLocation() error = %v, wantErr %v", err, tt.wantErr)
			}
			if len(fake.LocationsSent) != tt.wantSent {
				t.Errorf("locations sent = %d, want %d", len(fake.LocationsSent), tt.wantSent)
			}
		})
	}
}
