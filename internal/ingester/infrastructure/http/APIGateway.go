package http

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strings"

	"github.com/LuisHFr15/delivery-tracker-distributed/internal/ingester/app/dtos"
	"github.com/LuisHFr15/delivery-tracker-distributed/internal/ingester/app/services"
	"github.com/aws/aws-lambda-go/events"
)

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

func NewAPIGatewayHandler(service *services.IngesterService) func(context.Context, events.APIGatewayV2HTTPRequest) (events.APIGatewayV2HTTPResponse, error) {
	return func(ctx context.Context, req events.APIGatewayV2HTTPRequest) (events.APIGatewayV2HTTPResponse, error) {
		method := req.RequestContext.HTTP.Method
		path := req.RequestContext.HTTP.Path
		log.Println(method, path)

		switch {
		case method == http.MethodGet && path == "/health":
			return respond(http.StatusOK, `{"status":"ok"}`), nil

		case method == http.MethodPost && path == "/order":
			log.Println(req.Body)
			var dto dtos.OrderEventDTO
			if err := json.Unmarshal([]byte(req.Body), &dto); err != nil {
				return respond(http.StatusBadRequest, errorBody(fmt.Errorf(`{"error":"invalid request format", "description": %w}`, err))), nil
			}
			if err := service.IngestOrder(ctx, dto); err != nil {
				return respond(http.StatusInternalServerError, errorBody(err)), nil
			}
			return respond(http.StatusAccepted, ""), nil

		case method == http.MethodPost && strings.HasPrefix(path, "/order/") && strings.HasSuffix(path, "/location"):
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
