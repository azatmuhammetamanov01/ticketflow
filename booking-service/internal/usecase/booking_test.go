package usecase

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/azatmuhammetamanov01/online-ticket-booking/booking-service/internal/client"
	"github.com/azatmuhammetamanov01/online-ticket-booking/booking-service/internal/domain"
	"github.com/azatmuhammetamanov01/online-ticket-booking/booking-service/internal/domain/mocks"
	eventpb "github.com/azatmuhammetamanov01/online-ticket-booking/booking-service/event"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func newTestUsecase() (*BookingUsecase, *mocks.MockBookingRepository, *mocks.MockEventClient) {
	repo := new(mocks.MockBookingRepository)
	eventClient := new(mocks.MockEventClient)
	uc := NewBookingUsecase(repo, eventClient)
	return uc, repo, eventClient
}

func TestCreateBooking_Success(t *testing.T) {
	uc, repo, eventClient := newTestUsecase()
	ctx := context.Background()

	eventClient.On("GetEvent", ctx, "event-1").Return(&eventpb.Event{
		Id:             "event-1",
		AvailableSeats: 100,
	}, nil)
	eventClient.On("ReserveTickets", ctx, "event-1", int32(2)).Return(nil)
	repo.On("Create", ctx, mock.AnythingOfType("*domain.Booking")).Return(nil)

	booking, err := uc.CreateBooking(ctx, "user-1", "event-1", 2)

	assert.NoError(t, err)
	assert.NotNil(t, booking)
	assert.Equal(t, "user-1", booking.UserID)
	assert.Equal(t, "event-1", booking.EventID)
	assert.Equal(t, int32(2), booking.TicketCount)
	repo.AssertExpectations(t)
	eventClient.AssertExpectations(t)
}

func TestCreateBooking_EmptyUserID(t *testing.T) {
	uc, _, _ := newTestUsecase()

	booking, err := uc.CreateBooking(context.Background(), "", "event-1", 2)

	assert.Nil(t, booking)
	assert.ErrorIs(t, err, domain.ErrInvalidInput)
}

func TestCreateBooking_EmptyEventID(t *testing.T) {
	uc, _, _ := newTestUsecase()

	booking, err := uc.CreateBooking(context.Background(), "user-1", "", 2)

	assert.Nil(t, booking)
	assert.ErrorIs(t, err, domain.ErrInvalidInput)
}

func TestCreateBooking_ZeroTickets(t *testing.T) {
	uc, _, _ := newTestUsecase()

	booking, err := uc.CreateBooking(context.Background(), "user-1", "event-1", 0)

	assert.Nil(t, booking)
	assert.ErrorIs(t, err, domain.ErrInvalidInput)
}

func TestCreateBooking_EventNotFound(t *testing.T) {
	uc, _, eventClient := newTestUsecase()
	ctx := context.Background()

	eventClient.On("GetEvent", ctx, "event-1").Return(nil, client.ErrEventNotFound)

	booking, err := uc.CreateBooking(ctx, "user-1", "event-1", 2)

	assert.Nil(t, booking)
	assert.ErrorIs(t, err, domain.ErrEventNotFound)
}

func TestCreateBooking_InsufficientSeats(t *testing.T) {
	uc, _, eventClient := newTestUsecase()
	ctx := context.Background()

	eventClient.On("GetEvent", ctx, "event-1").Return(&eventpb.Event{
		Id:             "event-1",
		AvailableSeats: 1,
	}, nil)

	booking, err := uc.CreateBooking(ctx, "user-1", "event-1", 5)

	assert.Nil(t, booking)
	assert.ErrorIs(t, err, domain.ErrInsufficientSeats)
}

func TestCreateBooking_ReserveTicketsFails(t *testing.T) {
	uc, _, eventClient := newTestUsecase()
	ctx := context.Background()

	eventClient.On("GetEvent", ctx, "event-1").Return(&eventpb.Event{
		Id:             "event-1",
		AvailableSeats: 100,
	}, nil)
	eventClient.On("ReserveTickets", ctx, "event-1", int32(2)).Return(client.ErrInsufficientSeats)

	booking, err := uc.CreateBooking(ctx, "user-1", "event-1", 2)

	assert.Nil(t, booking)
	assert.ErrorIs(t, err, domain.ErrInsufficientSeats)
}

func TestGetBooking_Success(t *testing.T) {
	uc, repo, _ := newTestUsecase()
	ctx := context.Background()

	expected := &domain.Booking{
		ID:          "booking-1",
		UserID:      "user-1",
		EventID:     "event-1",
		TicketCount: 2,
		Status:      domain.BookingStatusPending,
		CreatedAt:   time.Now(),
	}
	repo.On("GetByID", ctx, "booking-1").Return(expected, nil)

	booking, err := uc.GetBooking(ctx, "booking-1")

	assert.NoError(t, err)
	assert.Equal(t, expected, booking)
}

func TestGetBooking_EmptyID(t *testing.T) {
	uc, _, _ := newTestUsecase()

	booking, err := uc.GetBooking(context.Background(), "")

	assert.Nil(t, booking)
	assert.ErrorIs(t, err, domain.ErrInvalidInput)
}

func TestGetBooking_NotFound(t *testing.T) {
	uc, repo, _ := newTestUsecase()
	ctx := context.Background()

	repo.On("GetByID", ctx, "booking-1").Return(nil, nil)

	booking, err := uc.GetBooking(ctx, "booking-1")

	assert.Nil(t, booking)
	assert.ErrorIs(t, err, domain.ErrBookingNotFound)
}

func TestListUserBookings_Success(t *testing.T) {
	uc, repo, _ := newTestUsecase()
	ctx := context.Background()

	expected := []*domain.Booking{
		{ID: "b-1", UserID: "user-1", EventID: "event-1", TicketCount: 2},
		{ID: "b-2", UserID: "user-1", EventID: "event-2", TicketCount: 1},
	}
	repo.On("ListByUserID", ctx, "user-1").Return(expected, nil)

	bookings, err := uc.ListUserBookings(ctx, "user-1")

	assert.NoError(t, err)
	assert.Len(t, bookings, 2)
}

func TestListUserBookings_EmptyUserID(t *testing.T) {
	uc, _, _ := newTestUsecase()

	bookings, err := uc.ListUserBookings(context.Background(), "")

	assert.Nil(t, bookings)
	assert.ErrorIs(t, err, domain.ErrInvalidInput)
}

func TestCancelBooking_Success(t *testing.T) {
	uc, repo, eventClient := newTestUsecase()
	ctx := context.Background()

	booking := &domain.Booking{
		ID:          "booking-1",
		UserID:      "user-1",
		EventID:     "event-1",
		TicketCount: 3,
		Status:      domain.BookingStatusPending,
	}
	repo.On("GetByID", ctx, "booking-1").Return(booking, nil)
	eventClient.On("ReleaseTickets", ctx, "event-1", int32(3)).Return(nil)
	repo.On("UpdateStatus", ctx, "booking-1", domain.BookingStatusCancelled).Return(nil)

	err := uc.CancelBooking(ctx, "booking-1")

	assert.NoError(t, err)
	repo.AssertExpectations(t)
	eventClient.AssertExpectations(t)
}

func TestCancelBooking_EmptyID(t *testing.T) {
	uc, _, _ := newTestUsecase()

	err := uc.CancelBooking(context.Background(), "")

	assert.ErrorIs(t, err, domain.ErrInvalidInput)
}

func TestCancelBooking_NotFound(t *testing.T) {
	uc, repo, _ := newTestUsecase()
	ctx := context.Background()

	repo.On("GetByID", ctx, "booking-1").Return(nil, nil)

	err := uc.CancelBooking(ctx, "booking-1")

	assert.ErrorIs(t, err, domain.ErrBookingNotFound)
}

func TestCancelBooking_AlreadyCancelled(t *testing.T) {
	uc, repo, _ := newTestUsecase()
	ctx := context.Background()

	booking := &domain.Booking{
		ID:     "booking-1",
		Status: domain.BookingStatusCancelled,
	}
	repo.On("GetByID", ctx, "booking-1").Return(booking, nil)

	err := uc.CancelBooking(ctx, "booking-1")

	assert.ErrorIs(t, err, domain.ErrAlreadyCancelled)
}

func TestCancelBooking_ReleaseTicketsFails(t *testing.T) {
	uc, repo, eventClient := newTestUsecase()
	ctx := context.Background()

	booking := &domain.Booking{
		ID:          "booking-1",
		EventID:     "event-1",
		TicketCount: 3,
		Status:      domain.BookingStatusPending,
	}
	repo.On("GetByID", ctx, "booking-1").Return(booking, nil)
	eventClient.On("ReleaseTickets", ctx, "event-1", int32(3)).Return(errors.New("event service unavailable"))

	err := uc.CancelBooking(ctx, "booking-1")

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "event service unavailable")
}
