package application

import (
	"context"
	"errors"
	"net/http"
	"time"

	"github.com/rs/zerolog"

	apperrors "reserveflow/backend/internal/infrastructure/errors"
	"reserveflow/backend/internal/infrastructure/observability"
	rediscache "reserveflow/backend/internal/infrastructure/redis"
	"reserveflow/backend/internal/modules/bookings/domain"
)

type Repository interface {
	HoldSeat(ctx context.Context, userID, sessionID, seatID string, ttl time.Duration) (*HoldResult, error)
	GetBooking(ctx context.Context, bookingID string) (*domain.Booking, error)
	GetMyBookings(ctx context.Context, userID string) ([]domain.Booking, error)
	CancelBooking(ctx context.Context, userID, bookingID string) (*SeatChange, error)
	ExpirePending(ctx context.Context, limit int) ([]SeatChange, error)
	ConfirmBookingAfterPaymentSuccess(ctx context.Context, bookingID string) (*SeatChange, error)
	MarkBookingPaymentFailed(ctx context.Context, bookingID string) (*SeatChange, error)
}

type Cache interface {
	Set(ctx context.Context, key string, value any, ttl time.Duration) error
	Del(ctx context.Context, keys ...string) error
}

type Service struct {
	repo    Repository
	cache   Cache
	holdTTL time.Duration
	log     zerolog.Logger
}

type HoldResult struct {
	BookingID  string              `json:"bookingId"`
	Status     string              `json:"status"`
	ExpiresAt  time.Time           `json:"expiresAt"`
	Seat       domain.SeatSnapshot `json:"seat"`
	TotalPrice float64             `json:"totalPrice"`
	SessionID  string              `json:"-"`
	UserID     string              `json:"-"`
}

type SeatChange struct {
	BookingID string
	UserID    string
	SessionID string
	SeatID    string
}

func NewService(repo Repository, cache Cache, holdTTL time.Duration, log zerolog.Logger) *Service {
	return &Service{repo: repo, cache: cache, holdTTL: holdTTL, log: log}
}

func (s *Service) HoldSeat(ctx context.Context, userID, sessionID, seatID string) (*HoldResult, error) {
	if userID == "" || sessionID == "" || seatID == "" {
		return nil, apperrors.Validation("sessionId and seatId are required")
	}
	result, err := s.repo.HoldSeat(ctx, userID, sessionID, seatID, s.holdTTL)
	if err != nil {
		observability.BookingHoldFailedTotal.Inc()
		return nil, mapBookingError(err)
	}
	observability.BookingHoldTotal.Inc()
	s.afterSeatChange(ctx, SeatChange{BookingID: result.BookingID, UserID: userID, SessionID: sessionID, SeatID: seatID}, true)
	return result, nil
}

func (s *Service) GetBooking(ctx context.Context, userID, bookingID string) (*domain.Booking, error) {
	booking, err := s.repo.GetBooking(ctx, bookingID)
	if err != nil {
		return nil, mapBookingError(err)
	}
	if booking.UserID != userID {
		return nil, apperrors.Forbidden("Booking belongs to another user")
	}
	return booking, nil
}

func (s *Service) GetMyBookings(ctx context.Context, userID string) ([]domain.Booking, error) {
	bookings, err := s.repo.GetMyBookings(ctx, userID)
	if err != nil {
		return nil, apperrors.Internal(err)
	}
	return bookings, nil
}

func (s *Service) CancelBooking(ctx context.Context, userID, bookingID string) error {
	change, err := s.repo.CancelBooking(ctx, userID, bookingID)
	if err != nil {
		return mapBookingError(err)
	}
	s.afterSeatChange(ctx, *change, false)
	return nil
}

func (s *Service) ExpireBooking(ctx context.Context, limit int) ([]SeatChange, error) {
	changes, err := s.repo.ExpirePending(ctx, limit)
	if err != nil {
		return nil, err
	}
	for _, change := range changes {
		observability.BookingExpiredTotal.Inc()
		s.afterSeatChange(ctx, change, false)
	}
	return changes, nil
}

func (s *Service) ConfirmBookingAfterPaymentSuccess(ctx context.Context, bookingID string) error {
	change, err := s.repo.ConfirmBookingAfterPaymentSuccess(ctx, bookingID)
	if err != nil {
		return mapBookingError(err)
	}
	observability.BookingConfirmedTotal.Inc()
	s.afterSeatChange(ctx, *change, false)
	return nil
}

func (s *Service) MarkBookingPaymentFailed(ctx context.Context, bookingID string) error {
	change, err := s.repo.MarkBookingPaymentFailed(ctx, bookingID)
	if err != nil {
		return mapBookingError(err)
	}
	s.afterSeatChange(ctx, *change, false)
	return nil
}

func (s *Service) afterSeatChange(ctx context.Context, change SeatChange, createHoldKey bool) {
	if s.cache == nil {
		return
	}
	if createHoldKey {
		if err := s.cache.Set(ctx, rediscache.HoldKey(change.SessionID, change.SeatID), change.UserID, s.holdTTL); err != nil {
			s.log.Warn().Err(err).Str("session_id", change.SessionID).Str("seat_id", change.SeatID).Msg("failed to write redis hold key")
		}
	} else {
		if err := s.cache.Del(ctx, rediscache.HoldKey(change.SessionID, change.SeatID)); err != nil {
			s.log.Warn().Err(err).Str("session_id", change.SessionID).Str("seat_id", change.SeatID).Msg("failed to delete redis hold key")
		}
	}
	if err := s.cache.Del(ctx, rediscache.SeatmapKey(change.SessionID)); err != nil {
		s.log.Warn().Err(err).Str("session_id", change.SessionID).Msg("failed to invalidate seat map cache")
	}
}

func mapBookingError(err error) error {
	switch {
	case errors.Is(err, domain.ErrSeatNotFound):
		return apperrors.New(apperrors.CodeNotFound, "Session seat not found", http.StatusNotFound)
	case errors.Is(err, domain.ErrSeatNotAvailable):
		return apperrors.Conflict(apperrors.CodeSeatNotAvailable, "Seat is not available")
	case errors.Is(err, domain.ErrBookingNotFound):
		return apperrors.New(apperrors.CodeBookingNotFound, "Booking not found", http.StatusNotFound)
	case errors.Is(err, domain.ErrBookingNotOwner):
		return apperrors.Forbidden("Booking belongs to another user")
	case errors.Is(err, domain.ErrBookingExpired):
		return apperrors.Conflict(apperrors.CodeBookingExpired, "Booking expired")
	case errors.Is(err, domain.ErrBookingNotPending):
		return apperrors.Conflict(apperrors.CodeBookingNotPending, "Booking is not pending")
	default:
		return apperrors.Internal(err)
	}
}
