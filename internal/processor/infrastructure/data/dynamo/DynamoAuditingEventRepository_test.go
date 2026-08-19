package dynamo

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/LuisHFr15/delivery-tracker-distributed/internal/processor/app/dtos"
	"github.com/LuisHFr15/delivery-tracker-distributed/internal/processor/domain/models/data"
	"github.com/LuisHFr15/delivery-tracker-distributed/internal/processor/domain/models/order"
	"github.com/LuisHFr15/delivery-tracker-distributed/internal/processor/domain/services"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	credentials2 "github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/feature/dynamodb/attributevalue"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
	"github.com/google/uuid"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/wait"
)

func GenerateClient(ctx context.Context, container testcontainers.Container) (*dynamodb.Client, error) {
	host, err := container.Host(ctx)
	if err != nil {
		return nil, err
	}

	port, err := container.MappedPort(ctx, "8000")
	if err != nil {
		return nil, err
	}

	endpoint := fmt.Sprintf("http://%s:%s/", host, port.Port())
	cfg, err := config.LoadDefaultConfig(ctx,
		config.WithDefaultRegion("us-east-1"),
		config.WithCredentialsProvider(credentials2.NewStaticCredentialsProvider("dummy", "dummy", "")),
	)
	if err != nil {
		return nil, err
	}

	client := dynamodb.NewFromConfig(cfg, func(o *dynamodb.Options) {
		o.BaseEndpoint = aws.String(endpoint)
	})
	return client, nil
}

func CreateTable(ctx context.Context, client *dynamodb.Client, tableName string) error {
	_, err := client.CreateTable(ctx, &dynamodb.CreateTableInput{
		TableName: aws.String(tableName),
		AttributeDefinitions: []types.AttributeDefinition{
			{
				AttributeName: aws.String("EventType"),
				AttributeType: types.ScalarAttributeTypeS,
			},
			{
				AttributeName: aws.String("Id"),
				AttributeType: types.ScalarAttributeTypeS,
			},
		},
		KeySchema: []types.KeySchemaElement{
			{
				AttributeName: aws.String("EventType"),
				KeyType:       types.KeyTypeHash,
			},
			{
				AttributeName: aws.String("Id"),
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

func GetEvents() []data.DynamoEvent {
	ordConverter := services.NewOrderEventConverter()
	locConverter := services.NewLocationEventConverter()
	ordEventDto := dtos.OrderEventDTO{
		EventID: uuid.New(),
		Order: dtos.OrderDTO{
			ID:       uuid.New(),
			ClientID: uuid.New(),
			Products: []dtos.OrderItemDTO{
				{
					ProductID: uuid.New(),
					Name:      "test",
					Price:     1.0,
					Quantity:  10,
				},
			},
			DeliveryID: nil,
			Destination: &order.Location{
				Lat: 0,
				Lng: 0,
			},
			Status: "NEW",
		},
		TransactionType: "CREATE",
		Timestamp:       time.Now(),
	}

	locEventDto := dtos.LocationEventDTO{
		EventID:   ordEventDto.EventID,
		OrderID:   ordEventDto.Order.ID,
		Latitude:  1,
		Longitude: 2,
		Timestamp: time.Now(),
	}

	return []data.DynamoEvent{
		ordConverter.Convert(ordEventDto),
		locConverter.Convert(locEventDto, "DELIVERING"),
	}
}

func PutItems(ctx context.Context, client *dynamodb.Client, items []data.DynamoEvent, tableName string) error {
	for _, item := range items {
		value, err := attributevalue.MarshalMapWithOptions(item, func(o *attributevalue.EncoderOptions) {
			o.UseEncodingMarshalers = true
		})
		if err != nil {
			return err
		}
		_, err = client.PutItem(ctx, &dynamodb.PutItemInput{
			TableName: aws.String(tableName),
			Item:      value,
		})
		if err != nil {
			return err
		}
	}
	return nil

}

func TestDynamoAuditingEventRepository_Integration(t *testing.T) {
	ctx := context.Background()
	tableName := "auditing"
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
	err = CreateTable(ctx, client, tableName)
	if err != nil {
		t.Fatalf("could not create table: %v", err)
	}
	items := GetEvents()
	fmt.Println("Inserting Items")
	err = PutItems(ctx, client, items, tableName)
	if err != nil {
		t.Fatalf("could not put items: %v", err)
	}
	fmt.Println("Completed Test for AuditingRepository")
}
