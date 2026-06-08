package orders

import (
	"github.com/google/uuid"
)

type Product struct {
	Id           uuid.UUID `json:"product_id"`
	ProductPrice int32     `json:"product_price"`
}
