package services

import (
	"errors"
	"testing"

	"github.com/LuisHFr15/delivery-tracker-distributed/internal/processor/domain/models/data"
	"github.com/LuisHFr15/delivery-tracker-distributed/internal/processor/domain/models/order"
	domainservices "github.com/LuisHFr15/delivery-tracker-distributed/internal/processor/domain/services"
	"github.com/google/uuid"
)

// -----------------------------------------------------------------------------
// OrderEventService — Type 1 (E2E) + Type 2 (collaboration), stdlib style.
// -----------------------------------------------------------------------------

func TestOrderEventService_ConvertEvent_Success(t *testing.T) {
	dto := validOrderEventDTO()
	audit := &fakeAuditRepo{}
	orders := &fakeOrderRepo{}

	svc := NewOrderEventService(dto, audit, orders)
	if err := svc.ConvertEvent(); err != nil {
		t.Fatalf("ConvertEvent() unexpected error = %v", err)
	}

	// One audit event, one persisted order.
	if len(audit.added) != 1 {
		t.Fatalf("audit events = %d, want 1", len(audit.added))
	}
	if len(orders.added) != 1 {
		t.Fatalf("persisted orders = %d, want 1", len(orders.added))
	}

	// Audit event values must correspond to the source DTO.
	ev := audit.added[0]
	if ev.EventType != "order_event" {
		t.Errorf("audit EventType = %q, want %q", ev.EventType, "order_event")
	}
	if ev.Status != dto.TransactionType {
		t.Errorf("audit Status = %q, want %q", ev.Status, dto.TransactionType)
	}
	if ev.OrderId != dto.Order.ID {
		t.Errorf("audit OrderId = %v, want %v", ev.OrderId, dto.Order.ID)
	}
	if ev.Location != *dto.Order.Destination {
		t.Errorf("audit Location = %v, want %v", ev.Location, *dto.Order.Destination)
	}

	// Persisted order values.
	got := orders.added[0]
	if got.Id != dto.Order.ID {
		t.Errorf("order Id = %v, want %v", got.Id, dto.Order.ID)
	}
	if got.Client.Id != dto.Order.ClientID {
		t.Errorf("order Client.Id = %v, want %v", got.Client.Id, dto.Order.ClientID)
	}
	if got.Status != dto.Order.Status {
		t.Errorf("order Status = %q, want %q", got.Status, dto.Order.Status)
	}
	if len(got.Products) != len(dto.Order.Products) {
		t.Fatalf("order Products len = %d, want %d", len(got.Products), len(dto.Order.Products))
	}
	if got.Products[0].Product.Id != dto.Order.Products[0].ProductID {
		t.Errorf("order product id = %v, want %v", got.Products[0].Product.Id, dto.Order.Products[0].ProductID)
	}
	if got.Products[0].Quantity != dto.Order.Products[0].Quantity {
		t.Errorf("order product qty = %d, want %d", got.Products[0].Quantity, dto.Order.Products[0].Quantity)
	}
}

// Documents behaviour #1: on a malformed order (nil destination) ConvertEvent
// still emits an audit event before validation fails, but persists no order.
func TestOrderEventService_ConvertEvent_MissingDestination(t *testing.T) {
	dto := validOrderEventDTO()
	dto.Order.Destination = nil

	audit := &fakeAuditRepo{}
	orders := &fakeOrderRepo{}

	svc := NewOrderEventService(dto, audit, orders)
	err := svc.ConvertEvent()

	if !errors.Is(err, order.ErrMissingDestination) {
		t.Fatalf("ConvertEvent() error = %v, want ErrMissingDestination", err)
	}
	if len(audit.added) != 1 {
		t.Errorf("audit events = %d, want 1 (audit is written before validation)", len(audit.added))
	}
	if len(orders.added) != 0 {
		t.Errorf("persisted orders = %d, want 0 (order failed validation)", len(orders.added))
	}
}

// A nil EventID is rejected after the audit is written (audit-first) but before
// the order is converted/persisted.
func TestOrderEventService_ConvertEvent_MissingEventID(t *testing.T) {
	dto := validOrderEventDTO()
	dto.EventID = uuid.Nil

	audit := &fakeAuditRepo{}
	orders := &fakeOrderRepo{}

	svc := NewOrderEventService(dto, audit, orders)
	err := svc.ConvertEvent()

	if !errors.Is(err, ErrMissingEventID) {
		t.Fatalf("ConvertEvent() error = %v, want ErrMissingEventID", err)
	}
	if len(audit.added) != 1 {
		t.Errorf("audit events = %d, want 1 (audit is written before validation)", len(audit.added))
	}
	if len(orders.added) != 0 {
		t.Errorf("persisted orders = %d, want 0 (event rejected)", len(orders.added))
	}
}

func TestOrderEventService_ConvertEvent_EmptyProducts(t *testing.T) {
	dto := validOrderEventDTO()
	dto.Order.Products = nil
	audit := &fakeAuditRepo{}
	orders := &fakeOrderRepo{}

	svc := NewOrderEventService(dto, audit, orders)
	if err := svc.ConvertEvent(); err != nil {
		t.Fatalf("ConvertEvent() unexpected error = %v", err)
	}
	if len(orders.added) != 1 {
		t.Fatalf("persisted orders = %d, want 1", len(orders.added))
	}
	if len(orders.added[0].Products) != 0 {
		t.Errorf("order Products len = %d, want 0", len(orders.added[0].Products))
	}
}

// Struct integrity: the caller's DTO must not be mutated by ConvertEvent
// (the service takes it by value).
func TestOrderEventService_ConvertEvent_InputIntegrity(t *testing.T) {
	dto := validOrderEventDTO()
	snapshot := dto
	snapshotDest := *dto.Order.Destination

	svc := NewOrderEventService(dto, &fakeAuditRepo{}, &fakeOrderRepo{})
	if err := svc.ConvertEvent(); err != nil {
		t.Fatalf("ConvertEvent() unexpected error = %v", err)
	}

	if dto.EventID != snapshot.EventID {
		t.Errorf("caller DTO EventID mutated: got %v, want %v", dto.EventID, snapshot.EventID)
	}
	if dto.Order.Status != snapshot.Order.Status {
		t.Errorf("caller DTO Status mutated: got %q, want %q", dto.Order.Status, snapshot.Order.Status)
	}
	if *dto.Order.Destination != snapshotDest {
		t.Errorf("caller DTO Destination mutated: got %v, want %v", *dto.Order.Destination, snapshotDest)
	}
}

// Value correspondence + Type 2 collaboration: with a valid EventID the audit
// id and the persisted order EventId agree with each other and with dto.EventID
// (no id divergence), and the audit event equals what OrderEventConverter
// produces independently (ignoring the time.Now() timestamp).
func TestOrderEventService_ConvertEvent_ValueCorrespondence(t *testing.T) {
	dto := validOrderEventDTO()
	audit := &fakeAuditRepo{}
	orders := &fakeOrderRepo{}

	svc := NewOrderEventService(dto, audit, orders)
	if err := svc.ConvertEvent(); err != nil {
		t.Fatalf("ConvertEvent() unexpected error = %v", err)
	}

	// No divergence: audit id == persisted order EventId == dto.EventID.
	if audit.added[0].Id != dto.EventID {
		t.Errorf("audit Id = %v, want %v (dto.EventID)", audit.added[0].Id, dto.EventID)
	}
	if orders.added[0].EventId != dto.EventID {
		t.Errorf("persisted order EventId = %v, want %v (dto.EventID)", orders.added[0].EventId, dto.EventID)
	}

	// Type 2: the service must delegate to OrderEventConverter. Rebuild the
	// event independently and compare everything except the wall-clock stamp.
	want := domainservices.NewOrderEventConverter().Convert(dto)
	got := audit.added[0]
	got.Timestamp = want.Timestamp // Convert uses time.Now(); ignore it.
	if got != want {
		t.Errorf("audit event = %+v, want (converter output) %+v", got, want)
	}
}

// -----------------------------------------------------------------------------
// LocationEventService — Type 1 (E2E) + Type 2 (collaboration), stdlib style.
// -----------------------------------------------------------------------------

func TestLocationEventService_ProcessEvent_Delivering(t *testing.T) {
	dest := &order.Location{Lat: 42.0, Lng: -72.0}
	seeded := seededOrder("NEW", dest)
	dto := locationDTO(seeded.Id, 42.5, -72.0) // different from destination

	audit := &fakeAuditRepo{}
	orders := &fakeOrderRepo{getResult: &seeded}
	proc := &fakeProcRepo{}
	writer := &fakeWriter{}

	svc := NewLocationEventService(dto, audit, proc, orders, writer)
	if err := svc.ProcessEvent(); err != nil {
		t.Fatalf("ProcessEvent() unexpected error = %v", err)
	}

	// Repo was queried with the right key.
	if orders.gotOrderID != dto.OrderID.String() {
		t.Errorf("GetByOrderId key = %q, want %q", orders.gotOrderID, dto.OrderID.String())
	}

	// Status transition: NEW -> DELIVERING.
	if seeded.Status != "DELIVERING" {
		t.Errorf("order Status = %q, want DELIVERING", seeded.Status)
	}
	if len(orders.added) != 1 || orders.added[0].Status != "DELIVERING" {
		t.Errorf("persisted order status wrong: %+v", orders.added)
	}

	// Audit event.
	if len(audit.added) != 1 {
		t.Fatalf("audit events = %d, want 1", len(audit.added))
	}
	if audit.added[0].EventType != "location_event" || audit.added[0].Status != "DELIVERING" {
		t.Errorf("audit event = %+v, want location_event/DELIVERING", audit.added[0])
	}

	// Processed order.
	if len(proc.added) != 1 {
		t.Fatalf("processed orders = %d, want 1", len(proc.added))
	}
	po := proc.added[0]
	if po.TimestampLocation != (order.Location{Lat: 42.5, Lng: -72.0}) {
		t.Errorf("processed TimestampLocation = %v, want actual location", po.TimestampLocation)
	}
	if po.FinalLocation != *dest {
		t.Errorf("processed FinalLocation = %v, want %v", po.FinalLocation, *dest)
	}
	if po.OrderStatus != "DELIVERING" {
		t.Errorf("processed OrderStatus = %q, want DELIVERING", po.OrderStatus)
	}

	// Notification.
	if len(writer.written) != 1 {
		t.Fatalf("notifications = %d, want 1", len(writer.written))
	}
	n := writer.written[0]
	if n.TimeToDelivery <= 0 {
		t.Errorf("notification TimeToDelivery = %v, want > 0", n.TimeToDelivery)
	}
	if n.ActualLocation != po.TimestampLocation || n.FinalDestination != po.FinalLocation {
		t.Errorf("notification locations mismatch: %+v vs processed %+v", n, po)
	}
}

func TestLocationEventService_ProcessEvent_Delivered(t *testing.T) {
	dest := &order.Location{Lat: 42.0, Lng: -72.0}
	seeded := seededOrder("NEW", dest)
	dto := locationDTO(seeded.Id, dest.Lat, dest.Lng) // exactly at destination

	orders := &fakeOrderRepo{getResult: &seeded}
	writer := &fakeWriter{}

	svc := NewLocationEventService(dto, &fakeAuditRepo{}, &fakeProcRepo{}, orders, writer)
	if err := svc.ProcessEvent(); err != nil {
		t.Fatalf("ProcessEvent() unexpected error = %v", err)
	}

	if seeded.Status != "DELIVERED" {
		t.Errorf("order Status = %q, want DELIVERED", seeded.Status)
	}
	if seeded.DeliveredAt == nil {
		t.Errorf("order DeliveredAt = nil, want set")
	}
	if len(writer.written) != 1 || writer.written[0].TimeToDelivery != 0 {
		t.Errorf("notification TimeToDelivery = %v, want 0 (arrived)", writer.written)
	}
}

func TestLocationEventService_ProcessEvent_SideEffectGuards(t *testing.T) {
	repoErr := errors.New("boom")

	tests := []struct {
		name       string
		orders     *fakeOrderRepo
		wantErr    bool
		wantErrIs  error
	}{
		{
			name:    "order_not_found",
			orders:  &fakeOrderRepo{getResult: nil, getErr: nil},
			wantErr: false,
		},
		{
			name:      "repo_error",
			orders:    &fakeOrderRepo{getResult: nil, getErr: repoErr},
			wantErr:   true,
			wantErrIs: repoErr,
		},
		{
			name:    "missing_destination",
			orders:  &fakeOrderRepo{getResult: ptr(seededOrder("NEW", nil))},
			wantErr: true,
		},
		{
			name:      "zero_value_dto",
			orders:    &fakeOrderRepo{getResult: nil, getErr: nil}, // rejected before repo is queried
			wantErr:   true,
			wantErrIs: ErrMissingEventID,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			audit := &fakeAuditRepo{}
			proc := &fakeProcRepo{}
			writer := &fakeWriter{}

			var dto = locationDTO(uuid.New(), 1, 1)
			if tt.name == "zero_value_dto" {
				dto = zeroLocationDTO()
			}

			svc := NewLocationEventService(dto, audit, proc, tt.orders, writer)
			err := svc.ProcessEvent()

			if (err != nil) != tt.wantErr {
				t.Fatalf("ProcessEvent() error = %v, wantErr %v", err, tt.wantErr)
			}
			if tt.wantErrIs != nil && !errors.Is(err, tt.wantErrIs) {
				t.Errorf("ProcessEvent() error = %v, want to wrap %v", err, tt.wantErrIs)
			}

			// None of these paths may produce audit/processed/notification writes.
			if len(audit.added) != 0 || len(proc.added) != 0 || len(writer.written) != 0 || len(tt.orders.added) != 0 {
				t.Errorf("unexpected side effects: audit=%d proc=%d writer=%d orderAdd=%d",
					len(audit.added), len(proc.added), len(writer.written), len(tt.orders.added))
			}
		})
	}
}

// Struct integrity: ProcessEvent works on a copy of its DTO; the service's own
// dto field (and thus the caller's input) must keep its original values.
func TestLocationEventService_ProcessEvent_InputIntegrity(t *testing.T) {
	dest := &order.Location{Lat: 42.0, Lng: -72.0}
	seeded := seededOrder("NEW", dest)
	dto := locationDTO(seeded.Id, 42.5, -72.0)
	wantEventID := dto.EventID
	orders := &fakeOrderRepo{getResult: &seeded}

	svc := NewLocationEventService(dto, &fakeAuditRepo{}, &fakeProcRepo{}, orders, &fakeWriter{})
	if err := svc.ProcessEvent(); err != nil {
		t.Fatalf("ProcessEvent() unexpected error = %v", err)
	}

	if svc.dto.EventID != wantEventID {
		t.Errorf("service dto.EventID mutated to %v, want %v", svc.dto.EventID, wantEventID)
	}
	if !svc.dto.Timestamp.IsZero() {
		t.Errorf("service dto.Timestamp mutated to %v, want zero", svc.dto.Timestamp)
	}
}

// A nil EventID is rejected up front, before the repository is queried and
// before any side effect can occur.
func TestLocationEventService_ProcessEvent_MissingEventID(t *testing.T) {
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

	if !errors.Is(err, ErrMissingEventID) {
		t.Fatalf("ProcessEvent() error = %v, want ErrMissingEventID", err)
	}
	if orders.gotOrderID != "" {
		t.Errorf("GetByOrderId was called with %q, want no call", orders.gotOrderID)
	}
	if len(audit.added) != 0 || len(proc.added) != 0 || len(writer.written) != 0 || len(orders.added) != 0 {
		t.Errorf("unexpected side effects: audit=%d proc=%d writer=%d orderAdd=%d",
			len(audit.added), len(proc.added), len(writer.written), len(orders.added))
	}
}

// Type 2 collaboration: the values reaching the repos/writer must equal what the
// domain services produce independently.
func TestLocationEventService_ProcessEvent_ValueCorrespondence(t *testing.T) {
	dest := &order.Location{Lat: 42.0, Lng: -72.0}
	actual := order.Location{Lat: 42.5, Lng: -72.0}
	seeded := seededOrder("NEW", dest)
	dto := locationDTO(seeded.Id, actual.Lat, actual.Lng)

	proc := &fakeProcRepo{}
	writer := &fakeWriter{}
	svc := NewLocationEventService(dto, &fakeAuditRepo{}, proc, &fakeOrderRepo{getResult: &seeded}, writer)
	if err := svc.ProcessEvent(); err != nil {
		t.Fatalf("ProcessEvent() unexpected error = %v", err)
	}

	// DeliveryCalculator delegation: notification ETA equals the calculator's.
	wantTTD, err := domainservices.NewDeliveryCalculator().CalculateTime(&actual, dest)
	if err != nil {
		t.Fatalf("calculator error = %v", err)
	}
	if writer.written[0].TimeToDelivery != wantTTD {
		t.Errorf("notification TimeToDelivery = %v, want %v (calculator)", writer.written[0].TimeToDelivery, wantTTD)
	}

	// ProcessedOrderFactory delegation: processed order equals the factory's
	// output for the mutated order (status already DELIVERING at this point).
	eventId := proc.added[0].EventId
	ts := proc.added[0].Timestamp
	want := data.NewProcessedOrderFactory().CreateProcessedOrder(seeded, eventId, actual, ts)
	got := proc.added[0]
	if got.OrderId != want.OrderId || got.ClientId != want.ClientId ||
		got.OrderStatus != want.OrderStatus || got.TimestampLocation != want.TimestampLocation ||
		got.FinalLocation != want.FinalLocation {
		t.Errorf("processed order = %+v, want (factory output) %+v", got, want)
	}
}
