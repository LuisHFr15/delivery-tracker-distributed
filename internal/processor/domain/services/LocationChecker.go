package services

import "github.com/LuisHFr15/delivery-tracker-distributed/internal/processor/domain/models/order"

type LocationChecker struct {
}

func NewLocationChecker() *LocationChecker {
	return &LocationChecker{}
}

func (c *LocationChecker) IsSameLocation(past order.Location, actual order.Location) bool {
	return actual.SameLocation(past)
}
