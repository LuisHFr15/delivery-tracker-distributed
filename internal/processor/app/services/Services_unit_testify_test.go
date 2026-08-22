package services

import (
	"errors"
	"testing"

	"github.com/LuisHFr15/delivery-tracker-distributed/internal/processor/domain/models/data"
	"github.com/LuisHFr15/delivery-tracker-distributed/internal/processor/domain/models/order"
	domainservices "github.com/LuisHFr15/delivery-tracker-distributed/internal/processor/domain/services"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// These tests mirror Services_unit_stdlib_test.go one-to-one so the two
// assertion styles can be compared side by side. Function names carry a
// _Testify suffix to avoid collisions within the same package.

// -----------------------------------------------------------------------------
// OrderEventService
// -----------------------------------------------------------------------------

func TestOrderEventService_ConvertEvent_Success_Testify(t *testing.T) {
	dto := validOrderEventDTO()
	audit := &fakeAuditRepo{}
	orders := &fakeOrderRepo{}

	svc := NewOrderEventService(dto, audit, orders)
	require.NoError(t, svc.ConvertEvent())

	require.Len(t, audit.added, 1)
	require.Len(t, orders.added, 1)

	ev := audit.added[0]
	assert.Equal(t, "order_event", ev.EventType)
	assert.Equal(t, dto.TransactionType, ev.Status)
	assert.Equal(t, dto.Order.ID, ev.OrderId)
	assert.Equal(t, *dto.Order.Destination, ev.Location)

	got := orders.added[0]
	assert.Equal(t, dto.Order.ID, got.Id)
	assert.Equal(t, dto.Order.ClientID, got.Client.Id)
	assert.Equal(t, dto.Order.Status, got.Status)
	require.Len(t, got.Products, len(dto.Order.Products))
	assert.Equal(t, dto.Order.Products[0].ProductID, got.Products[0].Product.Id)
	assert.Equal(t, dto.Order.Products[0].Quantity, got.Products[0].Quantity)
}

func TestOrderEventService_ConvertEvent_MissingDestination_Testify(t *testing.T) {
	dto := validOrderEventDTO()
	dto.Order.Destination = nil

	audit := &fakeAuditRepo{}
	orders := &fakeOrderRepo{}

	svc := NewOrderEventService(dto, audit, orders)
	err := svc.ConvertEvent()

	require.ErrorIs(t, err, order.ErrMissingDestination)
	assert.Len(t, audit.added, 1, "audit is written before validation")
	assert.Empty(t, orders.added, "order failed validation")
}

func TestOrderEventService_ConvertEvent_MissingEventID_Testify(t *testing.T) {
	dto := validOrderEventDTO()
	dto.EventID = uuid.Nil

	audit := &fakeAuditRepo{}
	orders := &fakeOrderRepo{}

	svc := NewOrderEventService(dto, audit, orders)
	err := svc.ConvertEvent()

	require.ErrorIs(t, err, ErrMissingEventID)
	assert.Len(t, audit.added, 1, "audit is written before validation")
	assert.Empty(t, orders.added, "event rejected")
}

func TestOrderEventService_ConvertEvent_EmptyProducts_Testify(t *testing.T) {
	dto := validOrderEventDTO()
	dto.Order.Products = nil
	orders := &fakeOrderRepo{}

	svc := NewOrderEventService(dto, &fakeAuditRepo{}, orders)
	require.NoError(t, svc.ConvertEvent())

	require.Len(t, orders.added, 1)
	assert.Empty(t, orders.added[0].Products)
}

func TestOrderEventService_ConvertEvent_InputIntegrity_Testify(t *testing.T) {
	dto := validOrderEventDTO()
	snapshot := dto
	snapshotDest := *dto.Order.Destination

	svc := NewOrderEventService(dto, &fakeAuditRepo{}, &fakeOrderRepo{})
	require.NoError(t, svc.ConvertEvent())

	assert.Equal(t, snapshot.EventID, dto.EventID)
	assert.Equal(t, snapshot.Order.Status, dto.Order.Status)
	assert.Equal(t, snapshotDest, *dto.Order.Destination)
}

func TestOrderEventService_ConvertEvent_ValueCorrespondence_Testify(t *testing.T) {
	dto := validOrderEventDTO()
	audit := &fakeAuditRepo{}
	orders := &fakeOrderRepo{}

	svc := NewOrderEventService(dto, audit, orders)
	require.NoError(t, svc.ConvertEvent())

	// No divergence: audit id == persisted order EventId == dto.EventID.
	assert.Equal(t, dto.EventID, audit.added[0].Id)
	assert.Equal(t, dto.EventID, orders.added[0].EventId)

	// Type 2: delegates to OrderEventConverter (ignore the time.Now() stamp).
	want := domainservices.NewOrderEventConverter().Convert(dto)
	got := audit.added[0]
	got.Timestamp = want.Timestamp
	assert.Equal(t, want, got)
}

// -----------------------------------------------------------------------------
// LocationEventService
// -----------------------------------------------------------------------------

func TestLocationEventService_ProcessEvent_Delivering_Testify(t *testing.T) {
	dest := &order.Location{Lat: 42.0, Lng: -72.0}
	seeded := seededOrder("NEW", dest)
	dto := locationDTO(seeded.Id, 42.5, -72.0)

	audit := &fakeAuditRepo{}
	orders := &fakeOrderRepo{getResult: &seeded}
	proc := &fakeProcRepo{}
	writer := &fakeWriter{}

	svc := NewLocationEventService(dto, audit, proc, orders, writer)
	require.NoError(t, svc.ProcessEvent())

	assert.Equal(t, dto.OrderID.String(), orders.gotOrderID)
	assert.Equal(t, "DELIVERING", seeded.Status)

	require.Len(t, orders.added, 1)
	assert.Equal(t, "DELIVERING", orders.added[0].Status)

	require.Len(t, audit.added, 1)
	assert.Equal(t, "location_event", audit.added[0].EventType)
	assert.Equal(t, "DELIVERING", audit.added[0].Status)

	require.Len(t, proc.added, 1)
	po := proc.added[0]
	assert.Equal(t, order.Location{Lat: 42.5, Lng: -72.0}, po.TimestampLocation)
	assert.Equal(t, *dest, po.FinalLocation)
	assert.Equal(t, "DELIVERING", po.OrderStatus)

	require.Len(t, writer.written, 1)
	n := writer.written[0]
	assert.Positive(t, n.TimeToDelivery)
	assert.Equal(t, po.TimestampLocation, n.ActualLocation)
	assert.Equal(t, po.FinalLocation, n.FinalDestination)
}

func TestLocationEventService_ProcessEvent_Delivered_Testify(t *testing.T) {
	dest := &order.Location{Lat: 42.0, Lng: -72.0}
	seeded := seededOrder("NEW", dest)
	dto := locationDTO(seeded.Id, dest.Lat, dest.Lng)

	orders := &fakeOrderRepo{getResult: &seeded}
	writer := &fakeWriter{}

	svc := NewLocationEventService(dto, &fakeAuditRepo{}, &fakeProcRepo{}, orders, writer)
	require.NoError(t, svc.ProcessEvent())

	assert.Equal(t, "DELIVERED", seeded.Status)
	assert.NotNil(t, seeded.DeliveredAt)
	require.Len(t, writer.written, 1)
	assert.Zero(t, writer.written[0].TimeToDelivery)
}

func TestLocationEventService_ProcessEvent_SideEffectGuards_Testify(t *testing.T) {
	repoErr := errors.New("boom")

	tests := []struct {
		name      string
		orders    *fakeOrderRepo
		zeroDTO   bool
		wantErr   bool
		wantErrIs error
	}{
		{name: "order_not_found", orders: &fakeOrderRepo{}, wantErr: false},
		{name: "repo_error", orders: &fakeOrderRepo{getErr: repoErr}, wantErr: true, wantErrIs: repoErr},
		{name: "missing_destination", orders: &fakeOrderRepo{getResult: ptr(seededOrder("NEW", nil))}, wantErr: true},
		{name: "zero_value_dto", orders: &fakeOrderRepo{}, zeroDTO: true, wantErr: true, wantErrIs: ErrMissingEventID},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			audit := &fakeAuditRepo{}
			proc := &fakeProcRepo{}
			writer := &fakeWriter{}

			dto := locationDTO(uuid.New(), 1, 1)
			if tt.zeroDTO {
				dto = zeroLocationDTO()
			}

			svc := NewLocationEventService(dto, audit, proc, tt.orders, writer)
			err := svc.ProcessEvent()

			if tt.wantErr {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
			}
			if tt.wantErrIs != nil {
				assert.ErrorIs(t, err, tt.wantErrIs)
			}

			assert.Empty(t, audit.added)
			assert.Empty(t, proc.added)
			assert.Empty(t, writer.written)
			assert.Empty(t, tt.orders.added)
		})
	}
}

func TestLocationEventService_ProcessEvent_InputIntegrity_Testify(t *testing.T) {
	dest := &order.Location{Lat: 42.0, Lng: -72.0}
	seeded := seededOrder("NEW", dest)
	dto := locationDTO(seeded.Id, 42.5, -72.0)
	wantEventID := dto.EventID
	orders := &fakeOrderRepo{getResult: &seeded}

	svc := NewLocationEventService(dto, &fakeAuditRepo{}, &fakeProcRepo{}, orders, &fakeWriter{})
	require.NoError(t, svc.ProcessEvent())

	assert.Equal(t, wantEventID, svc.dto.EventID)
	assert.True(t, svc.dto.Timestamp.IsZero())
}

func TestLocationEventService_ProcessEvent_MissingEventID_Testify(t *testing.T) {
	dest := &order.Location{Lat: 42.0, Lng: -72.0}
	seeded := seededOrder("NEW", dest)
	dto := locationDTO(seeded.Id, 42.5, -72.0)
	dto.EventID = uuid.Nil

	audit := &fakeAuditRepo{}
	orders := &fakeOrderRepo{getResult: &seeded}
	proc := &fakeProcRepo{}
	writer := &fakeWriter{}

	svc := NewLocationEventService(dto, audit, proc, orders, writer)
	err := svc.ProcessEvent()

	require.ErrorIs(t, err, ErrMissingEventID)
	assert.Empty(t, orders.gotOrderID, "repository must not be queried")
	assert.Empty(t, audit.added)
	assert.Empty(t, proc.added)
	assert.Empty(t, writer.written)
	assert.Empty(t, orders.added)
}

func TestLocationEventService_ProcessEvent_ValueCorrespondence_Testify(t *testing.T) {
	dest := &order.Location{Lat: 42.0, Lng: -72.0}
	actual := order.Location{Lat: 42.5, Lng: -72.0}
	seeded := seededOrder("NEW", dest)
	dto := locationDTO(seeded.Id, actual.Lat, actual.Lng)

	proc := &fakeProcRepo{}
	writer := &fakeWriter{}
	svc := NewLocationEventService(dto, &fakeAuditRepo{}, proc, &fakeOrderRepo{getResult: &seeded}, writer)
	require.NoError(t, svc.ProcessEvent())

	wantTTD, err := domainservices.NewDeliveryCalculator().CalculateTime(&actual, dest)
	require.NoError(t, err)
	assert.Equal(t, wantTTD, writer.written[0].TimeToDelivery)

	require.Len(t, proc.added, 1)
	eventId := proc.added[0].EventId
	ts := proc.added[0].Timestamp
	want := data.NewProcessedOrderFactory().CreateProcessedOrder(seeded, eventId, actual, ts)
	assert.Equal(t, want, proc.added[0])
}
