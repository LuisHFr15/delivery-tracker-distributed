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

func CreateOrderTable(ctx context.Context, client *dynamodb.Client, tableName string) error {
	_, err := client.CreateTable(ctx, &dynamodb.CreateTableInput{
		TableName: aws.String(tableName),
		AttributeDefinitions: []types.AttributeDefinition{
			{
				AttributeName: aws.String("OrderId"),
				AttributeType: types.ScalarAttributeTypeS,
			},
		},
		KeySchema: []types.KeySchemaElement{
			{
				AttributeName: aws.String("OrderId"),
				KeyType:       types.KeyTypeHash,
			},
		},
		BillingMode: types.BillingModePayPerRequest,
	})
	if err != nil {
		return err
	}
	return nil
}

func GetOrderRecord() data.OrderRecord {
	o := order.Order{
		EventId: uuid.New(),
		Id:      uuid.New(),
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
		Status:      "NEW",
		CreatedAt:   time.Now(),
	}
	return data.FromOrder(o)
}

func PutOrderRecord(ctx context.Context, client *dynamodb.Client, record data.OrderRecord, tableName string) error {
	value, err := attributevalue.MarshalMapWithOptions(record, func(o *attributevalue.EncoderOptions) {
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

func GetItem(ctx context.Context, client *dynamodb.Client, tableName string, ordId uuid.UUID) (*data.OrderRecord, error) {
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
	var item []data.OrderRecord
	err = attributevalue.UnmarshalListOfMapsWithOptions(output.Items, &item, func(o *attributevalue.DecoderOptions) {
		o.UseEncodingUnmarshalers = true
	})
	if err != nil {
		return nil, err
	}
	return new(item[0]), nil
}

func TestDynamoOrderRepository_Integration(t *testing.T) {
	ctx := context.Background()
	tableName := "orders"
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
	err = CreateOrderTable(ctx, client, tableName)
	if err != nil {
		t.Fatalf("could not create table: %v", err)
	}
	record := GetOrderRecord()
	fmt.Println("Inserting Item")
	err = PutOrderRecord(ctx, client, record, tableName)
	if err != nil {
		t.Fatalf("could not put item: %v", err)
	}
	fmt.Println("Getting Item")
	out, err := GetItem(ctx, client, tableName, record.OrderId)
	if err != nil {
		t.Fatalf("could not get item: %v", err)
	}
	if out == nil || out.OrderId != record.OrderId {
		t.Fatalf("did not get expected order")
	}
	fmt.Println("Completed Test for OrderRepository")
}
