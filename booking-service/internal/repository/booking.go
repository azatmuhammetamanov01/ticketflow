package repository

import (
	"context"
	"database/sql"
	"time"

	"github.com/google/uuid"
)

type BookingStatus int32

const (
	BookingStatusUnspecified BookingStatus = 0
	BookingStatusPending     BookingStatus = 1
	BookingStatusConfirmed   BookingStatus = 2
	BookingStatusCancelled   BookingStatus = 3
)

type Booking struct {
	ID          string
	UserID      string
	EventID     string
	TicketCount int32
	Status      BookingStatus
	CreatedAt   time.Time
}

type BookingRepository interface {
	Create(ctx context.Context, booking *Booking) error
	GetByID(ctx context.Context, id string) (*Booking, error)
	ListByUserID(ctx context.Context, userID string) ([]*Booking, error)
	UpdateStatus(ctx context.Context, id string, status BookingStatus) error
}

type PostgresBookingRepository struct {
	db *sql.DB
}

func NewPostgresBookingRepository(db *sql.DB) *PostgresBookingRepository {
	return &PostgresBookingRepository{db: db}
}

func (r *PostgresBookingRepository) Create(ctx context.Context, booking *Booking) error {
	booking.ID = uuid.New().String()
	booking.CreatedAt = time.Now()
	booking.Status = BookingStatusPending

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

func (r *PostgresBookingRepository) GetByID(ctx context.Context, id string) (*Booking, error) {
	query := `
		SELECT id, user_id, event_id, ticket_count, status, created_at
		FROM bookings
		WHERE id = $1
	`

	booking := &Booking{}
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

func (r *PostgresBookingRepository) ListByUserID(ctx context.Context, userID string) ([]*Booking, error) {
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

	var bookings []*Booking
	for rows.Next() {
		booking := &Booking{}
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

func (r *PostgresBookingRepository) UpdateStatus(ctx context.Context, id string, status BookingStatus) error {
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
