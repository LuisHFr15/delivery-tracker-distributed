package http

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/LuisHFr15/delivery-tracker-distributed/internal/ingester/app/dtos"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type FakeIngesterService struct {
	IngestedOrders    []dtos.OrderEventDTO
	IngestedLocations []dtos.LocationEventDTO
	errToReturn       error
}

func (f *FakeIngesterService) IngestOrder(ctx context.Context, order dtos.OrderEventDTO) error {
	if f.errToReturn != nil {
		return f.errToReturn
	}
	f.IngestedOrders = append(f.IngestedOrders, order)
	return nil
}

func (f *FakeIngesterService) IngestLocation(ctx context.Context, dto dtos.LocationEventDTO, orderId uuid.UUID) error {
	if f.errToReturn != nil {
		return f.errToReturn
	}
	f.IngestedLocations = append(f.IngestedLocations, dto)
	return nil
}

func TestIngesterHandler_TrackOrder(t *testing.T) {
	gin.SetMode(gin.TestMode)

	tests := []struct {
		name         string
		payload      interface{}
		serviceErr   error
		expectedCode int
	}{
		{
			name: "success",
			payload: dtos.OrderEventDTO{
				Order: dtos.OrderDTO{ID: uuid.New()},
			},
			expectedCode: http.StatusAccepted,
		},
		{
			name:         "invalid json",
			payload:      "invalid json string",
			expectedCode: http.StatusBadRequest,
		},
		{
			name: "service error",
			payload: dtos.OrderEventDTO{
				Order: dtos.OrderDTO{ID: uuid.New()},
			},
			serviceErr:   errors.New("db error"),
			expectedCode: http.StatusInternalServerError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			fakeService := &FakeIngesterService{errToReturn: tt.serviceErr}
			handler := NewIngesterHandler(fakeService)

			w := httptest.NewRecorder()
			c, _ := gin.CreateTestContext(w)

			var reqBody []byte
			var err error
			if s, ok := tt.payload.(string); ok {
				reqBody = []byte(s)
			} else {
				reqBody, err = json.Marshal(tt.payload)
				if err != nil {
					t.Fatalf("failed to marshal payload: %v", err)
				}
			}

			c.Request, err = http.NewRequest(http.MethodPost, "/order", bytes.NewBuffer(reqBody))
			if err != nil {
				t.Fatalf("failed to create request: %v", err)
			}
			c.Request.Header.Set("Content-Type", "application/json")

			handler.TrackOrder(c)

			if c.Writer.Status() != tt.expectedCode {
				t.Errorf("expected status %d, got %d", tt.expectedCode, c.Writer.Status())
			}
		})
	}
}

func TestIngesterHandler_UpdateLocation(t *testing.T) {
	gin.SetMode(gin.TestMode)

	validOrderID := uuid.New()

	tests := []struct {
		name         string
		orderIDPath  string
		payload      interface{}
		serviceErr   error
		expectedCode int
	}{
		{
			name:        "success",
			orderIDPath: validOrderID.String(),
			payload: dtos.LocationEventDTO{
				Latitude:  -23.5505,
				Longitude: -46.6333,
			},
			expectedCode: http.StatusAccepted,
		},
		{
			name:         "invalid order id format",
			orderIDPath:  "invalid-uuid",
			payload:      dtos.LocationEventDTO{},
			expectedCode: http.StatusBadRequest,
		},
		{
			name:         "invalid json body",
			orderIDPath:  validOrderID.String(),
			payload:      "invalid json string",
			expectedCode: http.StatusBadRequest,
		},
		{
			name:        "service error",
			orderIDPath: validOrderID.String(),
			payload: dtos.LocationEventDTO{
				Latitude:  -23.5505,
				Longitude: -46.6333,
			},
			serviceErr:   errors.New("failed to process"),
			expectedCode: http.StatusInternalServerError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			fakeService := &FakeIngesterService{errToReturn: tt.serviceErr}
			handler := NewIngesterHandler(fakeService)

			w := httptest.NewRecorder()
			c, _ := gin.CreateTestContext(w)

			c.Params = gin.Params{{Key: "id", Value: tt.orderIDPath}}

			var reqBody []byte
			var err error
			if s, ok := tt.payload.(string); ok {
				reqBody = []byte(s)
			} else {
				reqBody, err = json.Marshal(tt.payload)
				if err != nil {
					t.Fatalf("failed to marshal payload: %v", err)
				}
			}

			c.Request, err = http.NewRequest(http.MethodPost, "/order/"+tt.orderIDPath+"/location", bytes.NewBuffer(reqBody))
			if err != nil {
				t.Fatalf("failed to create request: %v", err)
			}
			c.Request.Header.Set("Content-Type", "application/json")

			handler.UpdateLocation(c)

			if c.Writer.Status() != tt.expectedCode {
				t.Errorf("expected status %d, got %d. Body: %s", tt.expectedCode, c.Writer.Status(), w.Body.String())
			}
		})
	}
}
