package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"net/http"

	infrahttp "github.com/LuisHFr15/delivery-tracker-distributed/internal/ingester/infrastructure/http"
	"github.com/LuisHFr15/delivery-tracker-distributed/internal/ingester/infrastructure/queues"

	"github.com/LuisHFr15/delivery-tracker-distributed/internal/ingester/app/services"

	"github.com/gin-gonic/gin"
)

func main() {
	fmt.Println("Ingester starting...")

	publisher := queues.NewKafkaPublisher()
	service := services.NewIngesterService(publisher)
	defer publisher.Close()

	handler := infrahttp.NewIngesterHandler(service)

	r := gin.Default()
	api := r.Group("/api")

	infrahttp.RegisterIngesterRoutes(api, handler)

	fmt.Println("Ingester server running on :8080")
	srv := &http.Server{Addr: ":8080", Handler: r}

	go func() {
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("listen: %s\n", err)
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	/*
		blocking operation until receive SIGINT or SIGTERM
	*/
	<-quit
	fmt.Println("Shutting down...")
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	publisher.Close()

	if err := srv.Shutdown(ctx); err != nil {
		log.Fatalf("shutdown failed: %v", err)
	}
	fmt.Println("Server stopped cleanly")
}
