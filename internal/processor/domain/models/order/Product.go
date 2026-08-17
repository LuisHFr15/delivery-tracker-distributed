package order

import (
	"github.com/google/uuid"
)

type Product struct {
	Id           uuid.UUID `json:"product_id" dynamodbav:"Id"`
	Name         string    `json:"name" dynamodbav:"Name"`
	ProductPrice int32     `json:"product_price" dynamodbav:"ProductPrice"`
}
