package data

import (
	"time"

	"github.com/LuisHFr15/delivery-tracker-distributed/internal/processor/domain/models/order"
	"github.com/google/uuid"
)

type DynamoEvent struct {
	TransactionType string         `dynamo:"TransactionType"`
	Id              uuid.UUID      `dynamo:"Id"`
	OrderId         string         `dynamo:"OrderId"`
	Status          string         `dynamo:"Status"`
	Location        order.Location `dynamo:"Location"`
	Timestamp       time.Time      `dynamo:"Timestamp"`
}
