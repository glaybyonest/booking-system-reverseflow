package application

import (
	"context"
	"errors"
	"time"

	"github.com/rs/zerolog"

	apperrors "reserveflow/backend/internal/infrastructure/errors"
	"reserveflow/backend/internal/infrastructure/observability"
	rediscache "reserveflow/backend/internal/infrastructure/redis"
	bookingdomain "reserveflow/backend/internal/modules/bookings/domain"
	paymentdomain "reserveflow/backend/internal/modules/payments/domain"
)

type Repository interface {
	ProcessMockPayment(ctx context.Context, userID, bookingID, idempotencyKey, forceStatus string) (*PaymentResult, error)
	GetPayment(ctx context.Context, paymentID string) (*paymentdomain.Payment, error)
	GetPaymentByIdempotencyKey(ctx context.Context, key string) (*paymentdomain.Payment, error)
}

type Cache interface {
	Get(ctx context.Context, key string) (string, error)
	Set(ctx context.Context, key string, value any, ttl time.Duration) error
	Del(ctx context.Context, keys ...string) error
}

type PaymentResult struct {
	Payment   paymentdomain.Payment
	SessionID string
	SeatIDs   []string
}

type Service struct {
	repo  Repository
	cache Cache
	log   zerolog.Logger
}

func NewService(repo Repository, cache Cache, log zerolog.Logger) *Service {
	return &Service{repo: repo, cache: cache, log: log}
}

func (s *Service) Process(ctx context.Context, userID, bookingID, idempotencyKey, forceStatus string) (*paymentdomain.Payment, error) {
	if userID == "" || bookingID == "" {
		return nil, apperrors.Validation("bookingId is required")
	}
	if !paymentdomain.ValidForceStatus(forceStatus) {
		return nil, apperrors.Validation("forceStatus must be succeeded or failed")
	}
	if idempotencyKey != "" {
		if payment := s.paymentFromCache(ctx, idempotencyKey); payment != nil {
			if err := validateIdempotentReplay(*payment, userID, bookingID, forceStatus); err != nil {
				return nil, err
			}
			return payment, nil
		}
		if payment, err := s.repo.GetPaymentByIdempotencyKey(ctx, idempotencyKey); err == nil {
			if err := validateIdempotentReplay(*payment, userID, bookingID, forceStatus); err != nil {
				return nil, err
			}
			return payment, nil
		}
	}
	result, err := s.repo.ProcessMockPayment(ctx, userID, bookingID, idempotencyKey, forceStatus)
	if err != nil {
		return nil, mapPaymentBookingError(err)
	}
	if idempotencyKey != "" {
		if err := validateIdempotentReplay(result.Payment, userID, bookingID, forceStatus); err != nil {
			return nil, err
		}
	}
	if result.Payment.Status == paymentdomain.StatusSucceeded {
		observability.PaymentSucceededTotal.Inc()
		observability.BookingConfirmedTotal.Inc()
	} else {
		observability.PaymentFailedTotal.Inc()
	}
	s.afterPayment(ctx, result, idempotencyKey)
	return &result.Payment, nil
}

func (s *Service) GetPayment(ctx context.Context, userID, paymentID string) (*paymentdomain.Payment, error) {
	payment, err := s.repo.GetPayment(ctx, paymentID)
	if err != nil {
		return nil, apperrors.New(apperrors.CodeNotFound, "Payment not found", 404)
	}
	if payment.UserID != userID {
		return nil, apperrors.Forbidden("Payment belongs to another user")
	}
	return payment, nil
}

func (s *Service) paymentFromCache(ctx context.Context, idempotencyKey string) *paymentdomain.Payment {
	if s.cache == nil {
		return nil
	}
	paymentID, err := s.cache.Get(ctx, rediscache.PaymentIdempotencyKey(idempotencyKey))
	if err != nil || paymentID == "" {
		return nil
	}
	payment, err := s.repo.GetPayment(ctx, paymentID)
	if err != nil {
		return nil
	}
	return payment
}

func (s *Service) afterPayment(ctx context.Context, result *PaymentResult, idempotencyKey string) {
	if s.cache == nil {
		return
	}
	if idempotencyKey != "" {
		if err := s.cache.Set(ctx, rediscache.PaymentIdempotencyKey(idempotencyKey), result.Payment.ID, 24*time.Hour); err != nil {
			s.log.Warn().Err(err).Str("payment_id", result.Payment.ID).Msg("failed to write payment idempotency cache")
		}
	}
	if result.SessionID != "" {
		for _, seatID := range result.SeatIDs {
			if err := s.cache.Del(ctx, rediscache.HoldKey(result.SessionID, seatID)); err != nil {
				s.log.Warn().Err(err).Str("session_id", result.SessionID).Str("seat_id", seatID).Msg("failed to delete hold key after payment")
			}
		}
	}
	if result.SessionID != "" {
		if err := s.cache.Del(ctx, rediscache.SeatmapKey(result.SessionID)); err != nil {
			s.log.Warn().Err(err).Str("session_id", result.SessionID).Msg("failed to invalidate seatmap after payment")
		}
	}
}

func mapPaymentBookingError(err error) error {
	switch {
	case errors.Is(err, paymentdomain.ErrIdempotencyConflict):
		return apperrors.Conflict(apperrors.CodeIdempotencyConflict, "Idempotency key was already used for a different payment request")
	case errors.Is(err, bookingdomain.ErrBookingNotFound):
		return apperrors.New(apperrors.CodeBookingNotFound, "Booking not found", 404)
	case errors.Is(err, bookingdomain.ErrBookingNotOwner):
		return apperrors.Forbidden("Booking belongs to another user")
	case errors.Is(err, bookingdomain.ErrBookingExpired):
		return apperrors.Conflict(apperrors.CodeBookingExpired, "Booking expired")
	case errors.Is(err, bookingdomain.ErrBookingNotPending):
		return apperrors.Conflict(apperrors.CodeBookingNotPending, "Booking is not pending")
	default:
		return apperrors.Internal(err)
	}
}

func validateIdempotentReplay(payment paymentdomain.Payment, userID, bookingID, forceStatus string) error {
	if payment.UserID != userID || payment.BookingID != bookingID {
		return apperrors.Conflict(apperrors.CodeIdempotencyConflict, "Idempotency key was already used for a different payment request")
	}
	if forceStatus != "" && payment.Status != forceStatus {
		return apperrors.Conflict(apperrors.CodeIdempotencyConflict, "Idempotency key was already used with a different forceStatus")
	}
	return nil
}
