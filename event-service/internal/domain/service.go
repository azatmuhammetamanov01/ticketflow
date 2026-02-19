package domain

import (
	"context"
	"time"
)

type EventService interface {
	CreateEvent(ctx context.Context, name string, startTime time.Time, totalSeats int32) (*Event, error)
	GetEvent(ctx context.Context, eventID string) (*Event, error)
	ListEvents(ctx context.Context, limit, offset int32) ([]*Event, int32, error)
	UpdateAvailableTickets(ctx context.Context, eventID string, quantity int32) (int32, error)
}
