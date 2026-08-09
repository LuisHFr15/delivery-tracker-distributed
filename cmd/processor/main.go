package main

import (
	"context"
	"log"
	"os/signal"
	"sync"
	"syscall"
	"time"

	"github.com/LuisHFr15/delivery-tracker-distributed/internal/processor/app/services"
	"github.com/LuisHFr15/delivery-tracker-distributed/internal/processor/infrastructure/data/dynamo"
	infrastructure "github.com/LuisHFr15/delivery-tracker-distributed/internal/processor/infrastructure/queue"
)

func main() {
	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	auditingRepo := dynamo.NewDynamoAuditingEventRepository(ctx)
	processedRepo := dynamo.NewDynamoProcessedOrderRepository(ctx)
	orderTopicReader := infrastructure.NewKafkaReadOrderTopic()
	locationTopicReader := infrastructure.NewKafkaReadLocationTopic()

	auditingRepo.RunWorker()
	processedRepo.RunWorker()

	var readersWg sync.WaitGroup

	readersWg.Add(1)
	go func() {
		defer readersWg.Done()
		for {
			cls, err := orderTopicReader.Read(ctx)
			if err != nil {
				// THIS IS HOW THE GRACEFUL SHUTDOWN WILL WORK HERE!!!!
				if ctx.Err() != nil {
					log.Println("Order reader stopping due to shutdown signal.")
					return
				}
				log.Printf("Failed to read order topic: %v", err)
				continue
			}

			orderService := services.NewOrderEventService(cls, auditingRepo)
			orderService.ConvertEvent()
		}
	}()

	readersWg.Add(1)
	go func() {
		defer readersWg.Done()
		for {
			cls, err := locationTopicReader.Read(ctx)
			if err != nil {
				if ctx.Err() != nil {
					log.Println("Location reader stopping due to shutdown signal.")
					return
				}
				log.Printf("Failed to read location topic: %v", err)
				continue
			}
			log.Printf("Read location topic %v", cls)
		}
	}()

	<-ctx.Done()
	log.Println("Shutdown signal received, starting graceful teardown...")

	shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	log.Println("Closing Kafka connections...")
	err := orderTopicReader.Close()
	if err != nil {
		log.Printf("Failed to close Kafka connections: %v", err)
	}
	err = locationTopicReader.Close()
	if err != nil {
		log.Printf("Failed to close Kafka connections: %v", err)
	}

	readersWg.Wait()
	log.Println("Readers stopped.")

	var reposWg sync.WaitGroup
	shutdownRepo := func(name string, closeFunc func()) {
		reposWg.Add(1)
		go func() {
			defer reposWg.Done()
			closeFunc()
			log.Printf("Shutdown component %s", name)
		}()
	}

	shutdownRepo("Processed Order Repo", processedRepo.StopWorker)
	shutdownRepo("Auditing Repo", auditingRepo.StopWorker)

	shutdownComplete := make(chan struct{})
	go func() {
		reposWg.Wait()
		close(shutdownComplete)
	}()

	select {
	case <-shutdownComplete:
		log.Println("Shutdown complete gracefully")
	case <-shutdownCtx.Done():
		log.Println("Shutdown timed out! Forcing exit.")
	}
}
