package ports

import "github.com/LuisHFr15/delivery-tracker-distributed/internal/processor/app/dtos"

type NotificationWriter interface {
	Write(dto dtos.ProcessedOrderDTO)
	RunWorker()
	StopWorker() error
}
