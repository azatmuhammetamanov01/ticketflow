package handler

import (
	"context"
	"errors"

	"github.com/azatmuhammetamanov01/online-ticket-booking/event-service/internal/repository"
	"github.com/azatmuhammetamanov01/online-ticket-booking/event-service/internal/service"
	pb "github.com/azatmuhammetamanov01/online-ticket-booking/event-service/proto"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/timestamppb"
)

type EventHandler struct {
	pb.UnimplementedEventSeriveServer
	svc service.EventService
}

func NewEventHandler(svc service.EventService) *EventHandler {
	return &EventHandler{svc: svc}
}

func (h *EventHandler) CreateEvent(ctx context.Context, req *pb.CreateEventRequest) (*pb.CreateEventResponse, error) {
	if req.StartTime == nil {
		return nil, status.Error(codes.InvalidArgument, "start_time is required")
	}

	event, err := h.svc.CreateEvent(ctx, req.Name, req.StartTime.AsTime(), req.TotalSeats)
	if err != nil {
		if errors.Is(err, service.ErrInvalidInput) {
			return nil, status.Error(codes.InvalidArgument, err.Error())
		}
		return nil, status.Error(codes.Internal, "failed to create event")
	}

	return &pb.CreateEventResponse{
		EventId: event.ID,
	}, nil
}

func (h *EventHandler) GetEvent(ctx context.Context, req *pb.GetEventRequest) (*pb.GetEventResponse, error) {
	event, err := h.svc.GetEvent(ctx, req.EventId)
	if err != nil {
		if errors.Is(err, service.ErrEventNotFound) {
			return nil, status.Error(codes.NotFound, err.Error())
		}
		if errors.Is(err, service.ErrInvalidInput) {
			return nil, status.Error(codes.InvalidArgument, err.Error())
		}
		return nil, status.Error(codes.Internal, "failed to get event")
	}

	return &pb.GetEventResponse{
		Event: toProtoEvent(event),
	}, nil
}

func (h *EventHandler) ListEvents(ctx context.Context, req *pb.ListEventsRequest) (*pb.ListEventsResponse, error) {
	events, totalCount, err := h.svc.ListEvents(ctx, req.Limit, req.Offset)
	if err != nil {
		return nil, status.Error(codes.Internal, "failed to list events")
	}

	pbEvents := make([]*pb.Event, len(events))
	for i, e := range events {
		pbEvents[i] = toProtoEvent(e)
	}

	return &pb.ListEventsResponse{
		Events:     pbEvents,
		TotalCount: totalCount,
	}, nil
}

func (h *EventHandler) UpdateAvailableTickets(ctx context.Context, req *pb.UpdateTicketsRequest) (*pb.UpdateTicketsResponse, error) {
	newAvailable, err := h.svc.UpdateAvailableTickets(ctx, req.EventId, req.Quantity)
	if err != nil {
		if errors.Is(err, service.ErrEventNotFound) {
			return &pb.UpdateTicketsResponse{
				AvailableSeats: 0,
				Success:        false,
			}, nil
		}
		if errors.Is(err, service.ErrInsufficientSeats) {
			return &pb.UpdateTicketsResponse{
				AvailableSeats: 0,
				Success:        false,
			}, nil
		}
		if errors.Is(err, service.ErrInvalidInput) {
			return nil, status.Error(codes.InvalidArgument, err.Error())
		}
		return nil, status.Error(codes.Internal, "failed to update tickets")
	}

	return &pb.UpdateTicketsResponse{
		AvailableSeats: newAvailable,
		Success:        true,
	}, nil
}

func toProtoEvent(e *repository.Event) *pb.Event {
	return &pb.Event{
		Id:             e.ID,
		Name:           e.Name,
		StartTime:      timestamppb.New(e.StartTime),
		TotalSeats:     e.TotalSeats,
		AvailableSeats: e.AvailableSeats,
		CreatedAt:      timestamppb.New(e.CreatedAt),
	}
}
