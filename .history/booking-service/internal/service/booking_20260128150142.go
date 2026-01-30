package service

import (
	"context"
	"errors"

	"github.com/azatmuhammetamanov01/online-ticket-booking/booking-service/internal/client"
	"github.com/azatmuhammetamanov01/online-ticket-booking/booking-service/internal/repository"
)

var (
	ErrBookingNotFound   = errors.New("booking not found")
	ErrInvalidInput      = errors.New("invalid input")
	ErrEventNotFound     = errors.New("event not found")
	ErrInsufficientSeats = errors.New("insufficient seats available")
)

type BookingService interface {
	CreateBooking(ctx context.Context, userID, eventID string, ticketCount int32) (*repository.Booking, error)
	GetBooking(ctx context.Context, bookingID string) (*repository.Booking, error)
	ListUserBookings(ctx context.Context, userID string) ([]*repository.Booking, error)
	CancelBooking(ctx context.Context, bookingID string) error
}

type bookingService struct {
	repo        repository.BookingRepository
	eventClient client.EventClient
}

func NewBookingService(repo repository.BookingRepository, eventClient client.EventClient) BookingService {
	return &bookingService{
		repo:        repo,
		eventClient: eventClient,
	}
}

func (s *bookingService) CreateBooking(ctx context.Context, userID, eventID string, ticketCount int32) (*repository.Booking, error) {
	if userID == "" || eventID == "" {
		return nil, ErrInvalidInput
	}
	if ticketCount <= 0 {
		return nil, ErrInvalidInput
	}

	event, err := s.eventClient.GetEvent(ctx, eventID)
	if err != nil {
		if errors.Is(err, client.ErrEventNotFound) {
			return nil, ErrEventNotFound
		}
		return nil, err
	}

	// 2. Yeterli koltuk var mı kontrol et
	if event.AvailableSeats < ticketCount {
		return nil, ErrInsufficientSeats
	}

	// 3. Koltukları rezerve et (event-service'te available_seats düşür)
	if err := s.eventClient.ReserveTickets(ctx, eventID, ticketCount); err != nil {
		if errors.Is(err, client.ErrInsufficientSeats) {
			return nil, ErrInsufficientSeats
		}
		return nil, err
	}

	// 4. Booking oluştur
	booking := &repository.Booking{
		UserID:      userID,
		EventID:     eventID,
		TicketCount: ticketCount,
	}

	if err := s.repo.Create(ctx, booking); err != nil {
		// TODO: Rollback - koltukları geri ver (saga pattern)
		return nil, err
	}

	return booking, nil
}

func (s *bookingService) GetBooking(ctx context.Context, bookingID string) (*repository.Booking, error) {
	if bookingID == "" {
		return nil, ErrInvalidInput
	}

	booking, err := s.repo.GetByID(ctx, bookingID)
	if err != nil {
		return nil, err
	}
	if booking == nil {
		return nil, ErrBookingNotFound
	}

	return booking, nil
}

func (s *bookingService) ListUserBookings(ctx context.Context, userID string) ([]*repository.Booking, error) {
	if userID == "" {
		return nil, ErrInvalidInput
	}

	return s.repo.ListByUserID(ctx, userID)
}

func (s *bookingService) CancelBooking(ctx context.Context, bookingID string) error {
	if bookingID == "" {
		return ErrInvalidInput
	}

	booking, err := s.repo.GetByID(ctx, bookingID)
	if err != nil {
		return err
	}
	if booking == nil {
		return ErrBookingNotFound
	}

	return s.repo.UpdateStatus(ctx, bookingID, repository.BookingStatusCancelled)
}
