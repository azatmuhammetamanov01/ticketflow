package usecase

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/azatmuhammetamanov01/online-ticket-booking/event-service/internal/domain"
	"github.com/azatmuhammetamanov01/online-ticket-booking/event-service/internal/domain/mocks"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestCreateEvent_Success(t *testing.T) {
	repo := new(mocks.MockEventRepository)
	uc := NewEventUsecase(repo)

	repo.On("Create", mock.Anything, mock.AnythingOfType("*domain.Event")).Return(nil)

	event, err := uc.CreateEvent(context.Background(), "Concert", time.Now().Add(24*time.Hour), 100)

	assert.NoError(t, err)
	assert.NotNil(t, event)
	assert.Equal(t, "Concert", event.Name)
	assert.Equal(t, int32(100), event.TotalSeats)
	repo.AssertExpectations(t)
}

func TestCreateEvent_EmptyName(t *testing.T) {
	repo := new(mocks.MockEventRepository)
	uc := NewEventUsecase(repo)

	event, err := uc.CreateEvent(context.Background(), "", time.Now().Add(24*time.Hour), 100)

	assert.ErrorIs(t, err, domain.ErrInvalidInput)
	assert.Nil(t, event)
}

func TestCreateEvent_ZeroSeats(t *testing.T) {
	repo := new(mocks.MockEventRepository)
	uc := NewEventUsecase(repo)

	event, err := uc.CreateEvent(context.Background(), "Concert", time.Now().Add(24*time.Hour), 0)

	assert.ErrorIs(t, err, domain.ErrInvalidInput)
	assert.Nil(t, event)
}

func TestCreateEvent_ZeroTime(t *testing.T) {
	repo := new(mocks.MockEventRepository)
	uc := NewEventUsecase(repo)

	event, err := uc.CreateEvent(context.Background(), "Concert", time.Time{}, 100)

	assert.ErrorIs(t, err, domain.ErrInvalidInput)
	assert.Nil(t, event)
}

func TestCreateEvent_RepoError(t *testing.T) {
	repo := new(mocks.MockEventRepository)
	uc := NewEventUsecase(repo)

	repo.On("Create", mock.Anything, mock.AnythingOfType("*domain.Event")).Return(errors.New("db error"))

	event, err := uc.CreateEvent(context.Background(), "Concert", time.Now().Add(24*time.Hour), 100)

	assert.Error(t, err)
	assert.Nil(t, event)
	repo.AssertExpectations(t)
}

func TestGetEvent_Success(t *testing.T) {
	repo := new(mocks.MockEventRepository)
	uc := NewEventUsecase(repo)

	expected := &domain.Event{
		ID:             "event-1",
		Name:           "Concert",
		TotalSeats:     100,
		AvailableSeats: 50,
	}
	repo.On("GetByID", mock.Anything, "event-1").Return(expected, nil)

	event, err := uc.GetEvent(context.Background(), "event-1")

	assert.NoError(t, err)
	assert.Equal(t, expected, event)
	repo.AssertExpectations(t)
}

func TestGetEvent_EmptyID(t *testing.T) {
	repo := new(mocks.MockEventRepository)
	uc := NewEventUsecase(repo)

	event, err := uc.GetEvent(context.Background(), "")

	assert.ErrorIs(t, err, domain.ErrInvalidInput)
	assert.Nil(t, event)
}

func TestGetEvent_NotFound(t *testing.T) {
	repo := new(mocks.MockEventRepository)
	uc := NewEventUsecase(repo)

	repo.On("GetByID", mock.Anything, "nonexistent").Return(nil, nil)

	event, err := uc.GetEvent(context.Background(), "nonexistent")

	assert.ErrorIs(t, err, domain.ErrEventNotFound)
	assert.Nil(t, event)
	repo.AssertExpectations(t)
}

func TestListEvents_Success(t *testing.T) {
	repo := new(mocks.MockEventRepository)
	uc := NewEventUsecase(repo)

	expected := []*domain.Event{
		{ID: "1", Name: "Concert"},
		{ID: "2", Name: "Theater"},
	}
	repo.On("List", mock.Anything, int32(10), int32(0)).Return(expected, int32(2), nil)

	events, total, err := uc.ListEvents(context.Background(), 10, 0)

	assert.NoError(t, err)
	assert.Equal(t, int32(2), total)
	assert.Len(t, events, 2)
	repo.AssertExpectations(t)
}

func TestUpdateAvailableTickets_Success(t *testing.T) {
	repo := new(mocks.MockEventRepository)
	uc := NewEventUsecase(repo)

	existing := &domain.Event{
		ID:             "event-1",
		AvailableSeats: 50,
	}
	repo.On("GetByID", mock.Anything, "event-1").Return(existing, nil)
	repo.On("UpdateAvailableSeats", mock.Anything, "event-1", int32(2)).Return(int32(48), nil)

	newAvailable, err := uc.UpdateAvailableTickets(context.Background(), "event-1", 2)

	assert.NoError(t, err)
	assert.Equal(t, int32(48), newAvailable)
	repo.AssertExpectations(t)
}

func TestUpdateAvailableTickets_EmptyID(t *testing.T) {
	repo := new(mocks.MockEventRepository)
	uc := NewEventUsecase(repo)

	available, err := uc.UpdateAvailableTickets(context.Background(), "", 2)

	assert.ErrorIs(t, err, domain.ErrInvalidInput)
	assert.Equal(t, int32(0), available)
}

func TestUpdateAvailableTickets_ZeroQuantity(t *testing.T) {
	repo := new(mocks.MockEventRepository)
	uc := NewEventUsecase(repo)

	available, err := uc.UpdateAvailableTickets(context.Background(), "event-1", 0)

	assert.ErrorIs(t, err, domain.ErrInvalidInput)
	assert.Equal(t, int32(0), available)
}

func TestUpdateAvailableTickets_InsufficientSeats(t *testing.T) {
	repo := new(mocks.MockEventRepository)
	uc := NewEventUsecase(repo)

	existing := &domain.Event{
		ID:             "event-1",
		AvailableSeats: 5,
	}
	repo.On("GetByID", mock.Anything, "event-1").Return(existing, nil)

	available, err := uc.UpdateAvailableTickets(context.Background(), "event-1", 10)

	assert.ErrorIs(t, err, domain.ErrInsufficientSeats)
	assert.Equal(t, int32(0), available)
	repo.AssertExpectations(t)
}

func TestUpdateAvailableTickets_EventNotFound(t *testing.T) {
	repo := new(mocks.MockEventRepository)
	uc := NewEventUsecase(repo)

	repo.On("GetByID", mock.Anything, "nonexistent").Return(nil, nil)

	available, err := uc.UpdateAvailableTickets(context.Background(), "nonexistent", 2)

	assert.ErrorIs(t, err, domain.ErrEventNotFound)
	assert.Equal(t, int32(0), available)
	repo.AssertExpectations(t)
}

func TestUpdateAvailableTickets_NegativeQuantity_AddsSeats(t *testing.T) {
	repo := new(mocks.MockEventRepository)
	uc := NewEventUsecase(repo)

	existing := &domain.Event{
		ID:             "event-1",
		AvailableSeats: 50,
	}
	repo.On("GetByID", mock.Anything, "event-1").Return(existing, nil)
	repo.On("UpdateAvailableSeats", mock.Anything, "event-1", int32(-2)).Return(int32(52), nil)

	newAvailable, err := uc.UpdateAvailableTickets(context.Background(), "event-1", -2)

	assert.NoError(t, err)
	assert.Equal(t, int32(52), newAvailable)
	repo.AssertExpectations(t)
}
