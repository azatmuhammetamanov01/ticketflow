package mocks

import (
	"context"

	eventpb "github.com/azatmuhammetamanov01/online-ticket-booking/booking-service/event"
	"github.com/stretchr/testify/mock"
)

type MockEventClient struct {
	mock.Mock
}

func (m *MockEventClient) GetEvent(ctx context.Context, eventID string) (*eventpb.Event, error) {
	args := m.Called(ctx, eventID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*eventpb.Event), args.Error(1)
}

func (m *MockEventClient) ReserveTickets(ctx context.Context, eventID string, quantity int32) error {
	args := m.Called(ctx, eventID, quantity)
	return args.Error(0)
}

func (m *MockEventClient) ReleaseTickets(ctx context.Context, eventID string, quantity int32) error {
	args := m.Called(ctx, eventID, quantity)
	return args.Error(0)
}

func (m *MockEventClient) Close() error {
	args := m.Called()
	return args.Error(0)
}
