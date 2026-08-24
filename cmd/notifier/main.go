package main

import (
	"context"
	"log"
	"os/signal"
	"sync"
	"syscall"
	"time"

	"github.com/LuisHFr15/delivery-tracker-distributed/internal/notifier/infrastructure/messaging"
	"github.com/LuisHFr15/delivery-tracker-distributed/internal/notifier/infrastructure/messaging/ports"
)

func main() {
	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	var notificationTopicReader ports.NotificationReader = messaging.NewKafkaReadNotificationTopic()

	var readersWg sync.WaitGroup
	readersWg.Add(1)
	go func() {
		defer readersWg.Done()
		for {
			notification, err := notificationTopicReader.Read(ctx)
			if err != nil {
				if ctx.Err() != nil {
					log.Println("Notification reader stopping due to shutdown signal.")
					return
				}
				log.Printf("Failed to read notification topic: %v", err)
				continue
			}

			log.Printf("notified the client %s about the order %s location update\n", notification.ClientId, notification.OrderId)
		}
	}()

	log.Println("Notifier service started and listening for processed orders...")

	<-ctx.Done()
	log.Println("Shutdown signal received, starting graceful teardown...")

	shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	log.Println("Closing Kafka connections...")
	err := notificationTopicReader.Close()
	if err != nil {
		log.Printf("Failed to close Kafka connection: %v", err)
	}

	readersWg.Wait()
	log.Println("Readers stopped.")

	select {
	case <-shutdownCtx.Done():
		log.Println("Shutdown timed out! Forcing exit.")
	default:
		log.Println("Shutdown complete gracefully")
	}
}
