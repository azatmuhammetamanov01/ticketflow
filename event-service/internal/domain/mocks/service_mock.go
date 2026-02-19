package mocks

import (
	"context"
	"time"

	"github.com/azatmuhammetamanov01/online-ticket-booking/event-service/internal/domain"
	"github.com/stretchr/testify/mock"
)

type MockEventService struct {
	mock.Mock
}

func (m *MockEventService) CreateEvent(ctx context.Context, name string, startTime time.Time, totalSeats int32) (*domain.Event, error) {
	args := m.Called(ctx, name, startTime, totalSeats)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.Event), args.Error(1)
}

func (m *MockEventService) GetEvent(ctx context.Context, eventID string) (*domain.Event, error) {
	args := m.Called(ctx, eventID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.Event), args.Error(1)
}

func (m *MockEventService) ListEvents(ctx context.Context, limit, offset int32) ([]*domain.Event, int32, error) {
	args := m.Called(ctx, limit, offset)
	if args.Get(0) == nil {
		return nil, args.Get(1).(int32), args.Error(2)
	}
	return args.Get(0).([]*domain.Event), args.Get(1).(int32), args.Error(2)
}

func (m *MockEventService) UpdateAvailableTickets(ctx context.Context, eventID string, quantity int32) (int32, error) {
	args := m.Called(ctx, eventID, quantity)
	return args.Get(0).(int32), args.Error(1)
}
