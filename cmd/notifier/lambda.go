package main

import (
	"context"
	"encoding/json"
	"log"

	notifierservices "github.com/LuisHFr15/delivery-tracker-distributed/internal/notifier/app/services"
	"github.com/LuisHFr15/delivery-tracker-distributed/internal/notifier/infrastructure/data/dynamo"
	"github.com/LuisHFr15/delivery-tracker-distributed/internal/processor/app/dtos"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
)

// runLambda is the serverless mode: SNS delivers processed-orders and the
// function persists one notification per record, flushing the repo before it
// returns.
func runLambda() {
	lambda.Start(handleSNS)
}

func handleSNS(ctx context.Context, event events.SNSEvent) error {
	repo := dynamo.NewDynamoNotifiedMessageRepository(ctx)
	go repo.RunWorker()

	for _, record := range event.Records {
		var dto dtos.ProcessedOrderDTO
		if err := json.Unmarshal([]byte(record.SNS.Message), &dto); err != nil {
			log.Printf("bad processed order message: %v", err)
			continue
		}
		if err := notifierservices.NewNotifierService(dto, repo).Notify(); err != nil {
			log.Printf("failed to notify: %v", err)
		}
	}

	if err := repo.StopWorker(); err != nil {
		log.Printf("failed to flush notified-message repo: %v", err)
	}
	return nil
}
