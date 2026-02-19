package grpc

import (
	"context"
	"errors"

	"github.com/azatmuhammetamanov01/online-ticket-booking/booking-service/internal/domain"
	pb "github.com/azatmuhammetamanov01/online-ticket-booking/booking-service/proto"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/timestamppb"
)

type BookingHandler struct {
	pb.UnimplementedBookingServiceServer
	svc domain.BookingService
}

func NewBookingHandler(svc domain.BookingService) *BookingHandler {
	return &BookingHandler{svc: svc}
}

func (h *BookingHandler) CreateBooking(ctx context.Context, req *pb.CreateBookingRequest) (*pb.CreateBookingResponse, error) {
	booking, err := h.svc.CreateBooking(ctx, req.UserId, req.EventId, req.TicketCount)
	if err != nil {
		if errors.Is(err, domain.ErrInvalidInput) {
			return nil, status.Error(codes.InvalidArgument, err.Error())
		}
		if errors.Is(err, domain.ErrEventNotFound) {
			return nil, status.Error(codes.NotFound, err.Error())
		}
		if errors.Is(err, domain.ErrInsufficientSeats) {
			return nil, status.Error(codes.FailedPrecondition, err.Error())
		}
		return nil, status.Error(codes.Internal, "failed to create booking")
	}

	return &pb.CreateBookingResponse{
		Booking: toProtoBooking(booking),
	}, nil
}

func (h *BookingHandler) GetBooking(ctx context.Context, req *pb.GetBookingRequest) (*pb.GetBookingResponse, error) {
	booking, err := h.svc.GetBooking(ctx, req.BookingId)
	if err != nil {
		if errors.Is(err, domain.ErrBookingNotFound) {
			return nil, status.Error(codes.NotFound, err.Error())
		}
		if errors.Is(err, domain.ErrInvalidInput) {
			return nil, status.Error(codes.InvalidArgument, err.Error())
		}
		return nil, status.Error(codes.Internal, "failed to get booking")
	}

	return &pb.GetBookingResponse{
		Booking: toProtoBooking(booking),
	}, nil
}

func (h *BookingHandler) ListUserBookings(ctx context.Context, req *pb.ListUserBookingsRequest) (*pb.ListUserBookingsResponse, error) {
	bookings, err := h.svc.ListUserBookings(ctx, req.UserId)
	if err != nil {
		if errors.Is(err, domain.ErrInvalidInput) {
			return nil, status.Error(codes.InvalidArgument, err.Error())
		}
		return nil, status.Error(codes.Internal, "failed to list bookings")
	}

	pbBookings := make([]*pb.Booking, len(bookings))
	for i, b := range bookings {
		pbBookings[i] = toProtoBooking(b)
	}

	return &pb.ListUserBookingsResponse{
		Bookings: pbBookings,
	}, nil
}

func (h *BookingHandler) CancelBooking(ctx context.Context, req *pb.CancelBookingRequest) (*pb.CancelBookingResponse, error) {
	err := h.svc.CancelBooking(ctx, req.BookingId)
	if err != nil {
		if errors.Is(err, domain.ErrBookingNotFound) {
			return &pb.CancelBookingResponse{
				Success: false,
				Message: "booking not found",
			}, nil
		}
		if errors.Is(err, domain.ErrAlreadyCancelled) {
			return &pb.CancelBookingResponse{
				Success: false,
				Message: "booking already cancelled",
			}, nil
		}
		if errors.Is(err, domain.ErrInvalidInput) {
			return nil, status.Error(codes.InvalidArgument, err.Error())
		}
		return nil, status.Error(codes.Internal, "failed to cancel booking")
	}

	return &pb.CancelBookingResponse{
		Success: true,
		Message: "booking cancelled successfully",
	}, nil
}

func toProtoBooking(b *domain.Booking) *pb.Booking {
	return &pb.Booking{
		Id:          b.ID,
		UserId:      b.UserID,
		EventId:     b.EventID,
		TicketCount: b.TicketCount,
		Status:      pb.BookingStatus(b.Status),
		CreatedAt:   timestamppb.New(b.CreatedAt),
	}
}
