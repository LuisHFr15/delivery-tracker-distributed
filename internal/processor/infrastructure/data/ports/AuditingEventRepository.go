package ports

import (
	"github.com/LuisHFr15/delivery-tracker-distributed/internal/processor/domain/models/data"
)

type AuditingEventRepository interface {
	Add(cls data.DynamoEvent)
	RunWorker()
	StopWorker()
}
