package repository

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"

	eventsdomain "reserveflow/backend/internal/modules/events/domain"
	integrationsapp "reserveflow/backend/internal/modules/integrations/application"
	seatsdomain "reserveflow/backend/internal/modules/seats/domain"
)

type sessionLayoutRow struct {
	SessionID     string
	EventID       string
	EventTitle    string
	EventSource   string
	BookingMode   string
	IsBookable    bool
	HallID        sql.NullString
	HallName      sql.NullString
	VenueID       sql.NullString
	VenueName     sql.NullString
	SessionLayout []byte
	HallLayout    []byte
}

func (r *PostgresRepository) GetSessionLayoutState(ctx context.Context, sessionID string) (*integrationsapp.SessionLayoutState, error) {
	row, err := r.loadSessionLayoutRow(ctx, r.db, sessionID)
	if err != nil {
		return nil, err
	}
	return buildSessionLayoutState(row)
}

func (r *PostgresRepository) UpsertSessionLayout(ctx context.Context, sessionID string, layout seatsdomain.StoredSeatLayout, now time.Time) (*integrationsapp.LayoutMutationResult, error) {
	tx, err := r.db.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		return nil, err
	}
	defer rollback(ctx, tx)

	row, err := r.loadSessionLayoutRow(ctx, tx, sessionID)
	if err != nil {
		return nil, err
	}

	hallID, err := r.ensureSessionOverrideHall(ctx, tx, row, now)
	if err != nil {
		return nil, err
	}

	encoded, err := json.Marshal(layout)
	if err != nil {
		return nil, err
	}
	if _, err := tx.Exec(ctx, `
		UPDATE sessions
		SET hall_id = $2, layout_json = $3, updated_at = $4
		WHERE id = $1
	`, sessionID, hallID, encoded, now); err != nil {
		return nil, err
	}

	result, err := r.syncSessionEffectiveLayout(ctx, tx, sessionID, now)
	if err != nil {
		return nil, err
	}
	if err := tx.Commit(ctx); err != nil {
		return nil, err
	}
	return result, nil
}

func (r *PostgresRepository) DeleteSessionLayout(ctx context.Context, sessionID string, now time.Time) (*integrationsapp.LayoutMutationResult, error) {
	tx, err := r.db.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		return nil, err
	}
	defer rollback(ctx, tx)

	if _, err := r.loadSessionLayoutRow(ctx, tx, sessionID); err != nil {
		return nil, err
	}
	if _, err := tx.Exec(ctx, `
		UPDATE sessions
		SET layout_json = NULL, updated_at = $2
		WHERE id = $1
	`, sessionID, now); err != nil {
		return nil, err
	}

	result, err := r.syncSessionEffectiveLayout(ctx, tx, sessionID, now)
	if err != nil {
		return nil, err
	}
	if err := tx.Commit(ctx); err != nil {
		return nil, err
	}
	return result, nil
}

func (r *PostgresRepository) GetHallLayoutState(ctx context.Context, hallID string) (*integrationsapp.HallLayoutState, error) {
	var (
		state   integrationsapp.HallLayoutState
		rawJSON []byte
	)
	err := r.db.QueryRow(ctx, `
		SELECT h.id, h.name, v.id, v.name, COALESCE(h.layout_json, '{}'::jsonb)::text
		FROM halls h
		JOIN venues v ON v.id = h.venue_id
		WHERE h.id = $1
	`, hallID).Scan(&state.HallID, &state.Name, &state.Venue.ID, &state.Venue.Name, &rawJSON)
	if err != nil {
		return nil, err
	}
	state.Layout, err = decodeLayoutText(rawJSON)
	if err != nil {
		return nil, err
	}

	rows, err := r.db.Query(ctx, `
		SELECT s.id, s.event_id, e.title, s.starts_at, s.is_bookable
		FROM sessions s
		JOIN events e ON e.id = s.event_id
		WHERE s.hall_id = $1
		ORDER BY s.starts_at ASC NULLS LAST, s.created_at ASC
	`, hallID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	state.Sessions = make([]integrationsapp.HallLayoutSessionSummary, 0)
	for rows.Next() {
		var item integrationsapp.HallLayoutSessionSummary
		var startsAt sql.NullTime
		if err := rows.Scan(&item.ID, &item.EventID, &item.EventTitle, &startsAt, &item.IsBookable); err != nil {
			return nil, err
		}
		if startsAt.Valid {
			value := startsAt.Time
			item.StartsAt = &value
		}
		state.Sessions = append(state.Sessions, item)
	}
	return &state, rows.Err()
}

func (r *PostgresRepository) UpsertHallLayout(ctx context.Context, hallID string, layout seatsdomain.StoredSeatLayout, now time.Time) (*integrationsapp.LayoutMutationResult, error) {
	tx, err := r.db.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		return nil, err
	}
	defer rollback(ctx, tx)

	var exists string
	if err := tx.QueryRow(ctx, `SELECT id FROM halls WHERE id = $1 FOR UPDATE`, hallID).Scan(&exists); err != nil {
		return nil, err
	}

	encoded, err := json.Marshal(layout)
	if err != nil {
		return nil, err
	}
	if _, err := tx.Exec(ctx, `
		UPDATE halls
		SET layout_json = $2, updated_at = $3
		WHERE id = $1
	`, hallID, encoded, now); err != nil {
		return nil, err
	}

	result, err := r.syncHallSessions(ctx, tx, hallID, now)
	if err != nil {
		return nil, err
	}
	if err := tx.Commit(ctx); err != nil {
		return nil, err
	}
	return result, nil
}

func (r *PostgresRepository) ensureDefaultHallForVenue(ctx context.Context, tx pgx.Tx, venueID string, venueName string, now time.Time) (string, error) {
	venueID = strings.TrimSpace(venueID)
	if venueID == "" {
		return "", nil
	}

	var hallID string
	err := tx.QueryRow(ctx, `
		SELECT id
		FROM halls
		WHERE venue_id = $1
		ORDER BY created_at ASC
		LIMIT 1
	`, venueID).Scan(&hallID)
	if err == nil {
		return hallID, nil
	}
	if !errors.Is(err, pgx.ErrNoRows) {
		return "", err
	}

	hallID = uuid.NewString()
	hallName := strings.TrimSpace(venueName)
	if hallName == "" {
		hallName = "Main Hall"
	}
	if _, err := tx.Exec(ctx, `
		INSERT INTO halls (id, venue_id, name, rows_count, seats_per_row, created_at, updated_at)
		VALUES ($1, $2, $3, 0, 0, $4, $4)
	`, hallID, venueID, hallName, now); err != nil {
		return "", err
	}
	return hallID, nil
}

func (r *PostgresRepository) ensureSessionOverrideHall(ctx context.Context, tx pgx.Tx, row sessionLayoutRow, now time.Time) (string, error) {
	if row.HallID.Valid && len(row.SessionLayout) > 0 {
		return row.HallID.String, nil
	}
	if !row.VenueID.Valid {
		return "", fmt.Errorf("session venue is missing")
	}

	overrideHallID := uuid.NewString()
	overrideHallName := strings.TrimSpace(row.EventTitle)
	if overrideHallName == "" {
		overrideHallName = "Session Layout"
	}
	if _, err := tx.Exec(ctx, `
		INSERT INTO halls (id, venue_id, name, rows_count, seats_per_row, created_at, updated_at)
		VALUES ($1, $2, $3, 0, 0, $4, $4)
	`, overrideHallID, row.VenueID.String, overrideHallName, now); err != nil {
		return "", err
	}
	return overrideHallID, nil
}

func (r *PostgresRepository) syncHallSessions(ctx context.Context, tx pgx.Tx, hallID string, now time.Time) (*integrationsapp.LayoutMutationResult, error) {
	rows, err := tx.Query(ctx, `
		SELECT id
		FROM sessions
		WHERE hall_id = $1 AND layout_json IS NULL
		ORDER BY starts_at ASC NULLS LAST, created_at ASC
	`, hallID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	sessionIDs := make([]string, 0)
	for rows.Next() {
		var sessionID string
		if err := rows.Scan(&sessionID); err != nil {
			return nil, err
		}
		sessionIDs = append(sessionIDs, sessionID)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	if len(sessionIDs) == 0 {
		return r.snapshotHallResult(ctx, tx, hallID)
	}

	var last *integrationsapp.LayoutMutationResult
	for _, sessionID := range sessionIDs {
		last, err = r.syncSessionEffectiveLayout(ctx, tx, sessionID, now)
		if err != nil {
			return nil, err
		}
	}
	return last, nil
}

func (r *PostgresRepository) syncSessionEffectiveLayout(ctx context.Context, tx pgx.Tx, sessionID string, now time.Time) (*integrationsapp.LayoutMutationResult, error) {
	row, err := r.loadSessionLayoutRow(ctx, tx, sessionID)
	if err != nil {
		return nil, err
	}
	state, err := buildSessionLayoutState(row)
	if err != nil {
		return nil, err
	}

	result := &integrationsapp.LayoutMutationResult{
		SessionID:  &row.SessionID,
		EventID:    row.EventID,
		EventTitle: row.EventTitle,
		HallID:     "",
	}
	if row.HallID.Valid {
		result.HallID = row.HallID.String
	}

	if state.EffectiveLayout == nil || !row.HallID.Valid {
		if _, err := tx.Exec(ctx, `
			UPDATE sessions
			SET is_bookable = FALSE, updated_at = $2
			WHERE id = $1
		`, sessionID, now); err != nil {
			return nil, err
		}
		if _, err := tx.Exec(ctx, `
			UPDATE session_seats
			SET is_active = FALSE, updated_at = $2
			WHERE session_id = $1
		`, sessionID, now); err != nil {
			return nil, err
		}
		bookingMode, visible, err := refreshEventBookability(ctx, tx, row.EventID, row.EventSource, now)
		if err != nil {
			return nil, err
		}
		result.BookingMode = bookingMode
		result.VisibleInCatalog = visible
		result.IsBookable = false
		return result, nil
	}

	layout := *state.EffectiveLayout
	seatIDs, err := upsertLayoutSeats(ctx, tx, row.HallID.String, layout, now)
	if err != nil {
		return nil, err
	}
	if err := upsertSessionSeats(ctx, tx, sessionID, seatIDs, now); err != nil {
		return nil, err
	}
	if _, err := tx.Exec(ctx, `
		UPDATE sessions
		SET is_bookable = TRUE, updated_at = $2
		WHERE id = $1
	`, sessionID, now); err != nil {
		return nil, err
	}

	bookingMode, visible, err := refreshEventBookability(ctx, tx, row.EventID, row.EventSource, now)
	if err != nil {
		return nil, err
	}
	result.BookingMode = bookingMode
	result.VisibleInCatalog = visible
	result.IsBookable = true
	result.MaterializedSeats = len(seatIDs)
	return result, nil
}

func (r *PostgresRepository) snapshotHallResult(ctx context.Context, tx pgx.Tx, hallID string) (*integrationsapp.LayoutMutationResult, error) {
	var (
		sessionID  sql.NullString
		eventID    sql.NullString
		eventTitle sql.NullString
	)
	if err := tx.QueryRow(ctx, `
		SELECT s.id, s.event_id, e.title
		FROM sessions s
		JOIN events e ON e.id = s.event_id
		WHERE s.hall_id = $1
		ORDER BY s.starts_at ASC NULLS LAST, s.created_at ASC
		LIMIT 1
	`, hallID).Scan(&sessionID, &eventID, &eventTitle); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return &integrationsapp.LayoutMutationResult{HallID: hallID}, nil
		}
		return nil, err
	}

	result := &integrationsapp.LayoutMutationResult{
		HallID:     hallID,
		EventID:    eventID.String,
		EventTitle: eventTitle.String,
	}
	if sessionID.Valid {
		value := sessionID.String
		result.SessionID = &value
	}
	var bookableSessions int
	if err := tx.QueryRow(ctx, `
		SELECT count(*) FROM sessions WHERE hall_id = $1 AND is_bookable = TRUE
	`, hallID).Scan(&bookableSessions); err != nil {
		return nil, err
	}
	result.IsBookable = bookableSessions > 0
	result.VisibleInCatalog = bookableSessions > 0
	return result, nil
}

func (r *PostgresRepository) loadSessionLayoutRow(ctx context.Context, q queryRower, sessionID string) (sessionLayoutRow, error) {
	var row sessionLayoutRow
	err := q.QueryRow(ctx, `
		SELECT
			s.id,
			s.event_id,
			e.title,
			e.source,
			e.booking_mode,
			s.is_bookable,
			h.id,
			h.name,
			v.id,
			v.name,
			COALESCE(s.layout_json, '{}'::jsonb)::text,
			COALESCE(h.layout_json, '{}'::jsonb)::text
		FROM sessions s
		JOIN events e ON e.id = s.event_id
		LEFT JOIN halls h ON h.id = s.hall_id
		LEFT JOIN venues v ON v.id = COALESCE(h.venue_id, e.venue_id)
		WHERE s.id = $1
		FOR UPDATE OF s
	`, sessionID).Scan(
		&row.SessionID,
		&row.EventID,
		&row.EventTitle,
		&row.EventSource,
		&row.BookingMode,
		&row.IsBookable,
		&row.HallID,
		&row.HallName,
		&row.VenueID,
		&row.VenueName,
		&row.SessionLayout,
		&row.HallLayout,
	)
	return row, err
}

type queryRower interface {
	QueryRow(ctx context.Context, sql string, args ...any) pgx.Row
}

func buildSessionLayoutState(row sessionLayoutRow) (*integrationsapp.SessionLayoutState, error) {
	sessionLayout, err := decodeLayoutText(row.SessionLayout)
	if err != nil {
		return nil, err
	}
	hallLayout, err := decodeLayoutText(row.HallLayout)
	if err != nil {
		return nil, err
	}

	state := &integrationsapp.SessionLayoutState{
		SessionID:      row.SessionID,
		EventID:        row.EventID,
		EventTitle:     row.EventTitle,
		Source:         row.EventSource,
		BookingMode:    row.BookingMode,
		IsBookable:     row.IsBookable,
		Layout:         sessionLayout,
		FallbackLayout: hallLayout,
		LayoutSource:   integrationsapp.LayoutScopeNone,
	}
	if row.HallID.Valid {
		state.Hall = &integrationsapp.LayoutHallRef{
			ID:   row.HallID.String,
			Name: row.HallName.String,
			Venue: integrationsapp.LayoutVenueRef{
				ID:   row.VenueID.String,
				Name: row.VenueName.String,
			},
		}
	}

	switch {
	case sessionLayout != nil:
		state.EffectiveLayout = sessionLayout
		state.LayoutSource = integrationsapp.LayoutScopeSession
	case hallLayout != nil:
		state.EffectiveLayout = hallLayout
		state.LayoutSource = integrationsapp.LayoutScopeHall
	}
	return state, nil
}

func decodeLayoutText(raw []byte) (*seatsdomain.StoredSeatLayout, error) {
	text := strings.TrimSpace(string(raw))
	if text == "" || text == "{}" || text == "null" {
		return nil, nil
	}
	return seatsdomain.DecodeStoredSeatLayout([]byte(text))
}

func upsertLayoutSeats(ctx context.Context, tx pgx.Tx, hallID string, layout seatsdomain.StoredSeatLayout, now time.Time) ([]string, error) {
	rowsCount, seatsPerRow := layout.Dimensions()
	if _, err := tx.Exec(ctx, `
		UPDATE halls
		SET rows_count = $2, seats_per_row = $3, updated_at = $4
		WHERE id = $1
	`, hallID, rowsCount, seatsPerRow, now); err != nil {
		return nil, err
	}

	seatIDs := make([]string, 0, len(layout.Seats))
	for _, seat := range layout.SortedSeats() {
		seatID := deterministicSeatID(hallID, seat.Row, seat.Number)
		seatType := strings.TrimSpace(seat.Category)
		if seatType == "" {
			seatType = "standard"
		}
		if _, err := tx.Exec(ctx, `
			INSERT INTO seats (id, hall_id, layout_key, row_label, seat_number, seat_type, base_price, created_at, updated_at)
			VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $8)
			ON CONFLICT (id) DO UPDATE SET
				layout_key = EXCLUDED.layout_key,
				row_label = EXCLUDED.row_label,
				seat_number = EXCLUDED.seat_number,
				seat_type = EXCLUDED.seat_type,
				base_price = EXCLUDED.base_price,
				updated_at = EXCLUDED.updated_at
		`, seatID, hallID, seat.Key, strings.TrimSpace(seat.Row), seat.Number, seatType, seat.Price, now); err != nil {
			return nil, err
		}
		seatIDs = append(seatIDs, seatID)
	}
	return seatIDs, nil
}

func upsertSessionSeats(ctx context.Context, tx pgx.Tx, sessionID string, seatIDs []string, now time.Time) error {
	for _, seatID := range seatIDs {
		sessionSeatID := deterministicSessionSeatID(sessionID, seatID)
		if _, err := tx.Exec(ctx, `
			INSERT INTO session_seats (id, session_id, seat_id, status, hold_expires_at, version, updated_at, is_active)
			VALUES ($1, $2, $3, 'available', NULL, 1, $4, TRUE)
			ON CONFLICT (session_id, seat_id) DO UPDATE SET
				is_active = TRUE,
				status = CASE
					WHEN session_seats.status = 'booked' THEN 'booked'
					ELSE 'available'
				END,
				hold_expires_at = NULL,
				updated_at = EXCLUDED.updated_at
		`, sessionSeatID, sessionID, seatID, now); err != nil {
			return err
		}
	}

	if len(seatIDs) == 0 {
		_, err := tx.Exec(ctx, `
			UPDATE session_seats
			SET is_active = FALSE, updated_at = $2
			WHERE session_id = $1
		`, sessionID, now)
		return err
	}

	_, err := tx.Exec(ctx, `
		UPDATE session_seats
		SET is_active = FALSE, updated_at = $2
		WHERE session_id = $1 AND NOT (seat_id = ANY($3::uuid[]))
	`, sessionID, now, seatIDs)
	return err
}

func refreshEventBookability(ctx context.Context, tx pgx.Tx, eventID, source string, now time.Time) (string, bool, error) {
	bookingMode := eventsdomain.BookingModeReserveFlowManaged
	visible := false

	var bookableCount int
	if err := tx.QueryRow(ctx, `
		SELECT count(*)
		FROM sessions
		WHERE event_id = $1 AND status = 'scheduled' AND is_bookable = TRUE
	`, eventID).Scan(&bookableCount); err != nil {
		return "", false, err
	}
	if bookableCount > 0 {
		visible = true
	}

	if _, err := tx.Exec(ctx, `
		UPDATE events
		SET booking_mode = $2, updated_at = $3
		WHERE id = $1
	`, eventID, bookingMode, now); err != nil {
		return "", false, err
	}
	return bookingMode, visible, nil
}

func deterministicSeatID(hallID, row string, number int) string {
	return uuid.NewMD5(uuid.NameSpaceURL, []byte(fmt.Sprintf("reserveflow:hall:%s:seat:%s:%d", hallID, strings.TrimSpace(row), number))).String()
}

func deterministicSessionSeatID(sessionID, seatID string) string {
	return uuid.NewMD5(uuid.NameSpaceURL, []byte(fmt.Sprintf("reserveflow:session:%s:seat:%s", sessionID, seatID))).String()
}
