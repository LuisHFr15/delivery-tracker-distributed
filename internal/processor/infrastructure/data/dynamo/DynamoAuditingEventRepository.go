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

type DynamoAuditingEventRepository struct {
	tableName string
	client    *dynamodb.Client
}

func NewDynamoAuditingEventRepository(tableName string, client *dynamodb.Client) *DynamoAuditingEventRepository {
	return &DynamoAuditingEventRepository{
		tableName: tableName,
		client:    client,
	}
}

func (d *DynamoAuditingEventRepository) Add(ctx context.Context, cls data.DynamoEvent) error {
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
