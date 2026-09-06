package queues

import (
	"context"
	"encoding/json"
	"log"
	"os"
	"time"

	"github.com/LuisHFr15/delivery-tracker-distributed/internal/ingester/app/dtos"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/sns"
)

// SnsPublisher fulfils the EventPublisher port on top of Amazon SNS. It is the
// serverless counterpart of KafkaPublisher: instead of buffering to a Kafka
// writer it publishes each event straight to its topic, which is the right
// semantic under Lambda where the invocation must finish the work before it
// returns.
type SnsPublisher struct {
	client        *sns.Client
	orderTopic    string
	locationTopic string
}

func NewSnsPublisher(ctx context.Context) *SnsPublisher {
	cfg, err := config.LoadDefaultConfig(ctx)
	if err != nil {
		log.Fatalf("unable to load SDK config, %v", err)
	}
	return &SnsPublisher{
		client:        sns.NewFromConfig(cfg),
		orderTopic:    os.Getenv("SNS_ORDER_TOPIC_ARN"),
		locationTopic: os.Getenv("SNS_LOCATION_TOPIC_ARN"),
	}
}

func (p *SnsPublisher) PublishOrder(dto dtos.OrderEventDTO) {
	p.publish(p.orderTopic, dto, dto.Order.ID.String())
}

func (p *SnsPublisher) PublishLocation(dto dtos.LocationEventDTO) {
	p.publish(p.locationTopic, dto, dto.OrderID.String())
}

func (p *SnsPublisher) publish(topicArn string, payload any, groupKey string) {
	value, err := json.Marshal(payload)
	if err != nil {
		log.Printf("failed to marshal event for %s: %v", topicArn, err)
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	_, err = p.client.Publish(ctx, &sns.PublishInput{
		TopicArn: aws.String(topicArn),
		Message:  aws.String(string(value)),
	})
	log.Printf("published %s to respective topic successfuly", string(value), topicArn)
	if err != nil {
		log.Printf("failed to publish event for order %s to %s: %v", groupKey, topicArn, err)
	}
}
