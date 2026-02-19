package grpc

import (
	"context"
	"testing"
	"time"

	"github.com/azatmuhammetamanov01/online-ticket-booking/event-service/internal/domain"
	"github.com/azatmuhammetamanov01/online-ticket-booking/event-service/internal/domain/mocks"
	pb "github.com/azatmuhammetamanov01/online-ticket-booking/event-service/proto"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/timestamppb"
)

func TestCreateEvent_Success(t *testing.T) {
	svc := new(mocks.MockEventService)
	handler := NewEventHandler(svc)

	startTime := time.Now().Add(24 * time.Hour)
	expected := &domain.Event{
		ID:         "event-1",
		Name:       "Concert",
		StartTime:  startTime,
		TotalSeats: 100,
	}
	svc.On("CreateEvent", mock.Anything, "Concert", mock.AnythingOfType("time.Time"), int32(100)).Return(expected, nil)

	resp, err := handler.CreateEvent(context.Background(), &pb.CreateEventRequest{
		Name:       "Concert",
		StartTime:  timestamppb.New(startTime),
		TotalSeats: 100,
	})

	assert.NoError(t, err)
	assert.Equal(t, "event-1", resp.EventId)
	svc.AssertExpectations(t)
}

func TestCreateEvent_NilStartTime(t *testing.T) {
	svc := new(mocks.MockEventService)
	handler := NewEventHandler(svc)

	_, err := handler.CreateEvent(context.Background(), &pb.CreateEventRequest{
		Name:       "Concert",
		StartTime:  nil,
		TotalSeats: 100,
	})

	st, ok := status.FromError(err)
	assert.True(t, ok)
	assert.Equal(t, codes.InvalidArgument, st.Code())
}

func TestCreateEvent_InvalidInput(t *testing.T) {
	svc := new(mocks.MockEventService)
	handler := NewEventHandler(svc)

	svc.On("CreateEvent", mock.Anything, "", mock.AnythingOfType("time.Time"), int32(100)).Return(nil, domain.ErrInvalidInput)

	startTime := time.Now().Add(24 * time.Hour)
	_, err := handler.CreateEvent(context.Background(), &pb.CreateEventRequest{
		Name:       "",
		StartTime:  timestamppb.New(startTime),
		TotalSeats: 100,
	})

	st, ok := status.FromError(err)
	assert.True(t, ok)
	assert.Equal(t, codes.InvalidArgument, st.Code())
	svc.AssertExpectations(t)
}

func TestGetEvent_Success(t *testing.T) {
	svc := new(mocks.MockEventService)
	handler := NewEventHandler(svc)

	now := time.Now()
	expected := &domain.Event{
		ID:             "event-1",
		Name:           "Concert",
		StartTime:      now,
		TotalSeats:     100,
		AvailableSeats: 50,
		CreatedAt:      now,
	}
	svc.On("GetEvent", mock.Anything, "event-1").Return(expected, nil)

	resp, err := handler.GetEvent(context.Background(), &pb.GetEventRequest{
		EventId: "event-1",
	})

	assert.NoError(t, err)
	assert.Equal(t, "event-1", resp.Event.Id)
	assert.Equal(t, "Concert", resp.Event.Name)
	assert.Equal(t, int32(100), resp.Event.TotalSeats)
	assert.Equal(t, int32(50), resp.Event.AvailableSeats)
	svc.AssertExpectations(t)
}

func TestGetEvent_NotFound(t *testing.T) {
	svc := new(mocks.MockEventService)
	handler := NewEventHandler(svc)

	svc.On("GetEvent", mock.Anything, "nonexistent").Return(nil, domain.ErrEventNotFound)

	_, err := handler.GetEvent(context.Background(), &pb.GetEventRequest{
		EventId: "nonexistent",
	})

	st, ok := status.FromError(err)
	assert.True(t, ok)
	assert.Equal(t, codes.NotFound, st.Code())
	svc.AssertExpectations(t)
}

func TestListEvents_Success(t *testing.T) {
	svc := new(mocks.MockEventService)
	handler := NewEventHandler(svc)

	now := time.Now()
	events := []*domain.Event{
		{ID: "1", Name: "Concert", StartTime: now, CreatedAt: now},
		{ID: "2", Name: "Theater", StartTime: now, CreatedAt: now},
	}
	svc.On("ListEvents", mock.Anything, int32(10), int32(0)).Return(events, int32(2), nil)

	resp, err := handler.ListEvents(context.Background(), &pb.ListEventsRequest{
		Limit:  10,
		Offset: 0,
	})

	assert.NoError(t, err)
	assert.Len(t, resp.Events, 2)
	assert.Equal(t, int32(2), resp.TotalCount)
	svc.AssertExpectations(t)
}

func TestUpdateAvailableTickets_Success(t *testing.T) {
	svc := new(mocks.MockEventService)
	handler := NewEventHandler(svc)

	svc.On("UpdateAvailableTickets", mock.Anything, "event-1", int32(2)).Return(int32(48), nil)

	resp, err := handler.UpdateAvailableTickets(context.Background(), &pb.UpdateTicketsRequest{
		EventId:  "event-1",
		Quantity: 2,
	})

	assert.NoError(t, err)
	assert.True(t, resp.Success)
	assert.Equal(t, int32(48), resp.AvailableSeats)
	svc.AssertExpectations(t)
}

func TestUpdateAvailableTickets_NotFound(t *testing.T) {
	svc := new(mocks.MockEventService)
	handler := NewEventHandler(svc)

	svc.On("UpdateAvailableTickets", mock.Anything, "nonexistent", int32(2)).Return(int32(0), domain.ErrEventNotFound)

	resp, err := handler.UpdateAvailableTickets(context.Background(), &pb.UpdateTicketsRequest{
		EventId:  "nonexistent",
		Quantity: 2,
	})

	assert.NoError(t, err)
	assert.False(t, resp.Success)
	assert.Equal(t, int32(0), resp.AvailableSeats)
	svc.AssertExpectations(t)
}

func TestUpdateAvailableTickets_InsufficientSeats(t *testing.T) {
	svc := new(mocks.MockEventService)
	handler := NewEventHandler(svc)

	svc.On("UpdateAvailableTickets", mock.Anything, "event-1", int32(100)).Return(int32(0), domain.ErrInsufficientSeats)

	resp, err := handler.UpdateAvailableTickets(context.Background(), &pb.UpdateTicketsRequest{
		EventId:  "event-1",
		Quantity: 100,
	})

	assert.NoError(t, err)
	assert.False(t, resp.Success)
	svc.AssertExpectations(t)
}
