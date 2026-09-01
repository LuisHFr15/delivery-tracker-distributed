package main

import (
	"context"
	"encoding/json"
	"log"
	"net/http"
	"strings"

	"github.com/LuisHFr15/delivery-tracker-distributed/internal/ingester/app/dtos"
	"github.com/LuisHFr15/delivery-tracker-distributed/internal/ingester/app/services"
	"github.com/LuisHFr15/delivery-tracker-distributed/internal/ingester/infrastructure/queues"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
)

// runLambda is the serverless mode: a single-invocation function behind API
// Gateway that publishes each event to SNS. It reuses the same IngesterService
// as the HTTP server, only swapping the Kafka publisher for the SNS one.
func runLambda() {
	ctx := context.Background()
	service := services.NewIngesterService(queues.NewSnsPublisher(ctx))
	lambda.Start(newAPIGatewayHandler(service))
}

func newAPIGatewayHandler(service *services.IngesterService) func(context.Context, events.APIGatewayV2HTTPRequest) (events.APIGatewayV2HTTPResponse, error) {
	return func(ctx context.Context, req events.APIGatewayV2HTTPRequest) (events.APIGatewayV2HTTPResponse, error) {
		method := req.RequestContext.HTTP.Method
		path := req.RequestContext.HTTP.Path

		switch {
		case method == http.MethodGet && path == "/api/health":
			return respond(http.StatusOK, `{"status":"ok"}`), nil

		case method == http.MethodPost && path == "/api/order":
			var dto dtos.OrderEventDTO
			if err := json.Unmarshal([]byte(req.Body), &dto); err != nil {
				return respond(http.StatusBadRequest, `{"error":"invalid request format"}`), nil
			}
			if err := service.IngestOrder(ctx, dto); err != nil {
				return respond(http.StatusInternalServerError, errorBody(err)), nil
			}
			return respond(http.StatusAccepted, ""), nil

		case method == http.MethodPost && strings.HasPrefix(path, "/api/order/") && strings.HasSuffix(path, "/location"):
			var dto dtos.LocationEventDTO
			if err := json.Unmarshal([]byte(req.Body), &dto); err != nil {
				return respond(http.StatusBadRequest, `{"error":"invalid request format"}`), nil
			}
			if err := service.IngestLocation(ctx, dto); err != nil {
				return respond(http.StatusInternalServerError, errorBody(err)), nil
			}
			return respond(http.StatusAccepted, ""), nil

		default:
			return respond(http.StatusNotFound, `{"error":"route not found"}`), nil
		}
	}
}

func respond(status int, body string) events.APIGatewayV2HTTPResponse {
	return events.APIGatewayV2HTTPResponse{
		StatusCode: status,
		Headers:    map[string]string{"Content-Type": "application/json"},
		Body:       body,
	}
}

func errorBody(err error) string {
	value, marshalErr := json.Marshal(map[string]string{"error": err.Error()})
	if marshalErr != nil {
		log.Printf("failed to marshal error response: %v", marshalErr)
		return `{"error":"internal error"}`
	}
	return string(value)
}
