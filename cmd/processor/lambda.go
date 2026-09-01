package main

import (
	"context"
	"encoding/json"
	"log"
	"os"

	"github.com/LuisHFr15/delivery-tracker-distributed/internal/processor/app/dtos"
	"github.com/LuisHFr15/delivery-tracker-distributed/internal/processor/app/services"
	"github.com/LuisHFr15/delivery-tracker-distributed/internal/processor/infrastructure/data/dynamo"
	"github.com/LuisHFr15/delivery-tracker-distributed/internal/processor/infrastructure/messaging"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
)

// runLambda is the serverless mode: SNS delivers order/location events to the
// function, which routes each record to the same app services the Kafka loop
// uses. The Dynamo repos and the SNS writer are built per invocation and
// flushed via StopWorker before returning, reusing the graceful-shutdown drain.
func runLambda() {
	lambda.Start(handleSNS)
}

func handleSNS(ctx context.Context, event events.SNSEvent) error {
	orderTopic := os.Getenv("SNS_ORDER_TOPIC_ARN")
	locationTopic := os.Getenv("SNS_LOCATION_TOPIC_ARN")

	auditingRepo := dynamo.NewDynamoAuditingEventRepository(ctx)
	processedRepo := dynamo.NewDynamoProcessedOrderRepository(ctx)
	orderRepo := dynamo.NewDynamoOrderRepository(ctx)
	notificationWriter := messaging.NewSnsNotificationWriter(ctx)

	go auditingRepo.RunWorker()
	go processedRepo.RunWorker()
	go orderRepo.RunWorker()
	go notificationWriter.RunWorker()

	for _, record := range event.Records {
		switch record.SNS.TopicArn {
		case orderTopic:
			var dto dtos.OrderEventDTO
			if err := json.Unmarshal([]byte(record.SNS.Message), &dto); err != nil {
				log.Printf("bad order event: %v", err)
				continue
			}
			svc := services.NewOrderEventService(dto, auditingRepo, orderRepo)
			if err := svc.ConvertEvent(); err != nil {
				log.Printf("failed to convert order event: %v", err)
			}
		case locationTopic:
			var dto dtos.LocationEventDTO
			if err := json.Unmarshal([]byte(record.SNS.Message), &dto); err != nil {
				log.Printf("bad location event: %v", err)
				continue
			}
			svc := services.NewLocationEventService(dto, auditingRepo, processedRepo, orderRepo, notificationWriter)
			if err := svc.ProcessEvent(); err != nil {
				log.Printf("failed to process location event: %v", err)
			}
		default:
			log.Printf("unexpected topic ARN: %s", record.SNS.TopicArn)
		}
	}

	// Flush the buffered writers before the invocation returns.
	for name, stop := range map[string]func() error{
		"Processed Order Repo": processedRepo.StopWorker,
		"Order Repo":           orderRepo.StopWorker,
		"Auditing Repo":        auditingRepo.StopWorker,
		"SNS Writer":           notificationWriter.StopWorker,
	} {
		if err := stop(); err != nil {
			log.Printf("failed to flush %s: %v", name, err)
		}
	}
	return nil
}
