package domain

import (
	"errors"
	"time"
)

const (
	StatusPending   = "pending"
	StatusSucceeded = "succeeded"
	StatusFailed    = "failed"
	ProviderMock    = "mock"
)

var ErrIdempotencyConflict = errors.New("idempotency key was already used for a different payment request")

type Payment struct {
	ID             string    `json:"paymentId"`
	BookingID      string    `json:"bookingId"`
	UserID         string    `json:"-"`
	Provider       string    `json:"provider"`
	Status         string    `json:"status"`
	Amount         float64   `json:"amount"`
	IdempotencyKey *string   `json:"idempotencyKey,omitempty"`
	CreatedAt      time.Time `json:"createdAt"`
	UpdatedAt      time.Time `json:"updatedAt"`
}

func ValidForceStatus(status string) bool {
	return status == "" || status == StatusSucceeded || status == StatusFailed
}
