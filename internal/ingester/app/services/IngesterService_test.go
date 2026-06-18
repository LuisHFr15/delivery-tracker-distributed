package services

import (
	"context"
	"errors"
	"testing"

	"github.com/LuisHFr15/delivery-tracker-distributed/internal/ingester/app/dtos"

	"github.com/google/uuid"
)

// FakePublisher é um fake "burro": apenas registra o que recebeu e devolve
// ErrToReturn. Não serializa nem valida — toda a lógica que queremos exercitar
// mora no IngesterService, não aqui. Se o fake tivesse lógica, o teste estaria
// verificando se fake e service concordam, não se o service está correto.
type FakePublisher struct {
	OrdersSent    []dtos.OrderEventDTO
	LocationsSent []dtos.LocationEventDTO
	ErrToReturn   error
}

func (p *FakePublisher) PublishOrder(_ context.Context, dto dtos.OrderEventDTO) error {
	if p.ErrToReturn != nil {
		return p.ErrToReturn
	}
	p.OrdersSent = append(p.OrdersSent, dto)
	return nil
}

func (p *FakePublisher) PublishLocation(_ context.Context, dto dtos.LocationEventDTO, _ uuid.UUID) error {
	if p.ErrToReturn != nil {
		return p.ErrToReturn
	}
	p.LocationsSent = append(p.LocationsSent, dto)
	return nil
}

func TestIngesterService_IngestOrder(t *testing.T) {
	tests := []struct {
		name       string
		order      dtos.OrderEventDTO
		publishErr error
		wantErr    bool
		wantSent   int
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
		{
			name:       "publisher error is propagated",
			order:      dtos.OrderEventDTO{Order: dtos.OrderDTO{ID: uuid.New()}},
			publishErr: errors.New("broker down"),
			wantErr:    true,
			wantSent:   0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			fake := &FakePublisher{ErrToReturn: tt.publishErr}
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
		name       string
		orderId    uuid.UUID
		publishErr error
		wantErr    bool
		wantSent   int
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
		{
			name:       "publisher error is propagated",
			orderId:    uuid.New(),
			publishErr: errors.New("broker down"),
			wantErr:    true,
			wantSent:   0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			fake := &FakePublisher{ErrToReturn: tt.publishErr}
			service := NewIngesterService(fake)

			err := service.IngestLocation(context.Background(), dtos.LocationEventDTO{}, tt.orderId)

			if (err != nil) != tt.wantErr {
				t.Errorf("IngestLocation() error = %v, wantErr %v", err, tt.wantErr)
			}
			if len(fake.LocationsSent) != tt.wantSent {
				t.Errorf("locations sent = %d, want %d", len(fake.LocationsSent), tt.wantSent)
			}
		})
	}
}
