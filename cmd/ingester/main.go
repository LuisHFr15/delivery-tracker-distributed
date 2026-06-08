package main

import (
	"fmt"

	"github.com/gin-gonic/gin"
)

func main() {
	fmt.Println("ingester starting...")
	r := gin.Default()
	group := r.Group("/api")

	group.GET("/test", func(c *gin.Context) {
		c.JSON(200, gin.H{"hello": "world"})
		return
	})
	r.Run(":8080")
}
