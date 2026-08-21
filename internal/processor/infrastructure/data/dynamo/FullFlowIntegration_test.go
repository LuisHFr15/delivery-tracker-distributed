//go:build integration

package dynamo

import (
	"context"
	"encoding/json"
	"net"
	"strconv"
	"testing"
	"time"

	"github.com/LuisHFr15/delivery-tracker-distributed/internal/processor/app/dtos"
	appservices "github.com/LuisHFr15/delivery-tracker-distributed/internal/processor/app/services"
	"github.com/LuisHFr15/delivery-tracker-distributed/internal/processor/domain/models/data"
	"github.com/LuisHFr15/delivery-tracker-distributed/internal/processor/domain/models/order"
	"github.com/LuisHFr15/delivery-tracker-distributed/internal/processor/infrastructure/messaging"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
	"github.com/google/uuid"
	kafkago "github.com/segmentio/kafka-go"
	"github.com/testcontainers/testcontainers-go"
	tckafka "github.com/testcontainers/testcontainers-go/modules/kafka"
	"github.com/testcontainers/testcontainers-go/wait"
)

// -----------------------------------------------------------------------------
// Full-path integration: Kafka (input topic) -> service -> DynamoDB + Kafka
// (output topic). Real KafkaReadOrderTopic/KafkaReadLocationTopic/
// KafkaWriteNotification and real Dynamo repositories are wired against
// testcontainers. Repos are built via struct literals so we can inject the
// dynamodb-local client without touching production constructors.
// -----------------------------------------------------------------------------

// startDynamo boots a dynamodb-local container and returns its client.
func startDynamo(t *testing.T, ctx context.Context) *dynamodb.Client {
	t.Helper()
	container, err := testcontainers.GenericContainer(ctx, testcontainers.GenericContainerRequest{
		ContainerRequest: testcontainers.ContainerRequest{
			Image:        "amazon/dynamodb-local:latest",
			ExposedPorts: []string{"8000/tcp"},
			WaitingFor:   wait.ForListeningPort("8000/tcp"),
		},
		Started: true,
	})
	if err != nil {
		t.Fatalf("could not start dynamodb-local: %v", err)
	}
	t.Cleanup(func() { _ = container.Terminate(ctx) })

	client, err := GenerateClient(ctx, container)
	if err != nil {
		t.Fatalf("could not create dynamo client: %v", err)
	}
	return client
}

// startKafka boots a Kafka container and returns the first broker address.
func startKafka(t *testing.T, ctx context.Context) string {
	t.Helper()
	kc, err := tckafka.Run(ctx, "confluentinc/confluent-local:7.5.0", tckafka.WithClusterID("test-cluster"))
	if err != nil {
		t.Fatalf("could not start kafka: %v", err)
	}
	t.Cleanup(func() { _ = kc.Terminate(ctx) })

	brokers, err := kc.Brokers(ctx)
	if err != nil || len(brokers) == 0 {
		t.Fatalf("could not get kafka brokers: %v", err)
	}
	return brokers[0]
}

// local repository builders — struct literals with the injected local client.
func newLocalOrderRepo(client *dynamodb.Client, table string) *DynamoOrderRepository {
	return &DynamoOrderRepository{tableName: table, client: client, buffer: make(chan *order.Order), done: make(chan struct{})}
}
func newLocalAuditRepo(client *dynamodb.Client, table string) *DynamoAuditingEventRepository {
	return &DynamoAuditingEventRepository{tableName: table, client: client, buffer: make(chan *data.DynamoEvent), done: make(chan struct{})}
}
func newLocalProcRepo(client *dynamodb.Client, table string) *DynamoProcessedOrderRepository {
	return &DynamoProcessedOrderRepository{tableName: table, client: client, buffer: make(chan *data.ProcessedOrder), done: make(chan struct{})}
}

// produceJSON writes a JSON message to a topic, retrying while the topic is
// being auto-created by the broker.
func produceJSON(t *testing.T, ctx context.Context, broker, topic string, key []byte, payload any) {
	t.Helper()
	value, err := json.Marshal(payload)
	if err != nil {
		t.Fatalf("marshal payload: %v", err)
	}
	w := &kafkago.Writer{
		Addr:                   kafkago.TCP(broker),
		Topic:                  topic,
		Balancer:               &kafkago.LeastBytes{},
		AllowAutoTopicCreation: true,
	}
	defer w.Close()

	deadline := time.Now().Add(30 * time.Second)
	for {
		wctx, cancel := context.WithTimeout(ctx, 5*time.Second)
		err = w.WriteMessages(wctx, kafkago.Message{Key: key, Value: value})
		cancel()
		if err == nil {
			return
		}
		if time.Now().After(deadline) {
			t.Fatalf("could not produce to %s: %v", topic, err)
		}
		time.Sleep(time.Second)
	}
}

// createTopic pre-creates a topic with a single partition. The production
// KafkaWriteNotification worker does not retry failed writes, so the output
// topic must exist before the notification is emitted (auto-creation races
// against that first write).
func createTopic(t *testing.T, broker, topic string) {
	t.Helper()
	conn, err := kafkago.Dial("tcp", broker)
	if err != nil {
		t.Fatalf("dial broker: %v", err)
	}
	defer conn.Close()

	controller, err := conn.Controller()
	if err != nil {
		t.Fatalf("get controller: %v", err)
	}
	ctrlConn, err := kafkago.Dial("tcp", net.JoinHostPort(controller.Host, strconv.Itoa(controller.Port)))
	if err != nil {
		t.Fatalf("dial controller: %v", err)
	}
	defer ctrlConn.Close()

	if err := ctrlConn.CreateTopics(kafkago.TopicConfig{Topic: topic, NumPartitions: 1, ReplicationFactor: 1}); err != nil {
		t.Fatalf("create topic %s: %v", topic, err)
	}
}

// consumeOne reads a single message from the beginning of a topic.
func consumeOne(t *testing.T, ctx context.Context, broker, topic string) []byte {
	t.Helper()
	r := kafkago.NewReader(kafkago.ReaderConfig{Brokers: []string{broker}, Topic: topic})
	defer r.Close()

	rctx, cancel := context.WithTimeout(ctx, 30*time.Second)
	defer cancel()
	msg, err := r.ReadMessage(rctx)
	if err != nil {
		t.Fatalf("could not consume from %s: %v", topic, err)
	}
	return msg.Value
}

// countAuditByType queries the auditing table by its EventType hash key.
func countAuditByType(ctx context.Context, client *dynamodb.Client, table, eventType string) (int, error) {
	out, err := client.Query(ctx, &dynamodb.QueryInput{
		TableName:              aws.String(table),
		KeyConditionExpression: aws.String("EventType = :et"),
		ExpressionAttributeValues: map[string]types.AttributeValue{
			":et": &types.AttributeValueMemberS{Value: eventType},
		},
	})
	if err != nil {
		return 0, err
	}
	return int(out.Count), nil
}

func TestOrderEventService_FullFlow_Integration(t *testing.T) {
	ctx := context.Background()
	client := startDynamo(t, ctx)
	broker := startKafka(t, ctx)
	t.Setenv("KAFKA_BROKER", broker)

	if err := CreateOrderTable(ctx, client, "orders"); err != nil {
		t.Fatalf("create orders table: %v", err)
	}
	if err := CreateTable(ctx, client, "auditing"); err != nil {
		t.Fatalf("create auditing table: %v", err)
	}

	orderRepo := newLocalOrderRepo(client, "orders")
	auditRepo := newLocalAuditRepo(client, "auditing")
	go orderRepo.RunWorker()
	go auditRepo.RunWorker()

	// Produce an order event on the topic the real reader consumes.
	orderID := uuid.New()
	clientID := uuid.New()
	event := dtos.OrderEventDTO{
		EventID: uuid.New(),
		Order: dtos.OrderDTO{
			ID:          orderID,
			ClientID:    clientID,
			Products:    []dtos.OrderItemDTO{{ProductID: uuid.New(), Name: "widget", Price: 100, Quantity: 3}},
			Destination: &order.Location{Lat: 10, Lng: 20},
			Status:      "NEW",
		},
		TransactionType: "CREATED",
		Timestamp:       time.Now().UTC(),
	}
	produceJSON(t, ctx, broker, "order-events", []byte(orderID.String()), event)

	// Read it through the production Kafka reader and run the service.
	reader := messaging.NewKafkaReadOrderTopic()
	defer reader.Close()
	rctx, cancel := context.WithTimeout(ctx, 30*time.Second)
	defer cancel()
	gotDTO, err := reader.Read(rctx)
	if err != nil {
		t.Fatalf("reader.Read: %v", err)
	}

	svc := appservices.NewOrderEventService(gotDTO, auditRepo, orderRepo)
	if err := svc.ConvertEvent(); err != nil {
		t.Fatalf("ConvertEvent: %v", err)
	}

	// Flush the async workers so the writes land in DynamoDB.
	if err := orderRepo.StopWorker(); err != nil {
		t.Fatalf("stop order worker: %v", err)
	}
	if err := auditRepo.StopWorker(); err != nil {
		t.Fatalf("stop audit worker: %v", err)
	}

	// Assert the persisted order.
	rec, err := GetItem(ctx, client, "orders", orderID)
	if err != nil {
		t.Fatalf("get order: %v", err)
	}
	if rec == nil || rec.OrderId != orderID {
		t.Fatalf("persisted order = %+v, want OrderId %v", rec, orderID)
	}
	if rec.ClientId != clientID {
		t.Errorf("persisted ClientId = %v, want %v", rec.ClientId, clientID)
	}
	if rec.Status != "NEW" {
		t.Errorf("persisted Status = %q, want NEW", rec.Status)
	}
	if len(rec.Products) != 1 || rec.Products[0].Quantity != 3 {
		t.Errorf("persisted Products = %+v, want one item qty 3", rec.Products)
	}

	// Assert the auditing event was written.
	n, err := countAuditByType(ctx, client, "auditing", "order_event")
	if err != nil {
		t.Fatalf("query auditing: %v", err)
	}
	if n < 1 {
		t.Errorf("auditing rows for order_event = %d, want >= 1", n)
	}
}

func TestLocationEventService_FullFlow_Integration(t *testing.T) {
	ctx := context.Background()
	client := startDynamo(t, ctx)
	broker := startKafka(t, ctx)
	t.Setenv("KAFKA_BROKER", broker)

	if err := CreateOrderTable(ctx, client, "orders"); err != nil {
		t.Fatalf("create orders table: %v", err)
	}
	if err := CreateProcessedOrderTable(ctx, client, "processed_orders"); err != nil {
		t.Fatalf("create processed_orders table: %v", err)
	}
	if err := CreateTable(ctx, client, "auditing"); err != nil {
		t.Fatalf("create auditing table: %v", err)
	}

	// Seed an existing order so GetByOrderId finds it.
	orderID := uuid.New()
	clientID := uuid.New()
	dest := order.Location{Lat: 10, Lng: 20}
	seeded := order.Order{
		Id:          orderID,
		EventId:     uuid.New(),
		Client:      order.Client{Id: clientID},
		Products:    []order.OrderItem{{Product: order.Product{Id: uuid.New(), Name: "widget", ProductPrice: 100}, Quantity: 2}},
		Destination: &dest,
		Status:      "NEW",
		CreatedAt:   time.Now().UTC(),
	}
	if err := PutOrderRecord(ctx, client, data.FromOrder(seeded), "orders"); err != nil {
		t.Fatalf("seed order: %v", err)
	}

	orderRepo := newLocalOrderRepo(client, "orders")
	auditRepo := newLocalAuditRepo(client, "auditing")
	procRepo := newLocalProcRepo(client, "processed_orders")
	go orderRepo.RunWorker()
	go auditRepo.RunWorker()
	go procRepo.RunWorker()

	// The notification writer does not retry, so ensure the output topic exists.
	createTopic(t, broker, "processed-orders")

	writer := messaging.NewKafkaWriteNotification()
	go writer.RunWorker()

	// Produce a location event that is NOT at the destination -> DELIVERING.
	locEvent := dtos.LocationEventDTO{
		EventID:   uuid.New(),
		OrderID:   orderID,
		Latitude:  10.5,
		Longitude: 20.0,
		Timestamp: time.Now().UTC(),
	}
	produceJSON(t, ctx, broker, "location-events", []byte(orderID.String()), locEvent)

	reader := messaging.NewKafkaReadLocationTopic()
	defer reader.Close()
	rctx, cancel := context.WithTimeout(ctx, 30*time.Second)
	defer cancel()
	gotDTO, err := reader.Read(rctx)
	if err != nil {
		t.Fatalf("reader.Read: %v", err)
	}

	svc := appservices.NewLocationEventService(gotDTO, auditRepo, procRepo, orderRepo, writer)
	if err := svc.ProcessEvent(); err != nil {
		t.Fatalf("ProcessEvent: %v", err)
	}

	// Flush all workers (including the Kafka notification writer).
	if err := orderRepo.StopWorker(); err != nil {
		t.Fatalf("stop order worker: %v", err)
	}
	if err := procRepo.StopWorker(); err != nil {
		t.Fatalf("stop proc worker: %v", err)
	}
	if err := auditRepo.StopWorker(); err != nil {
		t.Fatalf("stop audit worker: %v", err)
	}
	if err := writer.StopWorker(); err != nil {
		t.Fatalf("stop notification writer: %v", err)
	}

	// Assert the processed order in DynamoDB.
	po, err := GetProcessedItem(ctx, client, "processed_orders", orderID)
	if err != nil {
		t.Fatalf("get processed order: %v", err)
	}
	if po == nil || po.OrderId != orderID {
		t.Fatalf("processed order = %+v, want OrderId %v", po, orderID)
	}
	if po.OrderStatus != "DELIVERING" {
		t.Errorf("processed OrderStatus = %q, want DELIVERING", po.OrderStatus)
	}
	if po.TimestampLocation != (order.Location{Lat: 10.5, Lng: 20.0}) {
		t.Errorf("processed TimestampLocation = %v, want actual location", po.TimestampLocation)
	}
	if po.FinalLocation != dest {
		t.Errorf("processed FinalLocation = %v, want %v", po.FinalLocation, dest)
	}

	// Assert the notification emitted to the output topic.
	raw := consumeOne(t, ctx, broker, "processed-orders")
	var notif dtos.ProcessedOrderDTO
	if err := json.Unmarshal(raw, &notif); err != nil {
		t.Fatalf("unmarshal notification: %v", err)
	}
	if notif.OrderId != orderID {
		t.Errorf("notification OrderId = %v, want %v", notif.OrderId, orderID)
	}
	if notif.ClientId != clientID {
		t.Errorf("notification ClientId = %v, want %v", notif.ClientId, clientID)
	}
	if notif.TimeToDelivery <= 0 {
		t.Errorf("notification TimeToDelivery = %v, want > 0 (order in transit)", notif.TimeToDelivery)
	}
	if notif.ActualLocation != po.TimestampLocation || notif.FinalDestination != po.FinalLocation {
		t.Errorf("notification locations = %+v, want to match processed order %+v", notif, po)
	}
}
