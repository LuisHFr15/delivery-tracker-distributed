package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	infrahttp "github.com/LuisHFr15/delivery-tracker-distributed/internal/ingester/infrastructure/http"
	"github.com/LuisHFr15/delivery-tracker-distributed/internal/ingester/infrastructure/queues"

	"github.com/LuisHFr15/delivery-tracker-distributed/internal/ingester/app/services"

	"github.com/gin-gonic/gin"
)

func main() {
	if os.Getenv("APP_RUNTIME") == "lambda" {
		runLambda()
		return
	}
	runServer()
}

// runServer is the local, long-running mode: an HTTP server publishing to Kafka,
// with a graceful shutdown on SIGINT/SIGTERM.
func runServer() {
	fmt.Println("Ingester starting...")
	errCh := make(chan error, 1)

	publisher := queues.NewKafkaPublisher()
	publisher.Start(errCh)
	service := services.NewIngesterService(publisher)

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
	select {
	case <-quit:
		log.Println("Shutting down gracefully")
	case err := <-errCh:
		log.Println("Critical error:", err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		log.Fatalf("shutdown failed: %v", err)
	}
	if err := publisher.Close(); err != nil {
		log.Fatalf("close publisher failed: %v", err)
	}
	fmt.Println("Server stopped cleanly")
}
