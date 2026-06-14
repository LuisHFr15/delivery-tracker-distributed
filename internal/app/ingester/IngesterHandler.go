package ingester

import (
	"main/internal/domain/models/events"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type EventPublisher interface {
	PublishOrderEvent(event events.OrderEvent) error
	PublishLocationEvent(event events.LocationEvent) error
}

type IngesterHandler struct {
	publisher EventPublisher
}

func NewIngesterHandler(publisher EventPublisher) *IngesterHandler {
	return &IngesterHandler{
		publisher: publisher,
	}
}

func (h *IngesterHandler) RegisterRoutes(router *gin.RouterGroup) {
	router.POST("/order", h.CreateOrder)
	router.POST("/order/:id/location", h.UpdateLocation)
}

func (h *IngesterHandler) CreateOrder(c *gin.Context) {
	var event events.OrderEvent

	if err := c.ShouldBindJSON(&event); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request format", "details": err.Error()})
		return
	}

	if event.Id == uuid.Nil {
		event.Id = uuid.New()
	}

	if err := h.publisher.PublishOrderEvent(event); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to enqueue event", "details": err.Error()})
		return
	}

	c.Status(http.StatusAccepted)
}

func (h *IngesterHandler) UpdateLocation(c *gin.Context) {
	orderIdStr := c.Param("id")
	orderId, err := uuid.Parse(orderIdStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid order id format"})
		return
	}

	var event events.LocationEvent
	if err := c.ShouldBindJSON(&event); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request format"})
		return
	}

	event.OrderId = orderId

	if err := h.publisher.PublishLocationEvent(event); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to enqueue location event", "details": err.Error()})
		return
	}

	c.Status(http.StatusAccepted)
}

// TODO: define the POST method to the cloudfare tunnel /metrics endpoint with the desired metrics
func MetricsMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()
		c.Next()

		latency := time.Since(start)
		status := c.Writer.Status()
		path := c.FullPath()

		_ = latency
		_ = status
		_ = path
	}
}
