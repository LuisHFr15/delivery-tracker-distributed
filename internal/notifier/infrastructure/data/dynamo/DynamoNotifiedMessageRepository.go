package dynamo

import (
	"context"
	"log"
	"os"
	"time"

	"github.com/LuisHFr15/delivery-tracker-distributed/internal/notifier/domain/models/data"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/feature/dynamodb/attributevalue"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
)

type DynamoNotifiedMessageRepository struct {
	tableName string
	client    *dynamodb.Client
	buffer    chan *data.NotifiedMessage
	done      chan struct{}
}

func NewDynamoNotifiedMessageRepository(ctx context.Context) *DynamoNotifiedMessageRepository {
	cfg, err := config.LoadDefaultConfig(ctx)
	if err != nil {
		log.Fatalf("unable to load SDK config, %v", err)
	}
	return &DynamoNotifiedMessageRepository{
		tableName: os.Getenv("DYNAMODB_NOTIFIED_MESSAGES_TABLE_NAME"),
		client:    dynamodb.NewFromConfig(cfg),
		buffer:    make(chan *data.NotifiedMessage),
		done:      make(chan struct{}),
	}
}

func (d *DynamoNotifiedMessageRepository) Add(msg data.NotifiedMessage) {
	d.buffer <- &msg
}

func (d *DynamoNotifiedMessageRepository) RunWorker() {
	defer close(d.done)
	for msg := range d.buffer {
		item, err := attributevalue.MarshalMapWithOptions(msg, func(o *attributevalue.EncoderOptions) {
			o.UseEncodingMarshalers = true
		})
		if err != nil {
			log.Printf("Error serializing NotifiedMessage item: %s", err)
			continue
		}

		reqCtx, cancel := context.WithTimeout(context.Background(), time.Second*5)
		_, err = d.client.PutItem(reqCtx, &dynamodb.PutItemInput{
			TableName: aws.String(d.tableName),
			Item:      item,
		})
		cancel()
		if err != nil {
			log.Printf("Error putting NotifiedMessage in DynamoDb: %s", err)
		}
	}
}

func (d *DynamoNotifiedMessageRepository) StopWorker() error {
	close(d.buffer)
	<-d.done
	return nil
}
