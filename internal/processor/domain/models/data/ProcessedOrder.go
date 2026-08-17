package data

import (
	"time"

	"github.com/LuisHFr15/delivery-tracker-distributed/internal/processor/domain/models/order"
	"github.com/google/uuid"
)

type ProcessedOrder struct {
	OrderId           uuid.UUID       `dynamodbav:"OrderId"`
	EventId           uuid.UUID       `dynamodbav:"EventId"`
	ClientId          uuid.UUID       `dynamodbav:"ClientId"`
	OrderStatus       string          `dynamodbav:"OrderStatus"`
	Products          []order.Product `dynamodbav:"Products"`
	TimestampLocation order.Location  `dynamodbav:"TimestampLocation"`
	FinalLocation     order.Location  `dynamodbav:"FinalLocation"`
	Timestamp         time.Time       `dynamodbav:"Timestamp"`
}
