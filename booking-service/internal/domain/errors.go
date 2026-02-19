package domain

import "errors"

var (
	ErrBookingNotFound   = errors.New("booking not found")
	ErrInvalidInput      = errors.New("invalid input")
	ErrEventNotFound     = errors.New("event not found")
	ErrInsufficientSeats = errors.New("insufficient seats available")
	ErrAlreadyCancelled  = errors.New("booking already cancelled")
)
