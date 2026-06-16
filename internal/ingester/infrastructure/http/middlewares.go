package http

import (
	"context"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
)

func MetricsMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()
		c.Next()
		_ = time.Since(start)
	}
}

func TimeoutMiddleware(timeout time.Duration) gin.HandlerFunc {
	return func(c *gin.Context) {
		ctx, cancel := context.WithTimeout(c.Request.Context(), timeout)
		defer cancel()

		c.Request = c.Request.WithContext(ctx)
		finished := make(chan struct{}, 1)

		go func() {
			c.Next()
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
