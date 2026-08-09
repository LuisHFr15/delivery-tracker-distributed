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

type DynamoProcessedOrderRepository struct {
	tableName string
	client    *dynamodb.Client
	buffer    chan *data.ProcessedOrder
}

func NewDynamoProcessedOrderRepository(ctx context.Context) *DynamoProcessedOrderRepository {
	cfg, err := config.LoadDefaultConfig(ctx)

	if err != nil {
		log.Fatalf("unable to load SDK config, %v", err)
	}
	return &DynamoProcessedOrderRepository{
		tableName: os.Getenv("DYNAMODB_PROCESSED_ORDER_TABLE_NAME"),
		client:    dynamodb.NewFromConfig(cfg),
		buffer:    make(chan *data.ProcessedOrder),
	}
}

func (d *DynamoProcessedOrderRepository) Add(cls data.ProcessedOrder) {
	d.buffer <- &cls
}

func (d *DynamoProcessedOrderRepository) RunWorker() {
	defer close(d.buffer)
	for cls := range d.buffer {
		item, err := attributevalue.MarshalMap(cls)

		if err != nil {
			log.Printf("Error serializing ProcessedOrder item: %s", err)
		}

		reqCtx, cancel := context.WithTimeout(context.Background(), time.Second*5)

		_, err = d.client.PutItem(reqCtx, &dynamodb.PutItemInput{
			TableName: aws.String(d.tableName),
			Item:      item,
		})

		cancel()
		if err != nil {
			log.Printf("Error putting ProcessedOrder in DynamoDb: %s", err)
		}
	}
}

func (d *DynamoProcessedOrderRepository) StopWorker() {
	close(d.buffer)
}
