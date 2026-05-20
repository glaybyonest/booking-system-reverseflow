package repository

import (
	"context"
	"encoding/json"
	"errors"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"reserveflow/backend/internal/modules/bookings/application"
	"reserveflow/backend/internal/modules/bookings/domain"
)

type PostgresRepository struct {
	db *pgxpool.Pool
}

func NewPostgresRepository(db *pgxpool.Pool) *PostgresRepository {
	return &PostgresRepository{db: db}
}

func (r *PostgresRepository) HoldSeats(ctx context.Context, userID, sessionID string, seatIDs []string, ttl time.Duration) (*application.HoldResult, error) {
	tx, err := r.db.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		return nil, err
	}
	defer rollback(ctx, tx)

	var isBookable bool
	if err := tx.QueryRow(ctx, `
		SELECT is_bookable
		FROM sessions
		WHERE id = $1
		FOR SHARE
	`, sessionID).Scan(&isBookable); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, domain.ErrSeatNotFound
		}
		return nil, err
	}
	if !isBookable {
		return nil, domain.ErrSessionNotBookable
	}

	var currentStatus string
	rows, err := tx.Query(ctx, `
		SELECT seat.id, seat.row_label, seat.seat_number, seat.base_price, ss.status
		FROM session_seats ss
		JOIN seats seat ON seat.id = ss.seat_id
		WHERE ss.session_id = $1 AND ss.seat_id = ANY($2::uuid[]) AND ss.is_active = TRUE
		ORDER BY seat.row_label, seat.seat_number
		FOR UPDATE
	`, sessionID, seatIDs)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	type holdCandidate struct {
		snapshot domain.SeatSnapshot
		price    float64
	}

	candidates := make([]holdCandidate, 0, len(seatIDs))
	snapshots := make([]domain.SeatSnapshot, 0, len(seatIDs))
	seatIDsOrdered := make([]string, 0, len(seatIDs))
	totalPrice := 0.0
	for rows.Next() {
		var item holdCandidate
		if err := rows.Scan(&item.snapshot.SeatID, &item.snapshot.Row, &item.snapshot.Number, &item.price, &currentStatus); err != nil {
			return nil, err
		}
		if currentStatus != "available" {
			return nil, domain.ErrSeatNotAvailable
		}
		seatIDsOrdered = append(seatIDsOrdered, item.snapshot.SeatID)
		snapshots = append(snapshots, item.snapshot)
		candidates = append(candidates, item)
		totalPrice += item.price
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	if len(seatIDsOrdered) != len(seatIDs) {
		return nil, domain.ErrSeatNotFound
	}

	now := time.Now().UTC()
	expiresAt := domain.HoldExpiresAt(now, ttl)
	bookingID := uuid.NewString()

	if _, err := tx.Exec(ctx, `
		UPDATE session_seats
		SET status = 'held', hold_expires_at = $3, version = version + 1, updated_at = $4
		WHERE session_id = $1 AND seat_id = ANY($2::uuid[]) AND is_active = TRUE
	`, sessionID, seatIDsOrdered, expiresAt, now); err != nil {
		return nil, err
	}
	if _, err := tx.Exec(ctx, `
		INSERT INTO bookings (id, user_id, session_id, status, expires_at, total_price, created_at, updated_at)
		VALUES ($1, $2, $3, 'pending', $4, $5, $6, $6)
	`, bookingID, userID, sessionID, expiresAt, totalPrice, now); err != nil {
		return nil, err
	}

	for _, item := range candidates {
		if _, err := tx.Exec(ctx, `
			INSERT INTO booking_items (id, booking_id, seat_id, price, status, created_at, updated_at)
			VALUES ($1, $2, $3, $4, 'held', $5, $5)
		`, uuid.NewString(), bookingID, item.snapshot.SeatID, item.price, now); err != nil {
			return nil, err
		}
	}

	payload := map[string]any{"userId": userID, "bookingId": bookingID, "sessionId": sessionID, "seatIds": seatIDsOrdered, "expiresAt": expiresAt}
	if err := insertOutbox(ctx, tx, "seat.held", "seat", seatIDsOrdered[0], payload); err != nil {
		return nil, err
	}
	if err := insertOutbox(ctx, tx, "booking.created", "booking", bookingID, payload); err != nil {
		return nil, err
	}
	if err := tx.Commit(ctx); err != nil {
		return nil, err
	}
	return &application.HoldResult{
		BookingID:  bookingID,
		Status:     domain.StatusPending,
		ExpiresAt:  expiresAt,
		Seats:      snapshots,
		TotalPrice: totalPrice,
		SessionID:  sessionID,
		UserID:     userID,
		SeatIDs:    seatIDsOrdered,
	}, nil
}

func (r *PostgresRepository) GetBooking(ctx context.Context, bookingID string) (*domain.Booking, error) {
	var booking domain.Booking
	err := r.db.QueryRow(ctx, `
		SELECT id, user_id, session_id, status, expires_at, total_price, created_at, updated_at
		FROM bookings
		WHERE id = $1
	`, bookingID).Scan(&booking.ID, &booking.UserID, &booking.SessionID, &booking.Status, &booking.ExpiresAt, &booking.TotalPrice, &booking.CreatedAt, &booking.UpdatedAt)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, domain.ErrBookingNotFound
		}
		return nil, err
	}
	items, err := r.listItems(ctx, bookingID)
	if err != nil {
		return nil, err
	}
	booking.Items = items
	return &booking, nil
}

func (r *PostgresRepository) GetMyBookings(ctx context.Context, userID string) ([]domain.Booking, error) {
	rows, err := r.db.Query(ctx, `
		SELECT id, user_id, session_id, status, expires_at, total_price, created_at, updated_at
		FROM bookings
		WHERE user_id = $1
		ORDER BY created_at DESC
	`, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	bookings := make([]domain.Booking, 0)
	for rows.Next() {
		var booking domain.Booking
		if err := rows.Scan(&booking.ID, &booking.UserID, &booking.SessionID, &booking.Status, &booking.ExpiresAt, &booking.TotalPrice, &booking.CreatedAt, &booking.UpdatedAt); err != nil {
			return nil, err
		}
		items, err := r.listItems(ctx, booking.ID)
		if err != nil {
			return nil, err
		}
		booking.Items = items
		bookings = append(bookings, booking)
	}
	return bookings, rows.Err()
}

func (r *PostgresRepository) CancelBooking(ctx context.Context, userID, bookingID string) (*application.SeatChange, error) {
	return r.transitionPending(ctx, userID, bookingID, "cancelled")
}

func (r *PostgresRepository) ExpirePending(ctx context.Context, limit int) ([]application.SeatChange, error) {
	tx, err := r.db.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		return nil, err
	}
	defer rollback(ctx, tx)

	rows, err := tx.Query(ctx, `
		SELECT id, user_id, session_id
		FROM bookings
		WHERE status = 'pending' AND expires_at < now()
		ORDER BY expires_at
		LIMIT $1
		FOR UPDATE SKIP LOCKED
	`, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	changes := make([]application.SeatChange, 0)
	for rows.Next() {
		var change application.SeatChange
		if err := rows.Scan(&change.BookingID, &change.UserID, &change.SessionID); err != nil {
			return nil, err
		}
		changes = append(changes, change)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	rows.Close()

	for i := range changes {
		change := &changes[i]
		seatIDs, err := lockBookingItems(ctx, tx, change.BookingID)
		if err != nil {
			return nil, err
		}
		change.SeatIDs = seatIDs
		if err := lockSessionSeats(ctx, tx, change.SessionID, seatIDs); err != nil {
			return nil, err
		}
		now := time.Now().UTC()
		if _, err := tx.Exec(ctx, `
			UPDATE bookings SET status = 'expired', updated_at = $2 WHERE id = $1 AND status = 'pending'
		`, change.BookingID, now); err != nil {
			return nil, err
		}
		if _, err := tx.Exec(ctx, `
			UPDATE booking_items SET status = 'expired', updated_at = $2 WHERE booking_id = $1
		`, change.BookingID, now); err != nil {
			return nil, err
		}
		if _, err := tx.Exec(ctx, `
			UPDATE session_seats SET status = 'available', hold_expires_at = NULL, version = version + 1, updated_at = $3
			WHERE session_id = $1 AND seat_id = ANY($2::uuid[])
		`, change.SessionID, change.SeatIDs, now); err != nil {
			return nil, err
		}
		if err := insertOutbox(ctx, tx, "booking.expired", "booking", change.BookingID, map[string]any{"userId": change.UserID, "bookingId": change.BookingID, "sessionId": change.SessionID, "seatIds": change.SeatIDs}); err != nil {
			return nil, err
		}
	}
	if err := tx.Commit(ctx); err != nil {
		return nil, err
	}
	return changes, nil
}

func (r *PostgresRepository) ConfirmBookingAfterPaymentSuccess(ctx context.Context, bookingID string) (*application.SeatChange, error) {
	tx, err := r.db.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		return nil, err
	}
	defer rollback(ctx, tx)
	change, err := r.lockPendingBooking(ctx, tx, "", bookingID)
	if err != nil {
		return nil, err
	}
	now := time.Now().UTC()
	if _, err := tx.Exec(ctx, `UPDATE bookings SET status = 'confirmed', updated_at = $2 WHERE id = $1`, bookingID, now); err != nil {
		return nil, err
	}
	if _, err := tx.Exec(ctx, `UPDATE booking_items SET status = 'booked', updated_at = $2 WHERE booking_id = $1`, bookingID, now); err != nil {
		return nil, err
	}
	if _, err := tx.Exec(ctx, `
		UPDATE session_seats SET status = 'booked', hold_expires_at = NULL, version = version + 1, updated_at = $3
		WHERE session_id = $1 AND seat_id = ANY($2::uuid[])
	`, change.SessionID, change.SeatIDs, now); err != nil {
		return nil, err
	}
	if err := insertOutbox(ctx, tx, "booking.confirmed", "booking", bookingID, map[string]any{"userId": change.UserID, "bookingId": bookingID, "sessionId": change.SessionID, "seatIds": change.SeatIDs}); err != nil {
		return nil, err
	}
	if err := tx.Commit(ctx); err != nil {
		return nil, err
	}
	return change, nil
}

func (r *PostgresRepository) MarkBookingPaymentFailed(ctx context.Context, bookingID string) (*application.SeatChange, error) {
	tx, err := r.db.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		return nil, err
	}
	defer rollback(ctx, tx)
	change, err := r.lockPendingBooking(ctx, tx, "", bookingID)
	if err != nil {
		return nil, err
	}
	now := time.Now().UTC()
	if _, err := tx.Exec(ctx, `UPDATE bookings SET status = 'payment_failed', updated_at = $2 WHERE id = $1`, bookingID, now); err != nil {
		return nil, err
	}
	if _, err := tx.Exec(ctx, `UPDATE booking_items SET status = 'payment_failed', updated_at = $2 WHERE booking_id = $1`, bookingID, now); err != nil {
		return nil, err
	}
	if _, err := tx.Exec(ctx, `
		UPDATE session_seats SET status = 'available', hold_expires_at = NULL, version = version + 1, updated_at = $3
		WHERE session_id = $1 AND seat_id = ANY($2::uuid[])
	`, change.SessionID, change.SeatIDs, now); err != nil {
		return nil, err
	}
	if err := insertOutbox(ctx, tx, "payment.failed", "booking", bookingID, map[string]any{"userId": change.UserID, "bookingId": bookingID, "sessionId": change.SessionID, "seatIds": change.SeatIDs}); err != nil {
		return nil, err
	}
	if err := tx.Commit(ctx); err != nil {
		return nil, err
	}
	return change, nil
}

func (r *PostgresRepository) transitionPending(ctx context.Context, userID, bookingID, target string) (*application.SeatChange, error) {
	tx, err := r.db.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		return nil, err
	}
	defer rollback(ctx, tx)
	change, err := r.lockPendingBooking(ctx, tx, userID, bookingID)
	if err != nil {
		return nil, err
	}
	now := time.Now().UTC()
	itemStatus := target
	if _, err := tx.Exec(ctx, `UPDATE bookings SET status = $2, updated_at = $3 WHERE id = $1`, bookingID, target, now); err != nil {
		return nil, err
	}
	if _, err := tx.Exec(ctx, `UPDATE booking_items SET status = $2, updated_at = $3 WHERE booking_id = $1`, bookingID, itemStatus, now); err != nil {
		return nil, err
	}
	if _, err := tx.Exec(ctx, `
		UPDATE session_seats
		SET status = 'available', hold_expires_at = NULL, version = version + 1, updated_at = $3
		WHERE session_id = $1 AND seat_id = ANY($2::uuid[])
	`, change.SessionID, change.SeatIDs, now); err != nil {
		return nil, err
	}
	if err := insertOutbox(ctx, tx, "booking."+target, "booking", bookingID, map[string]any{"userId": change.UserID, "bookingId": bookingID, "sessionId": change.SessionID, "seatIds": change.SeatIDs}); err != nil {
		return nil, err
	}
	if err := tx.Commit(ctx); err != nil {
		return nil, err
	}
	return change, nil
}

func (r *PostgresRepository) lockPendingBooking(ctx context.Context, tx pgx.Tx, userID, bookingID string) (*application.SeatChange, error) {
	var change application.SeatChange
	var status string
	var expiresAt *time.Time
	err := tx.QueryRow(ctx, `
		SELECT id, user_id, session_id, status, expires_at
		FROM bookings
		WHERE id = $1
		FOR UPDATE
	`, bookingID).Scan(&change.BookingID, &change.UserID, &change.SessionID, &status, &expiresAt)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, domain.ErrBookingNotFound
		}
		return nil, err
	}
	if userID != "" && change.UserID != userID {
		return nil, domain.ErrBookingNotOwner
	}
	if status != domain.StatusPending {
		return nil, domain.ErrBookingNotPending
	}
	if expiresAt != nil && expiresAt.Before(time.Now().UTC()) {
		return nil, domain.ErrBookingExpired
	}
	seatIDs, err := lockBookingItems(ctx, tx, bookingID)
	if err != nil {
		return nil, err
	}
	change.SeatIDs = seatIDs
	if err := lockSessionSeats(ctx, tx, change.SessionID, change.SeatIDs); err != nil {
		return nil, err
	}
	return &change, nil
}

func (r *PostgresRepository) listItems(ctx context.Context, bookingID string) ([]domain.BookingItem, error) {
	rows, err := r.db.Query(ctx, `
		SELECT bi.id, bi.seat_id, s.row_label, s.seat_number, bi.price, bi.status
		FROM booking_items bi
		JOIN seats s ON s.id = bi.seat_id
		WHERE bi.booking_id = $1
		ORDER BY s.row_label, s.seat_number
	`, bookingID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	items := make([]domain.BookingItem, 0)
	for rows.Next() {
		var item domain.BookingItem
		if err := rows.Scan(&item.ID, &item.SeatID, &item.Row, &item.Number, &item.Price, &item.Status); err != nil {
			return nil, err
		}
		items = append(items, item)
	}
	return items, rows.Err()
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
	if err := rows.Err(); err != nil {
		return nil, err
	}
	if len(seatIDs) == 0 {
		return nil, domain.ErrSeatNotFound
	}
	return seatIDs, nil
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
		return domain.ErrSeatNotFound
	}
	return nil
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
