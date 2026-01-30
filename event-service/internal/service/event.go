package service

import (
	"context"
	"errors"
	"time"

	"github.com/azatmuhammetamanov01/online-ticket-booking/event-service/internal/repository"
)

var (
	ErrEventNotFound      = errors.New("event not found")
	ErrInvalidInput       = errors.New("invalid input")
	ErrInsufficientSeats  = errors.New("insufficient available seats")
)

type EventService interface {
	CreateEvent(ctx context.Context, name string, startTime time.Time, totalSeats int32) (*repository.Event, error)
	GetEvent(ctx context.Context, eventID string) (*repository.Event, error)
	ListEvents(ctx context.Context, limit, offset int32) ([]*repository.Event, int32, error)
	UpdateAvailableTickets(ctx context.Context, eventID string, quantity int32) (int32, error)
}

type eventService struct {
	repo repository.EventRepository
}

func NewEventService(repo repository.EventRepository) EventService {
	return &eventService{repo: repo}
}

func (s *eventService) CreateEvent(ctx context.Context, name string, startTime time.Time, totalSeats int32) (*repository.Event, error) {
	if name == "" {
		return nil, ErrInvalidInput
	}
	if totalSeats <= 0 {
		return nil, ErrInvalidInput
	}
	if startTime.IsZero() {
		return nil, ErrInvalidInput
	}

	event := &repository.Event{
		Name:       name,
		StartTime:  startTime,
		TotalSeats: totalSeats,
	}

	if err := s.repo.Create(ctx, event); err != nil {
		return nil, err
	}

	return event, nil
}

func (s *eventService) GetEvent(ctx context.Context, eventID string) (*repository.Event, error) {
	if eventID == "" {
		return nil, ErrInvalidInput
	}

	event, err := s.repo.GetByID(ctx, eventID)
	if err != nil {
		return nil, err
	}
	if event == nil {
		return nil, ErrEventNotFound
	}

	return event, nil
}

func (s *eventService) ListEvents(ctx context.Context, limit, offset int32) ([]*repository.Event, int32, error) {
	return s.repo.List(ctx, limit, offset)
}

func (s *eventService) UpdateAvailableTickets(ctx context.Context, eventID string, quantity int32) (int32, error) {
	if eventID == "" {
		return 0, ErrInvalidInput
	}
	if quantity == 0 {
		return 0, ErrInvalidInput
	}

	event, err := s.repo.GetByID(ctx, eventID)
	if err != nil {
		return 0, err
	}
	if event == nil {
		return 0, ErrEventNotFound
	}

	if quantity > 0 && event.AvailableSeats < quantity {
		return 0, ErrInsufficientSeats
	}

	newAvailable, err := s.repo.UpdateAvailableSeats(ctx, eventID, quantity)
	if err != nil {
		return 0, err
	}
	if newAvailable == 0 && quantity > 0 {
		return 0, ErrInsufficientSeats
	}

	return newAvailable, nil
}
