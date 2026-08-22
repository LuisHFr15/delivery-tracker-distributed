package services

import "errors"

// ErrMissingEventID is returned when an inbound event has no correlation id.
var ErrMissingEventID = errors.New("event has no event id")
