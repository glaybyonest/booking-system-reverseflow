package domain

import (
	"errors"
	"time"
)

const (
	StatusPending       = "pending"
	StatusConfirmed     = "confirmed"
	StatusCancelled     = "cancelled"
	StatusExpired       = "expired"
	StatusPaymentFailed = "payment_failed"

	ItemStatusHeld          = "held"
	ItemStatusBooked        = "booked"
	ItemStatusCancelled     = "cancelled"
	ItemStatusExpired       = "expired"
	ItemStatusPaymentFailed = "payment_failed"
)

var (
	ErrSeatNotFound      = errors.New("seat not found for session")
	ErrSeatNotAvailable  = errors.New("seat is not available")
	ErrBookingNotFound   = errors.New("booking not found")
	ErrBookingNotOwner   = errors.New("booking does not belong to user")
	ErrBookingExpired    = errors.New("booking expired")
	ErrBookingNotPending = errors.New("booking is not pending")
)

type Booking struct {
	ID         string        `json:"id"`
	UserID     string        `json:"userId"`
	SessionID  string        `json:"sessionId"`
	Status     string        `json:"status"`
	ExpiresAt  *time.Time    `json:"expiresAt,omitempty"`
	TotalPrice float64       `json:"totalPrice"`
	CreatedAt  time.Time     `json:"createdAt"`
	UpdatedAt  time.Time     `json:"updatedAt"`
	Items      []BookingItem `json:"items"`
}

type BookingItem struct {
	ID     string  `json:"id"`
	SeatID string  `json:"seatId"`
	Row    string  `json:"row"`
	Number int     `json:"number"`
	Price  float64 `json:"price"`
	Status string  `json:"status"`
}

type SeatSnapshot struct {
	SeatID string `json:"seatId"`
	Row    string `json:"row"`
	Number int    `json:"number"`
}

func CanConfirm(status string, expiresAt *time.Time, now time.Time) bool {
	return status == StatusPending && expiresAt != nil && expiresAt.After(now)
}

func CanCancel(status string) bool {
	return status == StatusPending
}

func CanExpire(status string, expiresAt *time.Time, now time.Time) bool {
	return status == StatusPending && expiresAt != nil && expiresAt.Before(now)
}

func HoldExpiresAt(now time.Time, ttl time.Duration) time.Time {
	return now.UTC().Add(ttl)
}
