package mocks

import (
	"context"

	"github.com/azatmuhammetamanov01/online-ticket-booking/event-service/internal/domain"
	"github.com/stretchr/testify/mock"
)

type MockEventRepository struct {
	mock.Mock
}

func (m *MockEventRepository) Create(ctx context.Context, event *domain.Event) error {
	args := m.Called(ctx, event)
	return args.Error(0)
}

func (m *MockEventRepository) GetByID(ctx context.Context, id string) (*domain.Event, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.Event), args.Error(1)
}

func (m *MockEventRepository) List(ctx context.Context, limit, offset int32) ([]*domain.Event, int32, error) {
	args := m.Called(ctx, limit, offset)
	if args.Get(0) == nil {
		return nil, args.Get(1).(int32), args.Error(2)
	}
	return args.Get(0).([]*domain.Event), args.Get(1).(int32), args.Error(2)
}

func (m *MockEventRepository) UpdateAvailableSeats(ctx context.Context, id string, quantity int32) (int32, error) {
	args := m.Called(ctx, id, quantity)
	return args.Get(0).(int32), args.Error(1)
}
