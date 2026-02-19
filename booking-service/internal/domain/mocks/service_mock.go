package mocks

import (
	"context"

	"github.com/azatmuhammetamanov01/online-ticket-booking/booking-service/internal/domain"
	"github.com/stretchr/testify/mock"
)

type MockBookingService struct {
	mock.Mock
}

func (m *MockBookingService) CreateBooking(ctx context.Context, userID, eventID string, ticketCount int32) (*domain.Booking, error) {
	args := m.Called(ctx, userID, eventID, ticketCount)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.Booking), args.Error(1)
}

func (m *MockBookingService) GetBooking(ctx context.Context, bookingID string) (*domain.Booking, error) {
	args := m.Called(ctx, bookingID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.Booking), args.Error(1)
}

func (m *MockBookingService) ListUserBookings(ctx context.Context, userID string) ([]*domain.Booking, error) {
	args := m.Called(ctx, userID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*domain.Booking), args.Error(1)
}

func (m *MockBookingService) CancelBooking(ctx context.Context, bookingID string) error {
	args := m.Called(ctx, bookingID)
	return args.Error(0)
}
