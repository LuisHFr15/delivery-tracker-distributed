package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"

	notifierservices "github.com/LuisHFr15/delivery-tracker-distributed/internal/notifier/app/services"
	"github.com/LuisHFr15/delivery-tracker-distributed/internal/notifier/infrastructure/data/dynamo"
	"github.com/LuisHFr15/delivery-tracker-distributed/internal/notifier/infrastructure/messaging"
	msgports "github.com/LuisHFr15/delivery-tracker-distributed/internal/notifier/infrastructure/messaging/ports"
)

func main() {
	if os.Getenv("APP_RUNTIME") == "lambda" {
		runLambda()
		return
	}
	runServer()
}

// runServer is the local, long-running mode: consume processed-orders from Kafka
// and persist a notification per event, tearing down gracefully on SIGINT/SIGTERM.
func runServer() {
	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	repo := dynamo.NewDynamoNotifiedMessageRepository(ctx)
	var reader msgports.ProcessedOrderReader = messaging.NewKafkaReadProcessedOrders()

	go repo.RunWorker()

	var readersWg sync.WaitGroup
	readersWg.Add(1)
	go func() {
		defer readersWg.Done()
		for {
			dto, err := reader.Read(ctx)
			if err != nil {
				if ctx.Err() != nil {
					log.Println("Processed-orders reader stopping due to shutdown signal.")
					return
				}
				log.Printf("Failed to read processed-orders topic: %v", err)
				continue
			}
			if err := notifierservices.NewNotifierService(dto, repo).Notify(); err != nil {
				log.Printf("Failed to notify: %v", err)
			}
		}
	}()

	<-ctx.Done()
	log.Println("Shutdown signal received, starting graceful teardown...")

	shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := reader.Close(); err != nil {
		log.Printf("Failed to close Kafka connection: %v", err)
	}
	readersWg.Wait()
	log.Println("Reader stopped.")

	shutdownComplete := make(chan struct{})
	go func() {
		if err := repo.StopWorker(); err != nil {
			log.Printf("Failed to stop notified-message repo: %v", err)
		}
		close(shutdownComplete)
	}()

	select {
	case <-shutdownComplete:
		log.Println("Shutdown complete gracefully")
	case <-shutdownCtx.Done():
		log.Println("Shutdown timed out! Forcing exit.")
	}
}
