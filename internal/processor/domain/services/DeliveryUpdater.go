package services

import (
	"fmt"

	"github.com/LuisHFr15/delivery-tracker-distributed/internal/processor/domain/models/order"
	"github.com/google/uuid"
)

type DeliveryUpdater struct {
}

func NewDeliveryUpdater() *DeliveryUpdater {
	return &DeliveryUpdater{}
}

func (c *DeliveryUpdater) ProcessDelivery(ord *order.Order, actualLoc order.Location, eventId uuid.UUID) error {
	if ord.Destination == nil {
		return fmt.Errorf("delivery-updater: order %s has no destination set", ord.Id)
	}
	if !ord.Destination.SameLocation(actualLoc) {
		err := ord.StartDelivering(eventId)
		if err != nil {
			return fmt.Errorf("delivery-updater: error starting delivery-updater: %v", err)
		}
		return nil
	}
	err := ord.Deliver(eventId, actualLoc)
	if err != nil {
		return fmt.Errorf("delivery-updater: error delivering order: %v", err)
	}
	return nil
}
