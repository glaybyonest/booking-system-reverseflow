//go:build integration

package tests

import (
	"context"
	"encoding/json"
	"path/filepath"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/stretchr/testify/require"

	eventsdomain "reserveflow/backend/internal/modules/events/domain"
	seatsdomain "reserveflow/backend/internal/modules/seats/domain"
)

type seededEventFixture struct {
	EventID   string
	VenueID   string
	HallID    string
	SessionID string
	Layout    seatsdomain.StoredSeatLayout
	SeatIDs   []string
}

type seedEventOptions struct {
	Source             string
	BookingMode        string
	Bookable           bool
	IncludeCoordinates bool
	Layout             *seatsdomain.StoredSeatLayout
	Title              string
	Category           string
}

func applyIntegrationMigrations(t *testing.T, ctx context.Context, databaseURL string) {
	t.Helper()
	for _, name := range []string{
		"000001_init.up.sql",
		"000003_external_events.up.sql",
		"000005_exact_layouts.up.sql",
	} {
		applySQL(t, ctx, databaseURL, filepath.Join("..", "migrations", name))
	}
}

func seedBookableKudaGoEvent(t *testing.T, ctx context.Context, db *pgxpool.Pool, title string) seededEventFixture {
	t.Helper()
	return seedEventFixture(t, ctx, db, seedEventOptions{
		Source:             eventsdomain.SourceKudaGo,
		BookingMode:        eventsdomain.BookingModeReserveFlowManaged,
		Bookable:           true,
		IncludeCoordinates: true,
		Layout:             layoutPtr(sampleStoredLayout(2, 3)),
		Title:              title,
		Category:           "concert",
	})
}

func seedEventFixture(t *testing.T, ctx context.Context, db *pgxpool.Pool, options seedEventOptions) seededEventFixture {
	t.Helper()

	now := time.Now().UTC()
	startsAt := now.Add(24 * time.Hour).Truncate(time.Second)
	endsAt := startsAt.Add(2 * time.Hour)
	source := strings.TrimSpace(options.Source)
	if source == "" {
		source = eventsdomain.SourceKudaGo
	}
	bookingMode := strings.TrimSpace(options.BookingMode)
	if bookingMode == "" {
		if source == eventsdomain.SourceKudaGo && options.Bookable {
			bookingMode = eventsdomain.BookingModeReserveFlowManaged
		} else {
			bookingMode = eventsdomain.BookingModeExternalLinkOnly
		}
	}
	title := strings.TrimSpace(options.Title)
	if title == "" {
		title = "Integration event"
	}
	category := strings.TrimSpace(options.Category)
	if category == "" {
		category = "concert"
	}

	fixture := seededEventFixture{
		EventID:   uuid.NewString(),
		VenueID:   uuid.NewString(),
		HallID:    uuid.NewString(),
		SessionID: uuid.NewString(),
	}

	rowsCount := 0
	seatsPerRow := 0
	var layoutJSON []byte
	if options.Layout != nil {
		fixture.Layout = *options.Layout
		rowsCount, seatsPerRow = fixture.Layout.Dimensions()
		var err error
		layoutJSON, err = json.Marshal(fixture.Layout)
		require.NoError(t, err)
	}

	var latitude any
	var longitude any
	if options.IncludeCoordinates {
		latitude = 55.7522
		longitude = 37.6156
	}

	_, err := db.Exec(ctx, `
		INSERT INTO venues (
			id, name, address, city, external_source, external_id,
			latitude, longitude, seat_map_provider, created_at, updated_at
		)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, 'internal_grid', $9, $9)
	`, fixture.VenueID, title+" venue", "Tverskaya 1", "Moscow", source, "venue-"+fixture.VenueID, latitude, longitude, now)
	require.NoError(t, err)

	_, err = db.Exec(ctx, `
		INSERT INTO halls (id, venue_id, name, rows_count, seats_per_row, layout_json, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $7)
	`, fixture.HallID, fixture.VenueID, title+" hall", rowsCount, seatsPerRow, layoutJSON, now)
	require.NoError(t, err)

	_, err = db.Exec(ctx, `
		INSERT INTO events (
			id, title, description, category, poster_url, status,
			source, external_source, external_id, source_url,
			imported_at, last_synced_at, raw_payload, booking_mode,
			starts_at, ends_at, normalized_title, dedupe_key, venue_id,
			created_at, updated_at
		)
		VALUES (
			$1, $2, $3, $4, $5, 'published',
			$6, $6, $7, $8,
			$9, $9, '{}'::jsonb, $10,
			$11, $12, $13, $14, $15,
			$9, $9
		)
	`, fixture.EventID, title, title+" description", category, "https://images.example.com/poster.jpg",
		source, "event-"+fixture.EventID, "https://kudago.com/msk/event/"+fixture.EventID+"/",
		now, bookingMode, startsAt, endsAt, strings.ToLower(title), strings.ToLower(title)+"|"+startsAt.Format("2006-01-02"), fixture.VenueID)
	require.NoError(t, err)

	_, err = db.Exec(ctx, `
		INSERT INTO sessions (
			id, event_id, hall_id, starts_at, ends_at, status,
			external_source, external_id, source_url, is_bookable, created_at, updated_at
		)
		VALUES (
			$1, $2, $3, $4, $5, 'scheduled',
			$6, $7, $8, $9, $10, $10
		)
	`, fixture.SessionID, fixture.EventID, fixture.HallID, startsAt, endsAt,
		source, "session-"+fixture.SessionID, "https://kudago.com/msk/event/"+fixture.EventID+"/tickets/", options.Bookable, now)
	require.NoError(t, err)

	if options.Layout == nil {
		return fixture
	}

	for _, seat := range fixture.Layout.SortedSeats() {
		seatID := uuid.NewString()
		fixture.SeatIDs = append(fixture.SeatIDs, seatID)
		_, err = db.Exec(ctx, `
			INSERT INTO seats (
				id, hall_id, layout_key, row_label, seat_number, seat_type, base_price, created_at, updated_at
			)
			VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $8)
		`, seatID, fixture.HallID, seat.Key, seat.Row, seat.Number, defaultSeatCategory(seat.Category), seat.Price, now)
		require.NoError(t, err)

		_, err = db.Exec(ctx, `
			INSERT INTO session_seats (
				id, session_id, seat_id, status, hold_expires_at, version, updated_at, is_active
			)
			VALUES ($1, $2, $3, 'available', NULL, 1, $4, $5)
		`, uuid.NewString(), fixture.SessionID, seatID, now, options.Bookable)
		require.NoError(t, err)
	}

	return fixture
}

func sampleStoredLayout(rows, seatsPerRow int) seatsdomain.StoredSeatLayout {
	seats := make([]seatsdomain.StoredLayoutSeat, 0, rows*seatsPerRow)
	for rowIndex := 0; rowIndex < rows; rowIndex++ {
		row := string(rune('A' + rowIndex))
		for seatIndex := 0; seatIndex < seatsPerRow; seatIndex++ {
			number := seatIndex + 1
			seats = append(seats, seatsdomain.StoredLayoutSeat{
				Key:    row + "-" + strconvItoa(number),
				Label:  row + "-" + strconvItoa(number),
				Row:    row,
				Number: number,
				X:      120 + seatIndex*54,
				Y:      180 + rowIndex*58,
				Price:  1200 + float64(rowIndex*200),
			})
		}
	}

	return seatsdomain.StoredSeatLayout{
		Version: 1,
		Canvas: seatsdomain.StoredLayoutCanvas{
			Width:  520,
			Height: 420,
		},
		Stage: &seatsdomain.StoredLayoutStage{
			Label:  "Stage",
			X:      140,
			Y:      60,
			Width:  240,
			Height: 52,
		},
		Seats: seats,
	}
}

func layoutPtr(layout seatsdomain.StoredSeatLayout) *seatsdomain.StoredSeatLayout {
	return &layout
}

func defaultSeatCategory(value string) string {
	value = strings.TrimSpace(value)
	if value == "" {
		return "standard"
	}
	return value
}

func strconvItoa(value int) string {
	return strconv.Itoa(value)
}
