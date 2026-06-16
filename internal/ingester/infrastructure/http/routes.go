package http

import (
	"time"

	"github.com/gin-gonic/gin"
)

func RegisterIngesterRoutes(router *gin.RouterGroup, handler *IngesterHandler) {
	router.Use(MetricsMiddleware(), TimeoutMiddleware(time.Millisecond*100))

	router.POST("/order", handler.CreateOrder)
	router.POST("/order/:id/location", handler.UpdateLocation)
}
