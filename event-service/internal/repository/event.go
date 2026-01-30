package repository

import (
	"context"
	"database/sql"
	"time"

	"github.com/google/uuid"
)

type Event struct {
	ID             string
	Name           string
	StartTime      time.Time
	TotalSeats     int32
	AvailableSeats int32
	CreatedAt      time.Time
}

type EventRepository interface {
	Create(ctx context.Context, event *Event) error
	GetByID(ctx context.Context, id string) (*Event, error)
	List(ctx context.Context, limit, offset int32) ([]*Event, int32, error)
	UpdateAvailableSeats(ctx context.Context, id string, quantity int32) (int32, error)
}

type PostgresEventRepository struct {
	db *sql.DB
}

func NewPostgresEventRepository(db *sql.DB) *PostgresEventRepository {
	return &PostgresEventRepository{db: db}
}

func (r *PostgresEventRepository) Create(ctx context.Context, event *Event) error {
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

func (r *PostgresEventRepository) GetByID(ctx context.Context, id string) (*Event, error) {
	query := `
		SELECT id, name, start_time, total_seats, available_seats, created_at
		FROM events
		WHERE id = $1
	`

	event := &Event{}
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

func (r *PostgresEventRepository) List(ctx context.Context, limit, offset int32) ([]*Event, int32, error) {
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

	var events []*Event
	for rows.Next() {
		event := &Event{}
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

func (r *PostgresEventRepository) UpdateAvailableSeats(ctx context.Context, id string, quantity int32) (int32, error) {
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
