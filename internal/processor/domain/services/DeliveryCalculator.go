package services

import (
	"time"

	"github.com/LuisHFr15/delivery-tracker-distributed/internal/processor/domain/models/order"
	"github.com/umahmood/haversine"
)

type DeliveryCalculator struct {
}

func NewDeliveryCalculator() *DeliveryCalculator {
	return &DeliveryCalculator{}
}
func (d *DeliveryCalculator) calculateDistance(actual order.Location, final order.Location) float64 {
	actualCoord := haversine.Coord{Lat: actual.Lat, Lon: actual.Lng}
	finalCoord := haversine.Coord{Lat: final.Lat, Lon: final.Lng}
	_, kms := haversine.Distance(actualCoord, finalCoord)
	return kms
}

func (d *DeliveryCalculator) CalculateTime(actual order.Location, final order.Location) time.Duration {
	distance := d.calculateDistance(actual, final)

	// assuming 40km/h as the standard
	speed := 40.0
	timeHours := distance / speed
	timeMin := time.Duration(timeHours * float64(time.Hour)).Minutes()
	return time.Duration(timeMin)
}
