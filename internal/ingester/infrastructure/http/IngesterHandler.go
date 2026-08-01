package http

import (
	"context"
	"net/http"

	"github.com/LuisHFr15/delivery-tracker-distributed/internal/ingester/app/dtos"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type IngesterServicer interface {
	IngestOrder(ctx context.Context, order dtos.OrderEventDTO) error
	IngestLocation(ctx context.Context, dto dtos.LocationEventDTO, orderId uuid.UUID) error
	Health(ctx context.Context) error
}

type IngesterHandler struct {
	service IngesterServicer
}

func NewIngesterHandler(service IngesterServicer) *IngesterHandler {
	return &IngesterHandler{
		service: service,
	}
}

func (h *IngesterHandler) TrackOrder(c *gin.Context) {
	var orderDto dtos.OrderEventDTO
	if err := c.ShouldBindJSON(&orderDto); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request format", "details": err.Error()})
		return
	}

	if err := h.service.IngestOrder(c.Request.Context(), orderDto); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to process order", "details": err.Error()})
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

	var locationDto dtos.LocationEventDTO
	if err := c.ShouldBindJSON(&locationDto); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request format"})
		return
	}

	if err := h.service.IngestLocation(c.Request.Context(), locationDto, orderId); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to process location", "details": err.Error()})
		return
	}

	c.Status(http.StatusAccepted)
}

func (h *IngesterHandler) Health(c *gin.Context) {
	if err := h.service.Health(c.Request.Context()); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
	}
	c.JSON(http.StatusOK, gin.H{"status": "ok"})
}
