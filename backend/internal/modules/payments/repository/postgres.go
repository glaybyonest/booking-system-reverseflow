package repository

import (
	"context"
	"encoding/json"
	"errors"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"

	bookingdomain "reserveflow/backend/internal/modules/bookings/domain"
	"reserveflow/backend/internal/modules/payments/application"
	paymentdomain "reserveflow/backend/internal/modules/payments/domain"
)

type PostgresRepository struct {
	db *pgxpool.Pool
}

func NewPostgresRepository(db *pgxpool.Pool) *PostgresRepository {
	return &PostgresRepository{db: db}
}

func (r *PostgresRepository) ProcessMockPayment(ctx context.Context, userID, bookingID, idempotencyKey, forceStatus string) (*application.PaymentResult, error) {
	if idempotencyKey != "" {
		if existing, err := r.GetPaymentByIdempotencyKey(ctx, idempotencyKey); err == nil {
			return &application.PaymentResult{Payment: *existing}, nil
		}
	}
	tx, err := r.db.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		return nil, err
	}
	defer rollback(ctx, tx)

	var ownerID, sessionID, status string
	var expiresAt *time.Time
	var amount float64
	err = tx.QueryRow(ctx, `
		SELECT id, user_id, session_id, status, expires_at, total_price
		FROM bookings
		WHERE id = $1
		FOR UPDATE
	`, bookingID).Scan(&bookingID, &ownerID, &sessionID, &status, &expiresAt, &amount)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, bookingdomain.ErrBookingNotFound
		}
		return nil, err
	}
	if ownerID != userID {
		return nil, bookingdomain.ErrBookingNotOwner
	}
	if status != bookingdomain.StatusPending {
		return nil, bookingdomain.ErrBookingNotPending
	}
	if expiresAt == nil || !expiresAt.After(time.Now().UTC()) {
		return nil, bookingdomain.ErrBookingExpired
	}

	seatIDs, err := lockBookingItems(ctx, tx, bookingID)
	if err != nil {
		return nil, err
	}
	if err := lockSessionSeats(ctx, tx, sessionID, seatIDs); err != nil {
		return nil, err
	}

	resultStatus := paymentdomain.StatusSucceeded
	if forceStatus == paymentdomain.StatusFailed {
		resultStatus = paymentdomain.StatusFailed
	}
	paymentID := uuid.NewString()
	now := time.Now().UTC()
	var idem *string
	if idempotencyKey != "" {
		idem = &idempotencyKey
	}
	if _, err := tx.Exec(ctx, `
		INSERT INTO payments (id, booking_id, provider, status, amount, idempotency_key, created_at, updated_at)
		VALUES ($1, $2, 'mock', 'pending', $3, $4, $5, $5)
	`, paymentID, bookingID, amount, idem, now); err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == "23505" && idempotencyKey != "" {
			_ = tx.Rollback(ctx)
			existing, existingErr := r.GetPaymentByIdempotencyKey(ctx, idempotencyKey)
			if existingErr != nil {
				return nil, existingErr
			}
			return &application.PaymentResult{Payment: *existing}, nil
		}
		return nil, err
	}

	if resultStatus == paymentdomain.StatusSucceeded {
		if _, err := tx.Exec(ctx, `UPDATE payments SET status = 'succeeded', updated_at = $2 WHERE id = $1`, paymentID, now); err != nil {
			return nil, err
		}
		if _, err := tx.Exec(ctx, `UPDATE bookings SET status = 'confirmed', updated_at = $2 WHERE id = $1`, bookingID, now); err != nil {
			return nil, err
		}
		if _, err := tx.Exec(ctx, `UPDATE booking_items SET status = 'booked', updated_at = $2 WHERE booking_id = $1`, bookingID, now); err != nil {
			return nil, err
		}
		if _, err := tx.Exec(ctx, `
			UPDATE session_seats SET status = 'booked', hold_expires_at = NULL, version = version + 1, updated_at = $3
			WHERE session_id = $1 AND seat_id = ANY($2::uuid[])
		`, sessionID, seatIDs, now); err != nil {
			return nil, err
		}
		payload := map[string]any{"userId": userID, "bookingId": bookingID, "paymentId": paymentID, "sessionId": sessionID, "seatIds": seatIDs}
		if err := insertOutbox(ctx, tx, "payment.succeeded", "payment", paymentID, payload); err != nil {
			return nil, err
		}
		if err := insertOutbox(ctx, tx, "booking.confirmed", "booking", bookingID, payload); err != nil {
			return nil, err
		}
	} else {
		if _, err := tx.Exec(ctx, `UPDATE payments SET status = 'failed', updated_at = $2 WHERE id = $1`, paymentID, now); err != nil {
			return nil, err
		}
		if _, err := tx.Exec(ctx, `UPDATE bookings SET status = 'payment_failed', updated_at = $2 WHERE id = $1`, bookingID, now); err != nil {
			return nil, err
		}
		if _, err := tx.Exec(ctx, `UPDATE booking_items SET status = 'payment_failed', updated_at = $2 WHERE booking_id = $1`, bookingID, now); err != nil {
			return nil, err
		}
		if _, err := tx.Exec(ctx, `
			UPDATE session_seats SET status = 'available', hold_expires_at = NULL, version = version + 1, updated_at = $3
			WHERE session_id = $1 AND seat_id = ANY($2::uuid[])
		`, sessionID, seatIDs, now); err != nil {
			return nil, err
		}
		payload := map[string]any{"userId": userID, "bookingId": bookingID, "paymentId": paymentID, "sessionId": sessionID, "seatIds": seatIDs}
		if err := insertOutbox(ctx, tx, "payment.failed", "payment", paymentID, payload); err != nil {
			return nil, err
		}
	}
	if err := tx.Commit(ctx); err != nil {
		return nil, err
	}
	return &application.PaymentResult{
		Payment: paymentdomain.Payment{
			ID:             paymentID,
			BookingID:      bookingID,
			UserID:         userID,
			Provider:       paymentdomain.ProviderMock,
			Status:         resultStatus,
			Amount:         amount,
			IdempotencyKey: idem,
			CreatedAt:      now,
			UpdatedAt:      now,
		},
		SessionID: sessionID,
		SeatIDs:   seatIDs,
	}, nil
}

func (r *PostgresRepository) GetPayment(ctx context.Context, paymentID string) (*paymentdomain.Payment, error) {
	return r.scanPayment(r.db.QueryRow(ctx, `
		SELECT p.id, p.booking_id, b.user_id, p.provider, p.status, p.amount, p.idempotency_key, p.created_at, p.updated_at
		FROM payments p
		JOIN bookings b ON b.id = p.booking_id
		WHERE p.id = $1
	`, paymentID))
}

func (r *PostgresRepository) GetPaymentByIdempotencyKey(ctx context.Context, key string) (*paymentdomain.Payment, error) {
	return r.scanPayment(r.db.QueryRow(ctx, `
		SELECT p.id, p.booking_id, b.user_id, p.provider, p.status, p.amount, p.idempotency_key, p.created_at, p.updated_at
		FROM payments p
		JOIN bookings b ON b.id = p.booking_id
		WHERE p.idempotency_key = $1
	`, key))
}

type rowScanner interface {
	Scan(dest ...any) error
}

func (r *PostgresRepository) scanPayment(row rowScanner) (*paymentdomain.Payment, error) {
	var payment paymentdomain.Payment
	err := row.Scan(&payment.ID, &payment.BookingID, &payment.UserID, &payment.Provider, &payment.Status, &payment.Amount, &payment.IdempotencyKey, &payment.CreatedAt, &payment.UpdatedAt)
	if err != nil {
		return nil, err
	}
	return &payment, nil
}

func insertOutbox(ctx context.Context, tx pgx.Tx, eventType, aggregateType, aggregateID string, payload any) error {
	encoded, err := json.Marshal(payload)
	if err != nil {
		return err
	}
	_, err = tx.Exec(ctx, `
		INSERT INTO outbox_events (id, event_type, aggregate_type, aggregate_id, payload, status, created_at)
		VALUES ($1, $2, $3, $4, $5, 'pending', $6)
	`, uuid.NewString(), eventType, aggregateType, aggregateID, encoded, time.Now().UTC())
	return err
}

func rollback(ctx context.Context, tx pgx.Tx) {
	_ = tx.Rollback(ctx)
}

func lockBookingItems(ctx context.Context, tx pgx.Tx, bookingID string) ([]string, error) {
	rows, err := tx.Query(ctx, `
		SELECT seat_id
		FROM booking_items
		WHERE booking_id = $1
		ORDER BY seat_id
		FOR UPDATE
	`, bookingID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	seatIDs := make([]string, 0)
	for rows.Next() {
		var seatID string
		if err := rows.Scan(&seatID); err != nil {
			return nil, err
		}
		seatIDs = append(seatIDs, seatID)
	}
	return seatIDs, rows.Err()
}

func lockSessionSeats(ctx context.Context, tx pgx.Tx, sessionID string, seatIDs []string) error {
	rows, err := tx.Query(ctx, `
		SELECT id
		FROM session_seats
		WHERE session_id = $1 AND seat_id = ANY($2::uuid[])
		FOR UPDATE
	`, sessionID, seatIDs)
	if err != nil {
		return err
	}
	defer rows.Close()

	count := 0
	for rows.Next() {
		count++
	}
	if err := rows.Err(); err != nil {
		return err
	}
	if count != len(seatIDs) {
		return bookingdomain.ErrSeatNotFound
	}
	return nil
}
