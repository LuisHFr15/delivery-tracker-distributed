package dynamo

import (
	"context"
	"log"
	"os"
	"time"

	"github.com/LuisHFr15/delivery-tracker-distributed/internal/processor/domain/models/data"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/feature/dynamodb/attributevalue"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
)

type DynamoAuditingEventRepository struct {
	tableName string
	client    *dynamodb.Client
	buffer    chan *data.DynamoEvent
}

func NewDynamoAuditingEventRepository(ctx context.Context) *DynamoAuditingEventRepository {
	cfg, err := config.LoadDefaultConfig(ctx)
	if err != nil {
		log.Fatalf("unable to load SDK config, %v", err)
	}
	return &DynamoAuditingEventRepository{
		tableName: os.Getenv("DYNAMODB_AUDITING_TABLE_NAME"),
		client:    dynamodb.NewFromConfig(cfg),
		buffer:    make(chan *data.DynamoEvent, 5),
	}
}

func (d *DynamoAuditingEventRepository) Add(cls data.DynamoEvent) {
	d.buffer <- &cls
}

func (d *DynamoAuditingEventRepository) RunWorker() {
	defer close(d.buffer)
	for cls := range d.buffer {
		item, err := attributevalue.MarshalMap(cls)

		if err != nil {
			log.Printf("Error serializing DynamoAuditingEvent item: %s", err)
		}

		reqCtx, cancel := context.WithTimeout(context.Background(), time.Second*5)

		// if the parent app context is currently shutting down, the dynamodb will still conclude the operation since it has the graceful shutdown
		_, err = d.client.PutItem(reqCtx, &dynamodb.PutItemInput{
			TableName: aws.String(d.tableName),
			Item:      item,
		})

		cancel()
		if err != nil {
			log.Println("Error putting item in DynamoDb", err)
		}
	}
	log.Println("DynamoAuditingEventRepository finished")
}

func (d *DynamoAuditingEventRepository) StopWorker() {
	close(d.buffer)
}
