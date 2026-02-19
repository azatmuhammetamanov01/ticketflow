package postgres

import (
	"context"
	"database/sql"
	"time"

	"github.com/azatmuhammetamanov01/online-ticket-booking/booking-service/internal/domain"
	"github.com/google/uuid"
)

type BookingRepository struct {
	db *sql.DB
}

func NewBookingRepository(db *sql.DB) *BookingRepository {
	return &BookingRepository{db: db}
}

func (r *BookingRepository) Create(ctx context.Context, booking *domain.Booking) error {
	booking.ID = uuid.New().String()
	booking.CreatedAt = time.Now()
	booking.Status = domain.BookingStatusPending

	query := `
		INSERT INTO bookings (id, user_id, event_id, ticket_count, status, created_at)
		VALUES ($1, $2, $3, $4, $5, $6)
	`

	_, err := r.db.ExecContext(ctx, query,
		booking.ID,
		booking.UserID,
		booking.EventID,
		booking.TicketCount,
		booking.Status,
		booking.CreatedAt,
	)

	return err
}

func (r *BookingRepository) GetByID(ctx context.Context, id string) (*domain.Booking, error) {
	query := `
		SELECT id, user_id, event_id, ticket_count, status, created_at
		FROM bookings
		WHERE id = $1
	`

	booking := &domain.Booking{}
	err := r.db.QueryRowContext(ctx, query, id).Scan(
		&booking.ID,
		&booking.UserID,
		&booking.EventID,
		&booking.TicketCount,
		&booking.Status,
		&booking.CreatedAt,
	)

	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}

	return booking, nil
}

func (r *BookingRepository) ListByUserID(ctx context.Context, userID string) ([]*domain.Booking, error) {
	query := `
		SELECT id, user_id, event_id, ticket_count, status, created_at
		FROM bookings
		WHERE user_id = $1
		ORDER BY created_at DESC
	`

	rows, err := r.db.QueryContext(ctx, query, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var bookings []*domain.Booking
	for rows.Next() {
		booking := &domain.Booking{}
		err := rows.Scan(
			&booking.ID,
			&booking.UserID,
			&booking.EventID,
			&booking.TicketCount,
			&booking.Status,
			&booking.CreatedAt,
		)
		if err != nil {
			return nil, err
		}
		bookings = append(bookings, booking)
	}

	return bookings, rows.Err()
}

func (r *BookingRepository) UpdateStatus(ctx context.Context, id string, status domain.BookingStatus) error {
	query := `UPDATE bookings SET status = $1 WHERE id = $2`
	result, err := r.db.ExecContext(ctx, query, status, id)
	if err != nil {
		return err
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return err
	}

	if rowsAffected == 0 {
		return sql.ErrNoRows
	}

	return nil
}
