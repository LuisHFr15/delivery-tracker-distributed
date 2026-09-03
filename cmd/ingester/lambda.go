package main

import (
	"context"

	"github.com/LuisHFr15/delivery-tracker-distributed/internal/ingester/app/services"
	"github.com/LuisHFr15/delivery-tracker-distributed/internal/ingester/infrastructure/http"
	"github.com/LuisHFr15/delivery-tracker-distributed/internal/ingester/infrastructure/queues"
	"github.com/aws/aws-lambda-go/lambda"
)

// runLambda is the serverless mode: a single-invocation function behind API
// Gateway that publishes each event to SNS. It reuses the same IngesterService
// as the HTTP server, only swapping the Kafka publisher for the SNS one.
func runLambda() {
	ctx := context.Background()
	service := services.NewIngesterService(queues.NewSnsPublisher(ctx))
	lambda.Start(http.NewAPIGatewayHandler(service))
}
