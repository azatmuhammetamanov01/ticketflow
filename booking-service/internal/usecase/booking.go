package usecase

import (
	"context"
	"errors"

	"github.com/azatmuhammetamanov01/online-ticket-booking/booking-service/internal/client"
	"github.com/azatmuhammetamanov01/online-ticket-booking/booking-service/internal/domain"
	"github.com/azatmuhammetamanov01/online-ticket-booking/booking-service/internal/logger"
	"go.uber.org/zap"
)

type BookingUsecase struct {
	repo        domain.BookingRepository
	eventClient client.EventClient
}

func NewBookingUsecase(repo domain.BookingRepository, eventClient client.EventClient) *BookingUsecase {
	return &BookingUsecase{
		repo:        repo,
		eventClient: eventClient,
	}
}

func (u *BookingUsecase) CreateBooking(ctx context.Context, userID, eventID string, ticketCount int32) (*domain.Booking, error) {
	if userID == "" || eventID == "" {
		return nil, domain.ErrInvalidInput
	}
	if ticketCount <= 0 {
		return nil, domain.ErrInvalidInput
	}

	event, err := u.eventClient.GetEvent(ctx, eventID)
	if err != nil {
		if errors.Is(err, client.ErrEventNotFound) {
			return nil, domain.ErrEventNotFound
		}
		return nil, err
	}

	if event.AvailableSeats < ticketCount {
		return nil, domain.ErrInsufficientSeats
	}

	if err := u.eventClient.ReserveTickets(ctx, eventID, ticketCount); err != nil {
		if errors.Is(err, client.ErrInsufficientSeats) {
			return nil, domain.ErrInsufficientSeats
		}
		return nil, err
	}

	booking := &domain.Booking{
		UserID:      userID,
		EventID:     eventID,
		TicketCount: ticketCount,
	}

	if err := u.repo.Create(ctx, booking); err != nil {
		return nil, err
	}

	return booking, nil
}

func (u *BookingUsecase) GetBooking(ctx context.Context, bookingID string) (*domain.Booking, error) {
	if bookingID == "" {
		return nil, domain.ErrInvalidInput
	}

	booking, err := u.repo.GetByID(ctx, bookingID)
	if err != nil {
		return nil, err
	}
	if booking == nil {
		return nil, domain.ErrBookingNotFound
	}

	return booking, nil
}

func (u *BookingUsecase) ListUserBookings(ctx context.Context, userID string) ([]*domain.Booking, error) {
	if userID == "" {
		return nil, domain.ErrInvalidInput
	}

	return u.repo.ListByUserID(ctx, userID)
}

func (u *BookingUsecase) CancelBooking(ctx context.Context, bookingID string) error {
	if bookingID == "" {
		return domain.ErrInvalidInput
	}

	booking, err := u.repo.GetByID(ctx, bookingID)
	if err != nil {
		return err
	}
	if booking == nil {
		return domain.ErrBookingNotFound
	}

	if booking.Status == domain.BookingStatusCancelled {
		return domain.ErrAlreadyCancelled
	}

	logger.Info("CancelBooking: releasing tickets",
		zap.Int32("ticketCount", booking.TicketCount),
		zap.String("eventID", booking.EventID),
	)
	if err := u.eventClient.ReleaseTickets(ctx, booking.EventID, booking.TicketCount); err != nil {
		logger.Error("CancelBooking: ReleaseTickets failed", zap.Error(err))
		return err
	}
	logger.Info("CancelBooking: tickets released successfully")

	return u.repo.UpdateStatus(ctx, bookingID, domain.BookingStatusCancelled)
}
