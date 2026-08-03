package dynamo

import (
	"context"
	"log"
	"os"

	"github.com/LuisHFr15/delivery-tracker-distributed/internal/processor/domain/models"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/feature/dynamodb/attributevalue"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
)

type DynamoLocationRepository struct {
	client *dynamodb.Client
}

func (d *DynamoLocationRepository) Add(ctx context.Context, loc models.Location) error {
	item, err := attributevalue.MarshalMap(loc)
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
