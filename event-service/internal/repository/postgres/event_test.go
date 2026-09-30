package postgres

import (
	"context"
	"database/sql"
	"errors"
	"regexp"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/azatmuhammetamanov01/online-ticket-booking/event-service/internal/domain"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func newMockRepository(t *testing.T) (*EventRepository, sqlmock.Sqlmock, func()) {
	t.Helper()

	db, mock, err := sqlmock.New()
	require.NoError(t, err)

	repo := NewEventRepository(db)

	cleanup := func() {
		mock.ExpectClose()
		require.NoError(t, db.Close())
		require.NoError(t, mock.ExpectationsWereMet())
	}

	return repo, mock, cleanup
}

func TestCreate_Success(t *testing.T) {
	repo, mock, cleanup := newMockRepository(t)
	defer cleanup()

	event := &domain.Event{
		Name:       "Concert",
		StartTime:  time.Date(2026, 6, 20, 18, 0, 0, 0, time.UTC),
		TotalSeats: 100,
	}

	query := `
		INSERT INTO events (id, name, start_time, total_seats, available_seats, created_at)
		VALUES ($1, $2, $3, $4, $5, $6)
	`

	mock.ExpectExec(regexp.QuoteMeta(query)).
		WithArgs(sqlmock.AnyArg(), "Concert", event.StartTime, int32(100), int32(100), sqlmock.AnyArg()).
		WillReturnResult(sqlmock.NewResult(1, 1))

	err := repo.Create(context.Background(), event)

	require.NoError(t, err)
	assert.NotEmpty(t, event.ID)
	assert.False(t, event.CreatedAt.IsZero())
	assert.Equal(t, int32(100), event.AvailableSeats)
}

func TestCreate_DBError(t *testing.T) {
	repo, mock, cleanup := newMockRepository(t)
	defer cleanup()

	event := &domain.Event{
		Name:       "Concert",
		StartTime:  time.Date(2026, 6, 20, 18, 0, 0, 0, time.UTC),
		TotalSeats: 100,
	}

	query := `
		INSERT INTO events (id, name, start_time, total_seats, available_seats, created_at)
		VALUES ($1, $2, $3, $4, $5, $6)
	`

	mock.ExpectExec(regexp.QuoteMeta(query)).
		WithArgs(sqlmock.AnyArg(), "Concert", event.StartTime, int32(100), int32(100), sqlmock.AnyArg()).
		WillReturnError(errors.New("insert failed"))

	err := repo.Create(context.Background(), event)

	assert.Error(t, err)
}

func TestGetByID_Success(t *testing.T) {
	repo, mock, cleanup := newMockRepository(t)
	defer cleanup()

	now := time.Date(2026, 6, 19, 12, 0, 0, 0, time.UTC)
	query := `
		SELECT id, name, start_time, total_seats, available_seats, created_at
		FROM events
		WHERE id = $1
	`

	rows := sqlmock.NewRows([]string{"id", "name", "start_time", "total_seats", "available_seats", "created_at"}).
		AddRow("event-1", "Concert", now, int32(100), int32(50), now)

	mock.ExpectQuery(regexp.QuoteMeta(query)).WithArgs("event-1").WillReturnRows(rows)

	event, err := repo.GetByID(context.Background(), "event-1")

	require.NoError(t, err)
	require.NotNil(t, event)
	assert.Equal(t, "event-1", event.ID)
	assert.Equal(t, "Concert", event.Name)
	assert.Equal(t, now, event.StartTime)
	assert.Equal(t, int32(100), event.TotalSeats)
	assert.Equal(t, int32(50), event.AvailableSeats)
	assert.Equal(t, now, event.CreatedAt)
}

func TestGetByID_NotFound(t *testing.T) {
	repo, mock, cleanup := newMockRepository(t)
	defer cleanup()

	query := `
		SELECT id, name, start_time, total_seats, available_seats, created_at
		FROM events
		WHERE id = $1
	`

	mock.ExpectQuery(regexp.QuoteMeta(query)).WithArgs("missing").WillReturnError(sql.ErrNoRows)

	event, err := repo.GetByID(context.Background(), "missing")

	require.NoError(t, err)
	assert.Nil(t, event)
}

func TestGetByID_DBError(t *testing.T) {
	repo, mock, cleanup := newMockRepository(t)
	defer cleanup()

	query := `
		SELECT id, name, start_time, total_seats, available_seats, created_at
		FROM events
		WHERE id = $1
	`

	mock.ExpectQuery(regexp.QuoteMeta(query)).WithArgs("event-1").WillReturnError(errors.New("query failed"))

	event, err := repo.GetByID(context.Background(), "event-1")

	assert.Error(t, err)
	assert.Nil(t, event)
}

func TestList_DefaultLimitAndSuccess(t *testing.T) {
	repo, mock, cleanup := newMockRepository(t)
	defer cleanup()

	countQuery := `SELECT COUNT(*) FROM events`
	listQuery := `
		SELECT id, name, start_time, total_seats, available_seats, created_at
		FROM events
		ORDER BY start_time ASC
		LIMIT $1 OFFSET $2
	`

	countRows := sqlmock.NewRows([]string{"count"}).AddRow(int32(2))
	eventRows := sqlmock.NewRows([]string{"id", "name", "start_time", "total_seats", "available_seats", "created_at"}).
		AddRow("event-1", "Concert", time.Date(2026, 6, 19, 12, 0, 0, 0, time.UTC), int32(100), int32(50), time.Date(2026, 6, 19, 11, 0, 0, 0, time.UTC)).
		AddRow("event-2", "Theater", time.Date(2026, 6, 20, 12, 0, 0, 0, time.UTC), int32(80), int32(20), time.Date(2026, 6, 19, 11, 5, 0, 0, time.UTC))

	mock.ExpectQuery(regexp.QuoteMeta(countQuery)).WillReturnRows(countRows)
	mock.ExpectQuery(regexp.QuoteMeta(listQuery)).WithArgs(int32(10), int32(0)).WillReturnRows(eventRows)

	events, total, err := repo.List(context.Background(), 0, 0)

	require.NoError(t, err)
	assert.Equal(t, int32(2), total)
	require.Len(t, events, 2)
	assert.Equal(t, "event-1", events[0].ID)
	assert.Equal(t, "event-2", events[1].ID)
}

func TestList_ClampsLimitToHundred(t *testing.T) {
	repo, mock, cleanup := newMockRepository(t)
	defer cleanup()

	countQuery := `SELECT COUNT(*) FROM events`
	listQuery := `
		SELECT id, name, start_time, total_seats, available_seats, created_at
		FROM events
		ORDER BY start_time ASC
		LIMIT $1 OFFSET $2
	`

	mock.ExpectQuery(regexp.QuoteMeta(countQuery)).WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(int32(0)))
	mock.ExpectQuery(regexp.QuoteMeta(listQuery)).WithArgs(int32(100), int32(25)).WillReturnRows(sqlmock.NewRows([]string{"id", "name", "start_time", "total_seats", "available_seats", "created_at"}))

	events, total, err := repo.List(context.Background(), 250, 25)

	require.NoError(t, err)
	assert.Empty(t, events)
	assert.Equal(t, int32(0), total)
}

func TestList_QueryError(t *testing.T) {
	repo, mock, cleanup := newMockRepository(t)
	defer cleanup()

	countQuery := `SELECT COUNT(*) FROM events`
	mock.ExpectQuery(regexp.QuoteMeta(countQuery)).WillReturnError(errors.New("count failed"))

	events, total, err := repo.List(context.Background(), 10, 0)

	assert.Error(t, err)
	assert.Nil(t, events)
	assert.Equal(t, int32(0), total)
}

func TestUpdateAvailableSeats_Success(t *testing.T) {
	repo, mock, cleanup := newMockRepository(t)
	defer cleanup()

	query := `
		UPDATE events
		SET available_seats = available_seats - $1
		WHERE id = $2 AND available_seats >= $1
		RETURNING available_seats
	`

	mock.ExpectQuery(regexp.QuoteMeta(query)).WithArgs(int32(2), "event-1").WillReturnRows(sqlmock.NewRows([]string{"available_seats"}).AddRow(int32(48)))

	newAvailable, err := repo.UpdateAvailableSeats(context.Background(), "event-1", 2)

	require.NoError(t, err)
	assert.Equal(t, int32(48), newAvailable)
}

func TestUpdateAvailableSeats_NotFound(t *testing.T) {
	repo, mock, cleanup := newMockRepository(t)
	defer cleanup()

	query := `
		UPDATE events
		SET available_seats = available_seats - $1
		WHERE id = $2 AND available_seats >= $1
		RETURNING available_seats
	`

	mock.ExpectQuery(regexp.QuoteMeta(query)).WithArgs(int32(2), "missing").WillReturnError(sql.ErrNoRows)

	newAvailable, err := repo.UpdateAvailableSeats(context.Background(), "missing", 2)

	require.NoError(t, err)
	assert.Equal(t, int32(0), newAvailable)
}

func TestUpdateAvailableSeats_DBError(t *testing.T) {
	repo, mock, cleanup := newMockRepository(t)
	defer cleanup()

	query := `
		UPDATE events
		SET available_seats = available_seats - $1
		WHERE id = $2 AND available_seats >= $1
		RETURNING available_seats
	`

	mock.ExpectQuery(regexp.QuoteMeta(query)).WithArgs(int32(2), "event-1").WillReturnError(errors.New("update failed"))

	newAvailable, err := repo.UpdateAvailableSeats(context.Background(), "event-1", 2)

	assert.Error(t, err)
	assert.Equal(t, int32(0), newAvailable)
}
