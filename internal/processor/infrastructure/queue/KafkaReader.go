package infrastructure

import (
	"context"
)

type KafkaReader interface {
	Read(ctx context.Context)
}
