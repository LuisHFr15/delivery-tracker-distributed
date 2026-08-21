package order

import "github.com/google/uuid"

type Client struct {
	Id        uuid.UUID  `json:"id" dynamodbav:"Id"`
	Addresses []Location `json:"addresses" dynamodbav:"Addresses"`
}
