package dynamo

import (
	"context"
	"log"
	"os"
	"time"

	"github.com/LuisHFr15/delivery-tracker-distributed/internal/processor/domain/models/data"
	"github.com/LuisHFr15/delivery-tracker-distributed/internal/processor/domain/models/order"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/feature/dynamodb/attributevalue"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
)

type DynamoOrderRepository struct {
	tableName string
	client    *dynamodb.Client
	buffer    chan *order.Order
	done      chan struct{}
}

func NewDynamoOrderRepository(ctx context.Context) *DynamoOrderRepository {
	cfg, err := config.LoadDefaultConfig(ctx)
	if err != nil {
		log.Fatalf("unable to load SDK config, %v", err)
	}
	return &DynamoOrderRepository{
		tableName: os.Getenv("DYNAMODB_ORDER_TABLE_NAME"),
		client:    dynamodb.NewFromConfig(cfg),
		buffer:    make(chan *order.Order),
		done:      make(chan struct{}),
	}
}

func (d *DynamoOrderRepository) Add(o order.Order) {
	d.buffer <- &o
}

func (d *DynamoOrderRepository) RunWorker() {
	defer close(d.done)
	for o := range d.buffer {
		record := data.FromOrder(*o)
		item, err := attributevalue.MarshalMapWithOptions(record, func(o *attributevalue.EncoderOptions) {
			o.UseEncodingMarshalers = true
		})

		if err != nil {
			log.Printf("Error serializing Order item: %s", err)
			continue
		}

		reqCtx, cancel := context.WithTimeout(context.Background(), time.Second*5)

		_, err = d.client.PutItem(reqCtx, &dynamodb.PutItemInput{
			TableName: aws.String(d.tableName),
			Item:      item,
		})

		cancel()
		if err != nil {
			log.Printf("Error putting Order in DynamoDb: %s", err)
		}
	}
}

func (d *DynamoOrderRepository) StopWorker() error {
	// close to avoid adding new inputs
	close(d.buffer)
	// wait until it empty the buffer
	<-d.done
	return nil
}

func (d *DynamoOrderRepository) GetByOrderId(ctx context.Context, orderId string) (*order.Order, error) {
	reqCtx, cancel := context.WithTimeout(ctx, time.Second*5)
	defer cancel()

	out, err := d.client.GetItem(reqCtx, &dynamodb.GetItemInput{
		TableName: aws.String(d.tableName),
		Key: map[string]types.AttributeValue{
			"OrderId": &types.AttributeValueMemberS{Value: orderId},
		},
	})
	if err != nil {
		return nil, err
	}

	if out.Item == nil {
		return nil, nil
	}

	var record data.OrderRecord
	if err := attributevalue.UnmarshalMapWithOptions(out.Item, &record, func(o *attributevalue.DecoderOptions) {
		o.UseEncodingUnmarshalers = true
	}); err != nil {
		return nil, err
	}

	domainOrder := record.ToDomain()
	return &domainOrder, nil
}
