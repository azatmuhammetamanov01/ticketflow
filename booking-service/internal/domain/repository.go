package domain

import "context"

type BookingRepository interface {
	Create(ctx context.Context, booking *Booking) error
	GetByID(ctx context.Context, id string) (*Booking, error)
	ListByUserID(ctx context.Context, userID string) ([]*Booking, error)
	UpdateStatus(ctx context.Context, id string, status BookingStatus) error
}
