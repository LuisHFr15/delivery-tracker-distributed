package data

import (
	"time"

	"github.com/LuisHFr15/delivery-tracker-distributed/internal/processor/domain/models/order"
	"github.com/google/uuid"
)

type ProcessedOrder struct {
	OrderId           uuid.UUID       `dynamo:"OrderId"`
	EventId           uuid.UUID       `dynamo:"EventId"`
	ClientId          uuid.UUID       `dynamo:"ClientId"`
	OrderStatus       string          `dynamo:"OrderStatus"`
	Products          []order.Product `dynamo:"Products"`
	TimestampLocation order.Location  `dynamo:"TimestampLocation"`
	FinalLocation     order.Location  `dynamo:"FinalLocation"`
	Timestamp         time.Time       `dynamo:"Timestamp"`
}
