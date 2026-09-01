package messaging

import (
	"context"
	"encoding/json"
	"log"
	"os"
	"time"

	"github.com/LuisHFr15/delivery-tracker-distributed/internal/processor/app/dtos"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/sns"
)

// SnsNotificationWriter fulfils the NotificationWriter port on top of Amazon
// SNS. It keeps the same buffered-worker contract as KafkaWriteNotification so
// the processor's teardown logic is unchanged: under Lambda, StopWorker is what
// flushes the buffer before the invocation returns.
type SnsNotificationWriter struct {
	client   *sns.Client
	topicArn string
	buffer   chan *dtos.ProcessedOrderDTO
	done     chan struct{}
}

func NewSnsNotificationWriter(ctx context.Context) *SnsNotificationWriter {
	cfg, err := config.LoadDefaultConfig(ctx)
	if err != nil {
		log.Fatalf("unable to load SDK config, %v", err)
	}
	return &SnsNotificationWriter{
		client:   sns.NewFromConfig(cfg),
		topicArn: os.Getenv("SNS_PROCESSED_ORDERS_TOPIC_ARN"),
		buffer:   make(chan *dtos.ProcessedOrderDTO, 100),
		done:     make(chan struct{}),
	}
}

func (w *SnsNotificationWriter) Write(dto dtos.ProcessedOrderDTO) {
	w.buffer <- &dto
}

func (w *SnsNotificationWriter) RunWorker() {
	defer close(w.done)
	for dto := range w.buffer {
		value, err := json.Marshal(dto)
		if err != nil {
			log.Printf("error marshalling processed order: %v", err)
			continue
		}

		reqCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		_, err = w.client.Publish(reqCtx, &sns.PublishInput{
			TopicArn: aws.String(w.topicArn),
			Message:  aws.String(string(value)),
		})
		cancel()
		if err != nil {
			log.Printf("error publishing processed order to SNS: %v", err)
		}
	}
}

func (w *SnsNotificationWriter) StopWorker() error {
	close(w.buffer)
	<-w.done
	return nil
}
