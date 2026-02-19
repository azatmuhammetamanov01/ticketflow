package grpc

import (
	"context"
	"testing"
	"time"

	"github.com/azatmuhammetamanov01/online-ticket-booking/booking-service/internal/domain"
	"github.com/azatmuhammetamanov01/online-ticket-booking/booking-service/internal/domain/mocks"
	pb "github.com/azatmuhammetamanov01/online-ticket-booking/booking-service/proto"
	"github.com/stretchr/testify/assert"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func newTestHandler() (*BookingHandler, *mocks.MockBookingService) {
	svc := new(mocks.MockBookingService)
	handler := NewBookingHandler(svc)
	return handler, svc
}

func TestCreateBooking_Success(t *testing.T) {
	h, svc := newTestHandler()
	ctx := context.Background()

	booking := &domain.Booking{
		ID:          "booking-1",
		UserID:      "user-1",
		EventID:     "event-1",
		TicketCount: 2,
		Status:      domain.BookingStatusPending,
		CreatedAt:   time.Now(),
	}
	svc.On("CreateBooking", ctx, "user-1", "event-1", int32(2)).Return(booking, nil)

	resp, err := h.CreateBooking(ctx, &pb.CreateBookingRequest{
		UserId:      "user-1",
		EventId:     "event-1",
		TicketCount: 2,
	})

	assert.NoError(t, err)
	assert.Equal(t, "booking-1", resp.Booking.Id)
	assert.Equal(t, "user-1", resp.Booking.UserId)
	assert.Equal(t, int32(2), resp.Booking.TicketCount)
}

func TestCreateBooking_InvalidInput(t *testing.T) {
	h, svc := newTestHandler()
	ctx := context.Background()

	svc.On("CreateBooking", ctx, "", "event-1", int32(2)).Return(nil, domain.ErrInvalidInput)

	resp, err := h.CreateBooking(ctx, &pb.CreateBookingRequest{
		UserId:      "",
		EventId:     "event-1",
		TicketCount: 2,
	})

	assert.Nil(t, resp)
	st, ok := status.FromError(err)
	assert.True(t, ok)
	assert.Equal(t, codes.InvalidArgument, st.Code())
}

func TestCreateBooking_EventNotFound(t *testing.T) {
	h, svc := newTestHandler()
	ctx := context.Background()

	svc.On("CreateBooking", ctx, "user-1", "event-1", int32(2)).Return(nil, domain.ErrEventNotFound)

	resp, err := h.CreateBooking(ctx, &pb.CreateBookingRequest{
		UserId:      "user-1",
		EventId:     "event-1",
		TicketCount: 2,
	})

	assert.Nil(t, resp)
	st, ok := status.FromError(err)
	assert.True(t, ok)
	assert.Equal(t, codes.NotFound, st.Code())
}

func TestCreateBooking_InsufficientSeats(t *testing.T) {
	h, svc := newTestHandler()
	ctx := context.Background()

	svc.On("CreateBooking", ctx, "user-1", "event-1", int32(100)).Return(nil, domain.ErrInsufficientSeats)

	resp, err := h.CreateBooking(ctx, &pb.CreateBookingRequest{
		UserId:      "user-1",
		EventId:     "event-1",
		TicketCount: 100,
	})

	assert.Nil(t, resp)
	st, ok := status.FromError(err)
	assert.True(t, ok)
	assert.Equal(t, codes.FailedPrecondition, st.Code())
}

func TestGetBooking_Success(t *testing.T) {
	h, svc := newTestHandler()
	ctx := context.Background()

	booking := &domain.Booking{
		ID:          "booking-1",
		UserID:      "user-1",
		EventID:     "event-1",
		TicketCount: 2,
		Status:      domain.BookingStatusPending,
		CreatedAt:   time.Now(),
	}
	svc.On("GetBooking", ctx, "booking-1").Return(booking, nil)

	resp, err := h.GetBooking(ctx, &pb.GetBookingRequest{BookingId: "booking-1"})

	assert.NoError(t, err)
	assert.Equal(t, "booking-1", resp.Booking.Id)
}

func TestGetBooking_NotFound(t *testing.T) {
	h, svc := newTestHandler()
	ctx := context.Background()

	svc.On("GetBooking", ctx, "booking-1").Return(nil, domain.ErrBookingNotFound)

	resp, err := h.GetBooking(ctx, &pb.GetBookingRequest{BookingId: "booking-1"})

	assert.Nil(t, resp)
	st, ok := status.FromError(err)
	assert.True(t, ok)
	assert.Equal(t, codes.NotFound, st.Code())
}

func TestListUserBookings_Success(t *testing.T) {
	h, svc := newTestHandler()
	ctx := context.Background()

	bookings := []*domain.Booking{
		{ID: "b-1", UserID: "user-1", EventID: "event-1", TicketCount: 2, CreatedAt: time.Now()},
		{ID: "b-2", UserID: "user-1", EventID: "event-2", TicketCount: 1, CreatedAt: time.Now()},
	}
	svc.On("ListUserBookings", ctx, "user-1").Return(bookings, nil)

	resp, err := h.ListUserBookings(ctx, &pb.ListUserBookingsRequest{UserId: "user-1"})

	assert.NoError(t, err)
	assert.Len(t, resp.Bookings, 2)
}

func TestListUserBookings_InvalidInput(t *testing.T) {
	h, svc := newTestHandler()
	ctx := context.Background()

	svc.On("ListUserBookings", ctx, "").Return(nil, domain.ErrInvalidInput)

	resp, err := h.ListUserBookings(ctx, &pb.ListUserBookingsRequest{UserId: ""})

	assert.Nil(t, resp)
	st, ok := status.FromError(err)
	assert.True(t, ok)
	assert.Equal(t, codes.InvalidArgument, st.Code())
}

func TestCancelBooking_Success(t *testing.T) {
	h, svc := newTestHandler()
	ctx := context.Background()

	svc.On("CancelBooking", ctx, "booking-1").Return(nil)

	resp, err := h.CancelBooking(ctx, &pb.CancelBookingRequest{BookingId: "booking-1"})

	assert.NoError(t, err)
	assert.True(t, resp.Success)
	assert.Equal(t, "booking cancelled successfully", resp.Message)
}

func TestCancelBooking_NotFound(t *testing.T) {
	h, svc := newTestHandler()
	ctx := context.Background()

	svc.On("CancelBooking", ctx, "booking-1").Return(domain.ErrBookingNotFound)

	resp, err := h.CancelBooking(ctx, &pb.CancelBookingRequest{BookingId: "booking-1"})

	assert.NoError(t, err)
	assert.False(t, resp.Success)
	assert.Equal(t, "booking not found", resp.Message)
}

func TestCancelBooking_AlreadyCancelled(t *testing.T) {
	h, svc := newTestHandler()
	ctx := context.Background()

	svc.On("CancelBooking", ctx, "booking-1").Return(domain.ErrAlreadyCancelled)

	resp, err := h.CancelBooking(ctx, &pb.CancelBookingRequest{BookingId: "booking-1"})

	assert.NoError(t, err)
	assert.False(t, resp.Success)
	assert.Equal(t, "booking already cancelled", resp.Message)
}
