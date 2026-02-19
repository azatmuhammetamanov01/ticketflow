package usecase

import (
	"context"
	"time"

	"github.com/azatmuhammetamanov01/online-ticket-booking/event-service/internal/domain"
)

type EventUsecase struct {
	repo domain.EventRepository
}

func NewEventUsecase(repo domain.EventRepository) *EventUsecase {
	return &EventUsecase{repo: repo}
}

func (u *EventUsecase) CreateEvent(ctx context.Context, name string, startTime time.Time, totalSeats int32) (*domain.Event, error) {
	if name == "" {
		return nil, domain.ErrInvalidInput
	}
	if totalSeats <= 0 {
		return nil, domain.ErrInvalidInput
	}
	if startTime.IsZero() {
		return nil, domain.ErrInvalidInput
	}

	event := &domain.Event{
		Name:       name,
		StartTime:  startTime,
		TotalSeats: totalSeats,
	}

	if err := u.repo.Create(ctx, event); err != nil {
		return nil, err
	}

	return event, nil
}

func (u *EventUsecase) GetEvent(ctx context.Context, eventID string) (*domain.Event, error) {
	if eventID == "" {
		return nil, domain.ErrInvalidInput
	}

	event, err := u.repo.GetByID(ctx, eventID)
	if err != nil {
		return nil, err
	}
	if event == nil {
		return nil, domain.ErrEventNotFound
	}

	return event, nil
}

func (u *EventUsecase) ListEvents(ctx context.Context, limit, offset int32) ([]*domain.Event, int32, error) {
	return u.repo.List(ctx, limit, offset)
}

func (u *EventUsecase) UpdateAvailableTickets(ctx context.Context, eventID string, quantity int32) (int32, error) {
	if eventID == "" {
		return 0, domain.ErrInvalidInput
	}
	if quantity == 0 {
		return 0, domain.ErrInvalidInput
	}

	event, err := u.repo.GetByID(ctx, eventID)
	if err != nil {
		return 0, err
	}
	if event == nil {
		return 0, domain.ErrEventNotFound
	}

	if quantity > 0 && event.AvailableSeats < quantity {
		return 0, domain.ErrInsufficientSeats
	}

	newAvailable, err := u.repo.UpdateAvailableSeats(ctx, eventID, quantity)
	if err != nil {
		return 0, err
	}
	if newAvailable == 0 && quantity > 0 {
		return 0, domain.ErrInsufficientSeats
	}

	return newAvailable, nil
}
