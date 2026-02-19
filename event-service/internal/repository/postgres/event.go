package postgres

import (
	"context"
	"database/sql"
	"time"

	"github.com/azatmuhammetamanov01/online-ticket-booking/event-service/internal/domain"
	"github.com/google/uuid"
)

type EventRepository struct {
	db *sql.DB
}

func NewEventRepository(db *sql.DB) *EventRepository {
	return &EventRepository{db: db}
}

func (r *EventRepository) Create(ctx context.Context, event *domain.Event) error {
	event.ID = uuid.New().String()
	event.CreatedAt = time.Now()
	event.AvailableSeats = event.TotalSeats

	query := `
		INSERT INTO events (id, name, start_time, total_seats, available_seats, created_at)
		VALUES ($1, $2, $3, $4, $5, $6)
	`

	_, err := r.db.ExecContext(ctx, query,
		event.ID,
		event.Name,
		event.StartTime,
		event.TotalSeats,
		event.AvailableSeats,
		event.CreatedAt,
	)

	return err
}

func (r *EventRepository) GetByID(ctx context.Context, id string) (*domain.Event, error) {
	query := `
		SELECT id, name, start_time, total_seats, available_seats, created_at
		FROM events
		WHERE id = $1
	`

	event := &domain.Event{}
	err := r.db.QueryRowContext(ctx, query, id).Scan(
		&event.ID,
		&event.Name,
		&event.StartTime,
		&event.TotalSeats,
		&event.AvailableSeats,
		&event.CreatedAt,
	)

	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}

	return event, nil
}

func (r *EventRepository) List(ctx context.Context, limit, offset int32) ([]*domain.Event, int32, error) {
	if limit <= 0 {
		limit = 10
	}
	if limit > 100 {
		limit = 100
	}

	countQuery := `SELECT COUNT(*) FROM events`
	var totalCount int32
	if err := r.db.QueryRowContext(ctx, countQuery).Scan(&totalCount); err != nil {
		return nil, 0, err
	}

	query := `
		SELECT id, name, start_time, total_seats, available_seats, created_at
		FROM events
		ORDER BY start_time ASC
		LIMIT $1 OFFSET $2
	`

	rows, err := r.db.QueryContext(ctx, query, limit, offset)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	var events []*domain.Event
	for rows.Next() {
		event := &domain.Event{}
		err := rows.Scan(
			&event.ID,
			&event.Name,
			&event.StartTime,
			&event.TotalSeats,
			&event.AvailableSeats,
			&event.CreatedAt,
		)
		if err != nil {
			return nil, 0, err
		}
		events = append(events, event)
	}

	return events, totalCount, rows.Err()
}

func (r *EventRepository) UpdateAvailableSeats(ctx context.Context, id string, quantity int32) (int32, error) {
	query := `
		UPDATE events
		SET available_seats = available_seats - $1
		WHERE id = $2 AND available_seats >= $1
		RETURNING available_seats
	`

	var newAvailable int32
	err := r.db.QueryRowContext(ctx, query, quantity, id).Scan(&newAvailable)
	if err == sql.ErrNoRows {
		return 0, nil
	}
	if err != nil {
		return 0, err
	}

	return newAvailable, nil
}
