package queues

import (
	"context"
	"encoding/json"
	"net"
	"os"
	"strconv"
	"testing"
	"time"

	"github.com/LuisHFr15/delivery-tracker-distributed/internal/ingester/app/dtos"
	"github.com/google/uuid"
	"github.com/segmentio/kafka-go"
	tc_kafka "github.com/testcontainers/testcontainers-go/modules/kafka"
)

func createTopics(brokerAddress string, topics []string) error {
	conn, err := kafka.Dial("tcp", brokerAddress)
	if err != nil {
		return err
	}
	defer conn.Close()

	controller, err := conn.Controller()
	if err != nil {
		return err
	}
	controllerConn, err := kafka.Dial("tcp", net.JoinHostPort(controller.Host, strconv.Itoa(controller.Port)))
	if err != nil {
		return err
	}
	defer controllerConn.Close()

	topicConfigs := make([]kafka.TopicConfig, len(topics))
	for i, t := range topics {
		topicConfigs[i] = kafka.TopicConfig{
			Topic:             t,
			NumPartitions:     1,
			ReplicationFactor: 1,
		}
	}

	return controllerConn.CreateTopics(topicConfigs...)
}

func TestKafkaPublisher_Integration(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test in short mode")
	}

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Minute)
	defer cancel()

	t.Log("Starting Kafka container...")
	kafkaContainer, err := tc_kafka.Run(ctx, "confluentinc/cp-kafka:latest")
	if err != nil {
		t.Fatalf("failed to start kafka container: %v", err)
	}
	defer func() {
		t.Log("Terminating Kafka container...")
		if err := kafkaContainer.Terminate(ctx); err != nil {
			t.Errorf("failed to terminate container: %v", err)
		}
	}()

	brokers, err := kafkaContainer.Brokers(ctx)
	if err != nil {
		t.Fatalf("failed to get brokers: %v", err)
	}
	if len(brokers) == 0 {
		t.Fatal("no brokers returned from container")
	}
	brokerAddress := brokers[0]
	t.Logf("Kafka broker running at: %s", brokerAddress)

	t.Log("Creating Kafka topics programmatically...")
	err = createTopics(brokerAddress, []string{"order-events", "location-events"})
	if err != nil {
		t.Fatalf("failed to create topics: %v", err)
	}

	os.Setenv("KAFKA_BROKER", brokerAddress)
	defer os.Unsetenv("KAFKA_BROKER")

	publisher := NewKafkaPublisher()
	defer publisher.Close()

	orderID := uuid.New()
	orderEvent := dtos.OrderEventDTO{
		EventID: uuid.New(),
		Order: dtos.OrderDTO{
			ID:       orderID,
			ClientID: uuid.New(),
			Status:   "CREATED",
		},
		TransactionType: "CREATE",
		Timestamp:       time.Now().Truncate(time.Millisecond),
	}

	t.Log("Publishing Order Event...")
	publisher.PublishOrder(orderEvent)

	t.Log("Consuming published Order Event for assertion...")
	orderReader := kafka.NewReader(kafka.ReaderConfig{
		Brokers:     []string{brokerAddress},
		Topic:       "order-events",
		GroupID:     "integration-test-order-group",
		StartOffset: kafka.FirstOffset,
	})
	defer orderReader.Close()

	readCtx, readCancel := context.WithTimeout(ctx, 15*time.Second)
	defer readCancel()

	msg, err := orderReader.ReadMessage(readCtx)
	if err != nil {
		t.Fatalf("failed to read message from kafka: %v", err)
	}

	if string(msg.Key) != orderID.String() {
		t.Errorf("expected message key %s, got %s", orderID.String(), string(msg.Key))
	}

	var consumedOrder dtos.OrderEventDTO
	err = json.Unmarshal(msg.Value, &consumedOrder)
	if err != nil {
		t.Fatalf("failed to unmarshal consumed message: %v", err)
	}

	if consumedOrder.Order.ID != orderEvent.Order.ID {
		t.Errorf("expected order ID %s, got %s", orderEvent.Order.ID, consumedOrder.Order.ID)
	}
	if consumedOrder.TransactionType != orderEvent.TransactionType {
		t.Errorf("expected transaction type %s, got %s", orderEvent.TransactionType, consumedOrder.TransactionType)
	}

	locationEvent := dtos.LocationEventDTO{
		EventID:   uuid.New(),
		OrderID:   orderID,
		Latitude:  -23.5505,
		Longitude: -46.6333,
		Timestamp: time.Now().Truncate(time.Millisecond),
	}

	t.Log("Publishing Location Event...")
	publisher.PublishLocation(locationEvent)

	t.Log("Consuming published Location Event for assertion...")
	locationReader := kafka.NewReader(kafka.ReaderConfig{
		Brokers:     []string{brokerAddress},
		Topic:       "location-events",
		GroupID:     "integration-test-location-group",
		StartOffset: kafka.FirstOffset,
	})
	defer locationReader.Close()

	locReadCtx, locReadCancel := context.WithTimeout(ctx, 15*time.Second)
	defer locReadCancel()

	locMsg, err := locationReader.ReadMessage(locReadCtx)
	if err != nil {
		t.Fatalf("failed to read location message from kafka: %v", err)
	}

	if string(locMsg.Key) != orderID.String() {
		t.Errorf("expected location message key %s, got %s", orderID.String(), string(locMsg.Key))
	}

	var consumedLocation dtos.LocationEventDTO
	err = json.Unmarshal(locMsg.Value, &consumedLocation)
	if err != nil {
		t.Fatalf("failed to unmarshal consumed location: %v", err)
	}

	if consumedLocation.OrderID != locationEvent.OrderID {
		t.Errorf("expected location order ID %s, got %s", locationEvent.OrderID, consumedLocation.OrderID)
	}
	if consumedLocation.Latitude != locationEvent.Latitude || consumedLocation.Longitude != locationEvent.Longitude {
		t.Errorf("expected lat/lng %f/%f, got %f/%f", locationEvent.Latitude, locationEvent.Longitude, consumedLocation.Latitude, consumedLocation.Longitude)
	}
}
