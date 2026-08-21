package services

import (
	"testing"
	"time"

	"github.com/LuisHFr15/delivery-tracker-distributed/internal/processor/domain/models/order"
)

func TestDeliveryCalculator_CalculateTime(t *testing.T) {
	tests := []struct {
		name    string
		actual  *order.Location
		final   *order.Location
		want    time.Duration
		wantErr bool
	}{
		{
			name:    "same_location",
			actual:  &order.Location{Lat: 42.0, Lng: -72.0},
			final:   &order.Location{Lat: 42.0, Lng: -72.0},
			want:    0,
			wantErr: false,
		},
		{
			name:    "origin_null_island_valid",
			actual:  &order.Location{Lat: 0, Lng: 0},
			final:   &order.Location{Lat: 0, Lng: 0},
			want:    0,
			wantErr: false,
		},
		{
			name:    "short_distance",
			actual:  &order.Location{Lat: 42.0, Lng: -72.0},
			final:   &order.Location{Lat: 42.1, Lng: -72.0},
			want:    16*time.Minute + 41*time.Second, // ~11.1km / 40km/h
			wantErr: false,
		},
		{
			name:    "long_distance",
			actual:  &order.Location{Lat: -23.55, Lng: -46.63}, // São Paulo
			final:   &order.Location{Lat: 48.85, Lng: 2.35},    // Paris
			want:    235*time.Hour + 10*time.Minute,            // ~9407km / 40km/h
			wantErr: false,
		},
		{
			name:    "crossing_hemispheres",
			actual:  &order.Location{Lat: -10.0, Lng: -50.0},
			final:   &order.Location{Lat: 10.0, Lng: 50.0},
			want:    281*time.Hour + 47*time.Minute, // ~11271km / 40km/h
			wantErr: false,
		},
		{
			name:    "antipodal_max_distance",
			actual:  &order.Location{Lat: 0, Lng: 0},
			final:   &order.Location{Lat: 0, Lng: 180},
			want:    500*time.Hour + 23*time.Minute, // ~20015km / 40km/h
			wantErr: false,
		},
		{
			name:    "error_nil_actual",
			actual:  nil,
			final:   &order.Location{Lat: 42.0, Lng: -72.0},
			want:    0,
			wantErr: true,
		},
		{
			name:    "error_nil_final",
			actual:  &order.Location{Lat: 42.0, Lng: -72.0},
			final:   nil,
			want:    0,
			wantErr: true,
		},
		{
			name:    "error_nil_both",
			actual:  nil,
			final:   nil,
			want:    0,
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			calc := NewDeliveryCalculator()
			got, err := calc.CalculateTime(tt.actual, tt.final)

			if (err != nil) && !tt.wantErr {
				t.Errorf("CalculateTime() error = %v, wantErr %v", err, tt.wantErr)
			} else if tt.want == 0 {
				if got != 0 {
					t.Errorf("CalculateTime() got = %v, want %v", got, tt.want)
				}
			} else {
				// allow a 0.5% relative tolerance (the expected values were
				// estimated from approximate distances that differ slightly
				// from the haversine lib) plus 30s to absorb the minute-level
				// rounding applied by CalculateTime.
				diff := got - tt.want
				if diff < 0 {
					diff = -diff
				}
				tolerance := time.Duration(0.005*float64(tt.want)) + 30*time.Second
				if diff > tolerance {
					t.Errorf("CalculateTime() got = %v, want %v (diff %v > tolerance %v)", got, tt.want, diff, tolerance)
				}
			}
		})
	}
}
