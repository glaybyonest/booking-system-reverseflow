package domain

import (
	"errors"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func TestBookingStatusTransitions(t *testing.T) {
	now := time.Now().UTC()
	future := now.Add(10 * time.Minute)
	past := now.Add(-time.Minute)

	require.True(t, CanConfirm(StatusPending, &future, now))
	require.False(t, CanConfirm(StatusPending, &past, now))
	require.False(t, CanConfirm(StatusConfirmed, &future, now))

	require.True(t, CanCancel(StatusPending))
	require.False(t, CanCancel(StatusConfirmed))

	require.True(t, CanExpire(StatusPending, &past, now))
	require.False(t, CanExpire(StatusPending, &future, now))
	require.False(t, CanExpire(StatusCancelled, &past, now))
}

func TestHoldExpiresAt(t *testing.T) {
	now := time.Date(2026, 5, 4, 12, 0, 0, 0, time.UTC)
	require.Equal(t, now.Add(10*time.Minute), HoldExpiresAt(now, 10*time.Minute))
}

func TestDomainErrors(t *testing.T) {
	require.True(t, errors.Is(ErrSeatNotAvailable, ErrSeatNotAvailable))
	require.True(t, errors.Is(ErrBookingExpired, ErrBookingExpired))
	require.True(t, errors.Is(ErrBookingNotPending, ErrBookingNotPending))
}
