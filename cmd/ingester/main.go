package main

import (
	"fmt"
	"main/internal/app/ingester"
	"main/internal/infrastructure/queues"

	"github.com/gin-gonic/gin"
)

func main() {
	fmt.Println("Ingester starting...")

	publisher := queues.NewKafkaPublisher()
	/*
		defer statement: follow LIFO, aka, stack-like steps
		it executes the deferred function after the surrounding function declares it is returning something (error, value, nil)
		in this case, is like delaying the close Kafka connection when the program ends its execution

		it executes exactly after the return arguments are set, but before exiting the surrounding function
	*/
	defer publisher.Close()

	handler := ingester.NewIngesterHandler(publisher)

	r := gin.Default()

	api := r.Group("/api")

	handler.RegisterRoutes(api)

	fmt.Println("Ingester server running on :8080")
	r.Run(":8080")
}
