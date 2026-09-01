package data

import (
	"time"

	"github.com/LuisHFr15/delivery-tracker-distributed/internal/processor/domain/models/order"
	"github.com/google/uuid"
)

// NotifiedMessage is the record persisted whenever the notifier acknowledges a
// processed order. Timestamp is the notification time and doubles as the sort
// key, so a single order keeps a full notification history.
type NotifiedMessage struct {
	OrderId        uuid.UUID      `dynamodbav:"OrderId"`
	ClientId       uuid.UUID      `dynamodbav:"ClientId"`
	OrderStatus    string         `dynamodbav:"OrderStatus"`
	ActualLocation order.Location `dynamodbav:"ActualLocation"`
	FinalLocation  order.Location `dynamodbav:"FinalLocation"`
	Timestamp      time.Time      `dynamodbav:"Timestamp"`
}
