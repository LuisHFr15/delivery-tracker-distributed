package http

import (
	"main/internal/ingester/app/dtos"
	"main/internal/ingester/app/services"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type IngesterHandler struct {
	service *services.IngesterService
}

func NewIngesterHandler(service *services.IngesterService) *IngesterHandler {
	return &IngesterHandler{
		service: service,
	}
}

func (h *IngesterHandler) CreateOrder(c *gin.Context) {
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
