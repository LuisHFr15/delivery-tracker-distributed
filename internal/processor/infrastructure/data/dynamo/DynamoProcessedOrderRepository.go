package dynamo

import (
	"context"
	"log"
	"os"

	"github.com/LuisHFr15/delivery-tracker-distributed/internal/processor/domain/models/data"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/feature/dynamodb/attributevalue"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
)

type DynamoProcessedOrderRepository struct {
	tableName string
	client    *dynamodb.Client
}

func NewDynamoProcessedOrderRepository(tableName string, client *dynamodb.Client) *DynamoProcessedOrderRepository {
	return &DynamoProcessedOrderRepository{
		tableName: tableName,
		client:    client,
	}
}

func (d *DynamoProcessedOrderRepository) Add(ctx context.Context, cls data.ProcessedOrder) error {
	item, err := attributevalue.MarshalMap(cls)
	if err != nil {
		return err
	}

	_, err = d.client.PutItem(ctx, &dynamodb.PutItemInput{
		TableName: aws.String(os.Getenv("TABLE_NAME")),
		Item:      item,
	})

	if err != nil {
		log.Println("Error putting item in DynamoDb", err)
		return err
	}

	return nil
}
