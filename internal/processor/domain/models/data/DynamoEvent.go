package data

import (
	"time"

	"github.com/LuisHFr15/delivery-tracker-distributed/internal/processor/domain/models/order"
	"github.com/google/uuid"
)

type DynamoEvent struct {
	EventType string         `dynamo:"EventType"`
	Id        uuid.UUID      `dynamo:"Id"`
	OrderId   uuid.UUID      `dynamo:"OrderId"`
	Status    string         `dynamo:"Status"`
	Location  order.Location `dynamo:"Location"`
	Timestamp time.Time      `dynamo:"Timestamp"`
}
