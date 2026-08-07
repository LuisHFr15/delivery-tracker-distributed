package dynamo

import (
	"context"
	"log"
	"time"

	"github.com/LuisHFr15/delivery-tracker-distributed/internal/processor/domain/models/data"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/feature/dynamodb/attributevalue"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
)

type DynamoAuditingEventRepository struct {
	tableName string
	client    *dynamodb.Client
	buffer    chan *data.DynamoEvent
}

func NewDynamoAuditingEventRepository(tableName string, client *dynamodb.Client) *DynamoAuditingEventRepository {
	return &DynamoAuditingEventRepository{
		tableName: tableName,
		client:    client,
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
