package app

import (
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func newMockApp(t *testing.T) (*App, sqlmock.Sqlmock, func()) {
	t.Helper()

	db, mock, err := sqlmock.New(sqlmock.MonitorPingsOption(true))
	require.NoError(t, err)

	app := &App{db: db}

	cleanup := func() {
		require.NoError(t, db.Close())
		require.NoError(t, mock.ExpectationsWereMet())
	}

	return app, mock, cleanup
}

func TestHealthCheck_OK(t *testing.T) {
	app, mock, cleanup := newMockApp(t)
	defer cleanup()

	mock.ExpectPing()
	mock.ExpectClose()

	recorder := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/health-check", nil)

	app.healthCheck(recorder, req)

	assert.Equal(t, http.StatusOK, recorder.Code)
	assert.Equal(t, "application/json", recorder.Header().Get("Content-Type"))

	var body map[string]string
	require.NoError(t, json.NewDecoder(recorder.Body).Decode(&body))
	assert.Equal(t, "ok", body["status"])
}

func TestHealthCheck_DBError(t *testing.T) {
	app, mock, cleanup := newMockApp(t)
	defer cleanup()

	mock.ExpectPing().WillReturnError(errors.New("db down"))
	mock.ExpectClose()

	recorder := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/health-check", nil)

	app.healthCheck(recorder, req)

	assert.Equal(t, http.StatusServiceUnavailable, recorder.Code)
	assert.Equal(t, "application/json", recorder.Header().Get("Content-Type"))

	var body map[string]string
	require.NoError(t, json.NewDecoder(recorder.Body).Decode(&body))
	assert.Equal(t, "unhealthy", body["status"])
}
