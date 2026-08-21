//go:build integration

package dynamo

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/LuisHFr15/delivery-tracker-distributed/internal/processor/domain/models/data"
	"github.com/LuisHFr15/delivery-tracker-distributed/internal/processor/domain/models/order"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/feature/dynamodb/attributevalue"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
	"github.com/google/uuid"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/wait"
)

func CreateProcessedOrderTable(ctx context.Context, client *dynamodb.Client, tableName string) error {
	_, err := client.CreateTable(ctx, &dynamodb.CreateTableInput{
		TableName: aws.String(tableName),
		AttributeDefinitions: []types.AttributeDefinition{
			{
				AttributeName: aws.String("OrderId"),
				AttributeType: types.ScalarAttributeTypeS,
			},
			{
				AttributeName: aws.String("Timestamp"),
				AttributeType: types.ScalarAttributeTypeS,
			},
		},
		KeySchema: []types.KeySchemaElement{
			{
				AttributeName: aws.String("OrderId"),
				KeyType:       types.KeyTypeHash,
			},
			{
				AttributeName: aws.String("Timestamp"),
				KeyType:       types.KeyTypeRange,
			},
		},
		BillingMode: types.BillingModePayPerRequest,
	})
	if err != nil {
		return err
	}
	return nil
}

func GetProcessedOrder() data.ProcessedOrder {
	o := order.Order{
		Id: uuid.New(),
		Products: []order.OrderItem{
			{
				Product: order.Product{
					Id:           uuid.New(),
					Name:         "test",
					ProductPrice: 100,
				},
				Quantity: 10,
			},
		},
		Client:      order.Client{Id: uuid.New()},
		Destination: &order.Location{Lat: 1, Lng: 2},
		Status:      "DELIVERING",
	}
	factory := data.NewProcessedOrderFactory()
	return factory.CreateProcessedOrder(o, uuid.New(), order.Location{Lat: 0, Lng: 0}, time.Now())
}

func PutProcessedOrder(ctx context.Context, client *dynamodb.Client, po data.ProcessedOrder, tableName string) error {
	value, err := attributevalue.MarshalMapWithOptions(po, func(o *attributevalue.EncoderOptions) {
		o.UseEncodingMarshalers = true
	})
	if err != nil {
		return err
	}
	_, err = client.PutItem(ctx, &dynamodb.PutItemInput{
		TableName: aws.String(tableName),
		Item:      value,
	})
	return err
}

func GetProcessedItem(ctx context.Context, client *dynamodb.Client, tableName string, ordId uuid.UUID) (*data.ProcessedOrder, error) {
	output, err := client.Query(ctx, &dynamodb.QueryInput{
		TableName:              aws.String(tableName),
		KeyConditionExpression: aws.String("OrderId = :oid"),
		ExpressionAttributeValues: map[string]types.AttributeValue{
			":oid": &types.AttributeValueMemberS{
				Value: ordId.String(),
			},
		},
	})
	if err != nil {
		return nil, err
	}
	var item []data.ProcessedOrder
	err = attributevalue.UnmarshalListOfMapsWithOptions(output.Items, &item, func(o *attributevalue.DecoderOptions) {
		o.UseEncodingUnmarshalers = true
	})
	if err != nil {
		return nil, err
	}
	return new(item[0]), nil
}

func TestDynamoProcessedOrderRepository_Integration(t *testing.T) {
	ctx := context.Background()
	tableName := "processed_orders"
	fmt.Println("Creating Container")
	container, err := testcontainers.GenericContainer(ctx, testcontainers.GenericContainerRequest{
		ContainerRequest: testcontainers.ContainerRequest{
			Image:        "amazon/dynamodb-local:latest",
			ExposedPorts: []string{"8000/tcp"},
			WaitingFor:   wait.ForListeningPort("8000/tcp"),
		},
		Started: true,
	})
	if err != nil {
		t.Fatalf("could not start local DynamoDB container: %v", err)
	}
	defer container.Terminate(ctx)
	fmt.Println("Starting Dynamo Client...")
	client, err := GenerateClient(ctx, container)
	if err != nil {
		t.Fatalf("could not create DynamoDB client: %v", err)
	}
	fmt.Println("Creating Dynamo Table:", tableName)
	err = CreateProcessedOrderTable(ctx, client, tableName)
	if err != nil {
		t.Fatalf("could not create table: %v", err)
	}
	po := GetProcessedOrder()
	fmt.Println("Inserting Item")
	err = PutProcessedOrder(ctx, client, po, tableName)
	if err != nil {
		t.Fatalf("could not put item: %v", err)
	}
	fmt.Println("Getting Item")
	out, err := GetProcessedItem(ctx, client, tableName, po.OrderId)
	if err != nil {
		t.Fatalf("could not get item: %v", err)
	}
	if out == nil || out.OrderId != po.OrderId {
		t.Fatalf("did not get expected processed order")
	}
	fmt.Println("Completed Test for ProcessedOrderRepository")
}
