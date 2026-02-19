package domain

import "context"

type EventRepository interface {
	Create(ctx context.Context, event *Event) error
	GetByID(ctx context.Context, id string) (*Event, error)
	List(ctx context.Context, limit, offset int32) ([]*Event, int32, error)
	UpdateAvailableSeats(ctx context.Context, id string, quantity int32) (int32, error)
}
