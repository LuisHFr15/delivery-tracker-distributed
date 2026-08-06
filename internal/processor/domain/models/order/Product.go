package order

import (
	"github.com/google/uuid"
)

type Product struct {
	Id           uuid.UUID `json:"product_id"`
	Name         string    `json:"name"`
	ProductPrice int32     `json:"product_price"`
}
