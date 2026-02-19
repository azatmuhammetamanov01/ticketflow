package domain

import "time"

type BookingStatus int32

const (
	BookingStatusUnspecified BookingStatus = 0
	BookingStatusPending    BookingStatus = 1
	BookingStatusConfirmed  BookingStatus = 2
	BookingStatusCancelled  BookingStatus = 3
)

type Booking struct {
	ID          string
	UserID      string
	EventID     string
	TicketCount int32
	Status      BookingStatus
	CreatedAt   time.Time
}
