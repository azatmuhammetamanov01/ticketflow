package domain

import "context"

type BookingService interface {
	CreateBooking(ctx context.Context, userID, eventID string, ticketCount int32) (*Booking, error)
	GetBooking(ctx context.Context, bookingID string) (*Booking, error)
	ListUserBookings(ctx context.Context, userID string) ([]*Booking, error)
	CancelBooking(ctx context.Context, bookingID string) error
}
