package domain

import "errors"

var (
	ErrEventNotFound     = errors.New("event not found")
	ErrInvalidInput      = errors.New("invalid input")
	ErrInsufficientSeats = errors.New("insufficient available seats")
)
