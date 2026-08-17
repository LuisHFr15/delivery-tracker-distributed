package data

import (
	"time"

	"github.com/LuisHFr15/delivery-tracker-distributed/internal/processor/domain/models/order"
	"github.com/google/uuid"
)

type DynamoEvent struct {
	EventType string         `dynamodbav:"EventType"`
	Id        uuid.UUID      `dynamodbav:"Id"`
	OrderId   uuid.UUID      `dynamodbav:"OrderId"`
	Status    string         `dynamodbav:"Status"`
	Location  order.Location `dynamodbav:"Location"`
	Timestamp time.Time      `dynamodbav:"Timestamp"`
}
