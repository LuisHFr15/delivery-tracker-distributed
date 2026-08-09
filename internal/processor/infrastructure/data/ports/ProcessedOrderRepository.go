package ports

import (
	"github.com/LuisHFr15/delivery-tracker-distributed/internal/processor/domain/models/data"
)

type ProcessedOrderRepository interface {
	Add(dto data.ProcessedOrder)
	RunWorker()
	StopWorker()
}
