package messaging

import (
	"encoding/json"
	"testing"
	"time"

	"github.com/LuisHFr15/delivery-tracker-distributed/internal/notifier/infrastructure/messaging/dtos"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
)

func TestProcessedOrderNotificationDTO_Unmarshal(t *testing.T) {
	orderID := uuid.New()
	clientID := uuid.New()
	timestamp := time.Now().UTC().Truncate(time.Second)

	rawJSON := `{
		"orderId": "` + orderID.String() + `",
		"clientId": "` + clientID.String() + `",
		"timeToDelivery": 120000000000,
		"timestamp": "` + timestamp.Format(time.RFC3339) + `",
		"actualLocation": {
			"lat": -23.55052,
			"lng": -46.633308
		},
		"finalDestination": {
			"lat": -23.561684,
			"lng": -46.655981
		}
	}`

	var notification dtos.ProcessedOrderNotificationDTO
	err := json.Unmarshal([]byte(rawJSON), &notification)

	assert.NoError(t, err)
	assert.Equal(t, orderID, notification.OrderId)
	assert.Equal(t, clientID, notification.ClientId)
	assert.Equal(t, 2*time.Minute, notification.TimeToDelivery)
	assert.Equal(t, -23.55052, notification.ActualLocation.Lat)
	assert.Equal(t, -46.633308, notification.ActualLocation.Lng)
	assert.Equal(t, -23.561684, notification.FinalDestination.Lat)
	assert.Equal(t, -46.655981, notification.FinalDestination.Lng)
}
