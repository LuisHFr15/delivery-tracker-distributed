package services

import (
	"errors"
	"time"

	"github.com/LuisHFr15/delivery-tracker-distributed/internal/processor/domain/models/order"
	"github.com/umahmood/haversine"
)

var ErrNilLocation = errors.New("location cannot be nil")

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

func (d *DeliveryCalculator) CalculateTime(actual *order.Location, final *order.Location) (time.Duration, error) {
	if actual == nil || final == nil {
		return 0, ErrNilLocation
	}

	distance := d.calculateDistance(*actual, *final)
	// assuming 40km/h as the standard
	speed := 40.0
	timeHours := distance / speed
	// round to the nearest minute for a human-friendly ETA
	return time.Duration(timeHours * float64(time.Hour)).Round(time.Minute), nil
}
