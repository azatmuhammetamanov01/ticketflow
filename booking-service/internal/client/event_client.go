package client

import (
	"context"
	"errors"

	eventpb "github.com/azatmuhammetamanov01/online-ticket-booking/booking-service/event"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/status"
)

var (
	ErrEventNotFound     = errors.New("event not found")
	ErrInsufficientSeats = errors.New("insufficient seats available")
	ErrEventService      = errors.New("event service error")
)

type EventClient interface {
	GetEvent(ctx context.Context, eventID string) (*eventpb.Event, error)
	ReserveTickets(ctx context.Context, eventID string, quantity int32) error
	ReleaseTickets(ctx context.Context, eventID string, quantity int32) error
	Close() error
}

type eventClient struct {
	conn   *grpc.ClientConn
	client eventpb.EventServiceClient
}

// NewEventClient creates a gRPC client to event-service
func NewEventClient(eventServiceAddr string) (EventClient, error) {
	conn, err := grpc.NewClient(
		eventServiceAddr,
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	)
	if err != nil {
		return nil, err
	}

	return &eventClient{
		conn:   conn,
		client: eventpb.NewEventServiceClient(conn),
	}, nil
}

func (c *eventClient) GetEvent(ctx context.Context, eventID string) (*eventpb.Event, error) {
	resp, err := c.client.GetEvent(ctx, &eventpb.GetEventRequest{
		EventId: eventID,
	})
	if err != nil {
		if st, ok := status.FromError(err); ok {
			if st.Code() == codes.NotFound {
				return nil, ErrEventNotFound
			}
		}
		return nil, ErrEventService
	}

	return resp.Event, nil
}

func (c *eventClient) ReserveTickets(ctx context.Context, eventID string, quantity int32) error {
	resp, err := c.client.UpdateAvailableTickets(ctx, &eventpb.UpdateTicketsRequest{
		EventId:  eventID,
		Quantity: quantity,
	})
	if err != nil {
		if st, ok := status.FromError(err); ok {
			switch st.Code() {
			case codes.NotFound:
				return ErrEventNotFound
			case codes.FailedPrecondition:
				return ErrInsufficientSeats
			}
		}
		return ErrEventService
	}

	if !resp.Success {
		return ErrInsufficientSeats
	}

	return nil
}

func (c *eventClient) ReleaseTickets(ctx context.Context, eventID string, quantity int32) error {
	_, err := c.client.UpdateAvailableTickets(ctx, &eventpb.UpdateTicketsRequest{
		EventId:  eventID,
		Quantity: -quantity,
	})
	if err != nil {
		if st, ok := status.FromError(err); ok {
			if st.Code() == codes.NotFound {
				return ErrEventNotFound
			}
		}
		return ErrEventService
	}
	return nil
}

func (c *eventClient) Close() error {
	return c.conn.Close()
}
