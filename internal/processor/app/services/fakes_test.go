package services

import (
	"context"

	"github.com/LuisHFr15/delivery-tracker-distributed/internal/processor/app/dtos"
	"github.com/LuisHFr15/delivery-tracker-distributed/internal/processor/domain/models/data"
	"github.com/LuisHFr15/delivery-tracker-distributed/internal/processor/domain/models/order"
	"github.com/google/uuid"
)

// ptr is a small generic helper mirroring the one used in DeliveryUpdater_test.go.
func ptr[T any](v T) *T { return &v }

// validOrderEventDTO builds a well-formed order event (event id, destination
// set, one product) that passes ConvertEvent's nil-EventID guard.
func validOrderEventDTO() dtos.OrderEventDTO {
	return dtos.OrderEventDTO{
		EventID: uuid.New(),
		Order: dtos.OrderDTO{
			ID:       uuid.New(),
			ClientID: uuid.New(),
			Products: []dtos.OrderItemDTO{
				{ProductID: uuid.New(), Name: "widget", Price: 100, Quantity: 2},
			},
			Destination: &order.Location{Lat: 42.0, Lng: -72.0},
			Status:      "NEW",
		},
		TransactionType: "CREATED",
	}
}

// zeroOrderEventDTO is the explicit zero value (nil destination, empty fields),
// kept as a named helper to make the "empty struct" edge cases self-documenting.
func zeroOrderEventDTO() dtos.OrderEventDTO { return dtos.OrderEventDTO{} }

// locationDTO builds a location event pointing at the given order id / coords.
func locationDTO(orderID uuid.UUID, lat, lng float64) dtos.LocationEventDTO {
	return dtos.LocationEventDTO{
		EventID:   uuid.New(),
		OrderID:   orderID,
		Latitude:  lat,
		Longitude: lng,
	}
}

// zeroLocationDTO is the explicit zero value (OrderID == uuid.Nil, 0/0 coords).
func zeroLocationDTO() dtos.LocationEventDTO { return dtos.LocationEventDTO{} }

// seededOrder builds an order.Order as if it had been fetched from the
// repository (the state LocationEventService.ProcessEvent mutates).
func seededOrder(status string, dest *order.Location) order.Order {
	return order.Order{
		Id:          uuid.New(),
		EventId:     uuid.New(),
		Client:      order.Client{Id: uuid.New()},
		Products:    []order.OrderItem{{Product: order.Product{Id: uuid.New(), Name: "widget", ProductPrice: 100}, Quantity: 2}},
		Destination: dest,
		Status:      status,
	}
}

// fakeAuditRepo is a synchronous in-memory spy for ports.AuditingEventRepository.
// Add records every event so tests can assert both count and content.
type fakeAuditRepo struct {
	added []data.DynamoEvent
}

func (f *fakeAuditRepo) Add(cls data.DynamoEvent) { f.added = append(f.added, cls) }
func (f *fakeAuditRepo) RunWorker()               {}
func (f *fakeAuditRepo) StopWorker() error        { return nil }

// fakeOrderRepo is a synchronous in-memory spy for ports.OrderRepository.
// GetByOrderId is configurable via getResult/getErr; the requested id is
// captured in gotOrderID so tests can assert the service queried the right key.
type fakeOrderRepo struct {
	added     []order.Order
	getResult *order.Order
	getErr    error
	gotOrderID string
}

func (f *fakeOrderRepo) Add(o order.Order) { f.added = append(f.added, o) }
func (f *fakeOrderRepo) RunWorker()        {}
func (f *fakeOrderRepo) StopWorker() error { return nil }
func (f *fakeOrderRepo) GetByOrderId(ctx context.Context, orderId string) (*order.Order, error) {
	f.gotOrderID = orderId
	return f.getResult, f.getErr
}

// fakeProcRepo is a synchronous in-memory spy for ports.ProcessedOrderRepository.
type fakeProcRepo struct {
	added []data.ProcessedOrder
}

func (f *fakeProcRepo) Add(dto data.ProcessedOrder) { f.added = append(f.added, dto) }
func (f *fakeProcRepo) RunWorker()                  {}
func (f *fakeProcRepo) StopWorker() error           { return nil }
func (f *fakeProcRepo) GetLatestByOrderId(ctx context.Context, orderId string) (*data.ProcessedOrder, error) {
	return nil, nil
}

// fakeWriter is a synchronous in-memory spy for messaging ports.NotificationWriter.
type fakeWriter struct {
	written []dtos.ProcessedOrderDTO
}

func (f *fakeWriter) Write(dto dtos.ProcessedOrderDTO) { f.written = append(f.written, dto) }
func (f *fakeWriter) RunWorker()                       {}
func (f *fakeWriter) StopWorker() error                { return nil }
