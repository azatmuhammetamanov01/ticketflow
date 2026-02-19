package domain

import "time"

type Event struct {
	ID             string
	Name           string
	StartTime      time.Time
	TotalSeats     int32
	AvailableSeats int32
	CreatedAt      time.Time
}
