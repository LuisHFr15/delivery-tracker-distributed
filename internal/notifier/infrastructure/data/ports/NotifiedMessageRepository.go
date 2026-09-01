package ports

import "github.com/LuisHFr15/delivery-tracker-distributed/internal/notifier/domain/models/data"

type NotifiedMessageRepository interface {
	Add(msg data.NotifiedMessage)
	RunWorker()
	StopWorker() error
}
