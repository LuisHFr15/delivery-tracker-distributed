package http

import (
	"context"
	"encoding/json"
	"log"
	"net/http"
	"strings"

	"github.com/LuisHFr15/delivery-tracker-distributed/internal/ingester/app/dtos"
	"github.com/LuisHFr15/delivery-tracker-distributed/internal/ingester/app/services"
	"github.com/aws/aws-lambda-go/events"
	"github.com/google/uuid"
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

		switch {
		case method == http.MethodGet && path == "/health":
			return respond(http.StatusOK, `{"status":"ok"}`), nil

		case method == http.MethodPost && path == "/order":
			var dto dtos.OrderEventDTO
			if err := json.Unmarshal([]byte(req.Body), &dto); err != nil {
				return respond(http.StatusBadRequest, errorBody(err)), nil
			}
			if dto.EventID == uuid.Nil {
				dto.EventID = uuid.New()
			}
			if err := service.IngestOrder(ctx, dto); err != nil {
				return respond(http.StatusInternalServerError, errorBody(err)), nil
			}
			return respond(http.StatusAccepted, `{"message": "order being processed"}`), nil

		case method == http.MethodPost && strings.HasPrefix(path, "/order/") && strings.HasSuffix(path, "/location"):
			rawID := strings.TrimPrefix(path, "/order/")
			rawID = strings.TrimSuffix(rawID, "/location")
			pathOrderID, err := uuid.Parse(rawID)

			if err != nil {
				return respond(http.StatusBadRequest, `{"error":"invalid order_id in path"}`), nil
			}

			var dto dtos.LocationEventDTO
			if err := json.Unmarshal([]byte(req.Body), &dto); err != nil {
				return respond(http.StatusBadRequest, errorBody(err)), nil
			}
			if dto.EventID == uuid.Nil {
				dto.EventID = uuid.New()
			}
			if dto.OrderID != pathOrderID {
				return respond(http.StatusBadRequest, `{"error":"invalid order_id in path"}`), nil
			}
			if err := service.IngestLocation(ctx, dto); err != nil {
				return respond(http.StatusInternalServerError, errorBody(err)), nil
			}
			return respond(http.StatusAccepted, `{"message": "location update being processed"}`), nil

		default:
			return respond(http.StatusNotFound, `{"error":"route not found"}`), nil
		}
	}
}
