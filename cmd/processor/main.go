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
	"github.com/LuisHFr15/delivery-tracker-distributed/internal/processor/infrastructure/messaging"
	"github.com/LuisHFr15/delivery-tracker-distributed/internal/processor/infrastructure/messaging/ports"
)

func main() {
	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	auditingRepo := dynamo.NewDynamoAuditingEventRepository(ctx)
	processedRepo := dynamo.NewDynamoProcessedOrderRepository(ctx)
	orderRepo := dynamo.NewDynamoOrderRepository(ctx)
	var orderTopicReader ports.OrderReader = messaging.NewKafkaReadOrderTopic()
	var locationTopicReader ports.LocationReader = messaging.NewKafkaReadLocationTopic()
	var notificationWriter ports.NotificationWriter = messaging.NewKafkaWriteNotification()

	go auditingRepo.RunWorker()
	go processedRepo.RunWorker()
	go orderRepo.RunWorker()
	go notificationWriter.RunWorker()

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

			orderService := services.NewOrderEventService(cls, auditingRepo, orderRepo)
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
			locationService := services.NewLocationEventService(cls, auditingRepo, processedRepo, orderRepo, notificationWriter)
			if err := locationService.ProcessEvent(); err != nil {
				log.Printf("Failed to process location event: %v", err)
			}
		}
	}()

	<-ctx.Done()
	log.Println("Shutdown signal received, starting graceful teardown...")

	shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	log.Println("Closing Kafka connections...")
	err := orderTopicReader.Close()
	if err != nil {
		log.Printf("Failed to close Kafka connection¡s: %v", err)
	}
	err = locationTopicReader.Close()
	if err != nil {
		log.Printf("Failed to close Kafka connections: %v", err)
	}

	readersWg.Wait()
	log.Println("Readers stopped.")

	var outputWg sync.WaitGroup
	shutdownOutput := func(name string, closeFunc func() error) {
		outputWg.Add(1)
		go func() {
			defer outputWg.Done()
			err := closeFunc()
			if err != nil {
				log.Printf("Failed to close connection to component %s: %v", name, err)
			}
			log.Printf("Shutdown component %s", name)
		}()
	}

	shutdownOutput("Processed Order Repo", processedRepo.StopWorker)
	shutdownOutput("Order Repo", orderRepo.StopWorker)
	shutdownOutput("Auditing Repo", auditingRepo.StopWorker)
	shutdownOutput("Kafka Writer", notificationWriter.StopWorker)

	shutdownComplete := make(chan struct{})
	go func() {
		outputWg.Wait()
		close(shutdownComplete)
	}()

	select {
	case <-shutdownComplete:
		log.Println("Shutdown complete gracefully")
	case <-shutdownCtx.Done():
		log.Println("Shutdown timed out! Forcing exit.")
	}
}
