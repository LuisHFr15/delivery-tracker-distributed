package ingester

import (
	"context"
	"main/internal/domain/models/events"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type EventPublisher interface {
	PublishOrderEvent(ctx context.Context, event events.OrderEvent) error
	PublishLocationEvent(ctx context.Context, event events.LocationEvent) error
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
	router.Use(MetricsMiddleware(), TimeoutMiddleware(time.Millisecond*100))
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

	if err := h.publisher.PublishOrderEvent(c.Request.Context(), event); err != nil {
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

	if err := h.publisher.PublishLocationEvent(c.Request.Context(), event); err != nil {
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

func TimeoutMiddleware(timeout time.Duration) gin.HandlerFunc {
	return func(c *gin.Context) {
		/*
			create and set the context with timeout of 100ms as the description said this is the limit
			Go validates through the runtime execution the time passed since the init of the execution, so it is pretty fast
			without even using CPU resources
			In parallel, the main thread has a timer manager in order to close the context when it is done
		*/
		ctx, cancel := context.WithTimeout(c.Request.Context(), timeout)
		defer cancel()
		c.Request = c.Request.WithContext(ctx)

		// create a channel that only receives a signal related to finish the main execution
		finished := make(chan struct{}, 1)

		// goroutine to stablish in a new thread the main execution in order to allow parallel execution
		go func() {
			c.Next()
			/*
				create a struct with no value only to say that there is something
				the same could be done with boolean, but we used struct{} since it can optimize resource usage at the max
			*/
			finished <- struct{}{}
		}()

		select {
		case <-finished:
			return
		case <-ctx.Done():
			c.AbortWithStatusJSON(http.StatusRequestTimeout, gin.H{
				"error":   "request timed out",
				"details": "server did not answered in 100ms",
			})
			return
		}
	}
}
