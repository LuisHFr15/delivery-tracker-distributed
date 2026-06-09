package orders

import "github.com/google/uuid"

type Client struct {
	Id        uuid.UUID  `json:"id"`
	Addresses []Location `json:"addresses"`
}
