package ports

import (
	"context"
)

type KafkaReader interface {
	Read(ctx context.Context)
}
