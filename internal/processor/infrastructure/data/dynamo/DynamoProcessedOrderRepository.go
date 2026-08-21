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
	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
)

type DynamoProcessedOrderRepository struct {
	tableName string
	client    *dynamodb.Client
	buffer    chan *data.ProcessedOrder
	done      chan struct{}
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
		done:      make(chan struct{}),
	}
}

func (d *DynamoProcessedOrderRepository) Add(cls data.ProcessedOrder) {
	d.buffer <- &cls
}

func (d *DynamoProcessedOrderRepository) RunWorker() {
	defer close(d.done)
	for cls := range d.buffer {
		item, err := attributevalue.MarshalMapWithOptions(cls, func(o *attributevalue.EncoderOptions) {
			o.UseEncodingMarshalers = true
		})

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

func (d *DynamoProcessedOrderRepository) StopWorker() error {
	close(d.buffer)
	<-d.done
	return nil
}

func (d *DynamoProcessedOrderRepository) GetLatestByOrderId(ctx context.Context, orderId string) (*data.ProcessedOrder, error) {
	paginator := dynamodb.NewQueryPaginator(d.client, &dynamodb.QueryInput{
		TableName:              aws.String(d.tableName),
		KeyConditionExpression: aws.String("OrderId = :oid"),
		ExpressionAttributeValues: map[string]types.AttributeValue{
			":oid": &types.AttributeValueMemberS{Value: orderId},
		},
	})

	var latest *data.ProcessedOrder
	for paginator.HasMorePages() {
		reqCtx, cancel := context.WithTimeout(ctx, time.Second*5)
		out, err := paginator.NextPage(reqCtx)
		cancel()
		if err != nil {
			return nil, err
		}

		var page []data.ProcessedOrder
		if err := attributevalue.UnmarshalListOfMapsWithOptions(out.Items, &page, func(o *attributevalue.DecoderOptions) {
			o.UseEncodingUnmarshalers = true
		}); err != nil {
			return nil, err
		}

		for i := range page {
			if latest == nil || page[i].Timestamp.After(latest.Timestamp) {
				latest = &page[i]
			}
		}
	}

	return latest, nil
}
