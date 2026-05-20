package repository

import (
	"context"
	"database/sql"
	"fmt"
	"strings"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"reserveflow/backend/internal/modules/events/domain"
)

type PostgresRepository struct {
	db *pgxpool.Pool
}

func NewPostgresRepository(db *pgxpool.Pool) *PostgresRepository {
	return &PostgresRepository{db: db}
}

// eventSelectColumns is the shared column list for scanEvent.
// Order must match the Scan() call in scanEvent exactly.
const eventSelectColumns = `
	e.id,
	e.title,
	e.description,
	e.long_description,
	e.category,
	e.poster_url,
	e.status,
	e.source,
	e.external_source,
	e.source_url,
	e.booking_mode,
	e.starts_at,
	e.ends_at,
	e.created_at,
	e.updated_at,
	e.age_restriction,
	e.price_min,
	e.price_max,
	e.tags,
	e.rating_count,
	v.id,
	v.name,
	v.address,
	v.city,
	v.latitude,
	v.longitude,
	v.metro_stations,
	v.venue_type_code,
	v.venue_type_name
`

func (r *PostgresRepository) ListEvents(ctx context.Context, query domain.ListQuery) ([]domain.Event, int, error) {
	conditions, args := buildEventFilters(query, false)
	where := strings.Join(conditions, " AND ")

	countSQL := fmt.Sprintf(`
		SELECT count(*)
		FROM events e
		LEFT JOIN venues v ON v.id = e.venue_id
		WHERE %s
	`, where)
	var total int
	if err := r.db.QueryRow(ctx, countSQL, args...).Scan(&total); err != nil {
		return nil, 0, err
	}

	args = append(args, query.Limit, query.Offset)
	listSQL := fmt.Sprintf(`
		SELECT %s
		FROM events e
		LEFT JOIN venues v ON v.id = e.venue_id
		WHERE %s
		ORDER BY e.starts_at ASC NULLS LAST, e.created_at DESC
		LIMIT $%d OFFSET $%d
	`, eventSelectColumns, where, len(args)-1, len(args))
	rows, err := r.db.Query(ctx, listSQL, args...)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	events := make([]domain.Event, 0)
	for rows.Next() {
		event, err := scanEvent(rows)
		if err != nil {
			return nil, 0, err
		}
		events = append(events, event)
	}
	return events, total, rows.Err()
}

func (r *PostgresRepository) ListMapEvents(ctx context.Context, query domain.ListQuery) ([]domain.MapEvent, error) {
	conditions, args := buildEventFilters(query, true)
	where := strings.Join(conditions, " AND ")

	sqlStr := fmt.Sprintf(`
		SELECT
			e.id,
			e.title,
			e.category,
			e.poster_url,
			e.source,
			e.booking_mode,
			e.starts_at,
			e.ends_at,
			e.price_min,
			v.id,
			v.name,
			v.address,
			v.city,
			v.latitude,
			v.longitude,
			v.metro_stations,
			v.venue_type_code,
			v.venue_type_name
		FROM events e
		LEFT JOIN venues v ON v.id = e.venue_id
		WHERE %s
		ORDER BY e.starts_at ASC NULLS LAST, e.created_at DESC
		LIMIT $%d OFFSET $%d
	`, where, len(args)+1, len(args)+2)
	args = append(args, query.Limit, query.Offset)

	rows, err := r.db.Query(ctx, sqlStr, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	events := make([]domain.MapEvent, 0)
	for rows.Next() {
		item, err := scanMapEvent(rows)
		if err != nil {
			return nil, err
		}
		events = append(events, item)
	}
	return events, rows.Err()
}

func (r *PostgresRepository) GetEvent(ctx context.Context, id string) (*domain.Event, error) {
	row := r.db.QueryRow(ctx, fmt.Sprintf(`
		SELECT %s
		FROM events e
		LEFT JOIN venues v ON v.id = e.venue_id
		WHERE e.id = $1
	`, eventSelectColumns), id)
	event, err := scanEvent(row)
	if err != nil {
		return nil, err
	}
	return &event, nil
}

func (r *PostgresRepository) GetEventExternalLinks(ctx context.Context, eventID string) ([]domain.ExternalLink, error) {
	rows, err := r.db.Query(ctx, `
		SELECT id, external_source, external_id, source_url, imported_at
		FROM event_external_links
		WHERE event_id = $1
		ORDER BY imported_at DESC
	`, eventID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	links := make([]domain.ExternalLink, 0)
	for rows.Next() {
		var item domain.ExternalLink
		var sourceURL sql.NullString
		if err := rows.Scan(&item.ID, &item.ExternalSource, &item.ExternalID, &sourceURL, &item.ImportedAt); err != nil {
			return nil, err
		}
		if sourceURL.Valid {
			item.SourceURL = &sourceURL.String
		}
		links = append(links, item)
	}
	return links, rows.Err()
}

func (r *PostgresRepository) GetEventSessions(ctx context.Context, eventID string) ([]domain.SessionSummary, error) {
	rows, err := r.db.Query(ctx, `
		SELECT
			s.id,
			s.event_id,
			s.hall_id,
			h.name,
			s.starts_at,
			s.ends_at,
			s.status,
			s.is_bookable,
			s.external_source,
			s.source_url
		FROM sessions s
		LEFT JOIN halls h ON h.id = s.hall_id
		WHERE s.event_id = $1
		ORDER BY s.starts_at ASC NULLS LAST, s.created_at ASC
	`, eventID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	sessions := make([]domain.SessionSummary, 0)
	for rows.Next() {
		var item domain.SessionSummary
		var hallID sql.NullString
		var hallName sql.NullString
		var startsAt sql.NullTime
		var endsAt sql.NullTime
		var externalSource sql.NullString
		var sourceURL sql.NullString
		if err := rows.Scan(
			&item.ID,
			&item.EventID,
			&hallID,
			&hallName,
			&startsAt,
			&endsAt,
			&item.Status,
			&item.IsBookable,
			&externalSource,
			&sourceURL,
		); err != nil {
			return nil, err
		}
		if hallID.Valid {
			item.HallID = &hallID.String
		}
		if hallName.Valid {
			item.HallName = &hallName.String
		}
		if startsAt.Valid {
			start := startsAt.Time
			item.StartsAt = &start
		}
		if endsAt.Valid {
			end := endsAt.Time
			item.EndsAt = &end
		}
		if externalSource.Valid {
			item.ExternalSource = &externalSource.String
		}
		if sourceURL.Valid {
			item.SourceURL = &sourceURL.String
		}
		sessions = append(sessions, item)
	}
	return sessions, rows.Err()
}

// ─── Scanners ─────────────────────────────────────────────────────────────────

type eventScanner interface {
	Scan(dest ...any) error
}

func scanEvent(scanner eventScanner) (domain.Event, error) {
	var item domain.Event
	var description sql.NullString
	var longDescription sql.NullString
	var category sql.NullString
	var posterURL sql.NullString
	var externalSource sql.NullString
	var sourceURL sql.NullString
	var startsAt sql.NullTime
	var endsAt sql.NullTime
	var ageRestriction sql.NullString
	var priceMin sql.NullFloat64
	var priceMax sql.NullFloat64
	var tags []string
	var ratingCount sql.NullInt32
	var venueID sql.NullString
	var venueName sql.NullString
	var venueAddress sql.NullString
	var venueCity sql.NullString
	var latitude sql.NullFloat64
	var longitude sql.NullFloat64
	var metroStations []string
	var venueTypeCode sql.NullString
	var venueTypeName sql.NullString

	err := scanner.Scan(
		&item.ID,
		&item.Title,
		&description,
		&longDescription,
		&category,
		&posterURL,
		&item.Status,
		&item.Source,
		&externalSource,
		&sourceURL,
		&item.BookingMode,
		&startsAt,
		&endsAt,
		&item.CreatedAt,
		&item.UpdatedAt,
		&ageRestriction,
		&priceMin,
		&priceMax,
		&tags,
		&ratingCount,
		&venueID,
		&venueName,
		&venueAddress,
		&venueCity,
		&latitude,
		&longitude,
		&metroStations,
		&venueTypeCode,
		&venueTypeName,
	)
	if err != nil {
		return domain.Event{}, err
	}

	item.IsImported = item.Source != domain.SourceManual
	if description.Valid {
		item.Description = &description.String
	}
	if longDescription.Valid {
		item.LongDescription = &longDescription.String
	}
	if category.Valid {
		item.Category = &category.String
	}
	if posterURL.Valid {
		item.PosterURL = &posterURL.String
	}
	if externalSource.Valid {
		item.ExternalSource = &externalSource.String
	}
	if sourceURL.Valid {
		item.SourceURL = &sourceURL.String
	}
	if startsAt.Valid {
		start := startsAt.Time
		item.StartsAt = &start
	}
	if endsAt.Valid {
		end := endsAt.Time
		item.EndsAt = &end
	}
	if ageRestriction.Valid {
		item.AgeRestriction = &ageRestriction.String
	}
	if priceMin.Valid {
		item.PriceMin = &priceMin.Float64
	}
	if priceMax.Valid {
		item.PriceMax = &priceMax.Float64
	}
	if len(tags) > 0 {
		item.Tags = tags
	}
	if ratingCount.Valid {
		n := int(ratingCount.Int32)
		item.RatingCount = &n
	}
	if venueID.Valid {
		item.Venue = &domain.VenueSummary{
			ID:      venueID.String,
			Name:    venueName.String,
			Address: venueAddress.String,
			City:    venueCity.String,
		}
		if latitude.Valid {
			item.Venue.Latitude = &latitude.Float64
		}
		if longitude.Valid {
			item.Venue.Longitude = &longitude.Float64
		}
		if len(metroStations) > 0 {
			item.Venue.MetroStations = metroStations
		}
		if venueTypeCode.Valid {
			item.Venue.VenueTypeCode = &venueTypeCode.String
		}
		if venueTypeName.Valid {
			item.Venue.VenueTypeName = &venueTypeName.String
		}
	}
	return item, nil
}

func scanMapEvent(scanner eventScanner) (domain.MapEvent, error) {
	var item domain.MapEvent
	var category sql.NullString
	var posterURL sql.NullString
	var startsAt sql.NullTime
	var endsAt sql.NullTime
	var priceMin sql.NullFloat64
	var venueID sql.NullString
	var venueName sql.NullString
	var venueAddress sql.NullString
	var venueCity sql.NullString
	var latitude float64
	var longitude float64
	var metroStations []string
	var venueTypeCode sql.NullString
	var venueTypeName sql.NullString

	err := scanner.Scan(
		&item.ID,
		&item.Title,
		&category,
		&posterURL,
		&item.Source,
		&item.BookingMode,
		&startsAt,
		&endsAt,
		&priceMin,
		&venueID,
		&venueName,
		&venueAddress,
		&venueCity,
		&latitude,
		&longitude,
		&metroStations,
		&venueTypeCode,
		&venueTypeName,
	)
	if err != nil {
		return domain.MapEvent{}, err
	}
	if category.Valid {
		item.Category = &category.String
	}
	if posterURL.Valid {
		item.PosterURL = &posterURL.String
	}
	if startsAt.Valid {
		start := startsAt.Time
		item.StartsAt = &start
	}
	if endsAt.Valid {
		end := endsAt.Time
		item.EndsAt = &end
	}
	if priceMin.Valid {
		item.PriceMin = &priceMin.Float64
	}

	venue := domain.VenueSummary{
		ID:        venueID.String,
		Name:      venueName.String,
		Address:   venueAddress.String,
		City:      venueCity.String,
		Latitude:  &latitude,
		Longitude: &longitude,
	}
	if len(metroStations) > 0 {
		venue.MetroStations = metroStations
	}
	if venueTypeCode.Valid {
		venue.VenueTypeCode = &venueTypeCode.String
	}
	if venueTypeName.Valid {
		venue.VenueTypeName = &venueTypeName.String
	}
	item.Venue = venue
	return item, nil
}

// ─── Filter builder ───────────────────────────────────────────────────────────

func buildEventFilters(query domain.ListQuery, requireCoordinates bool) ([]string, []any) {
	conditions := []string{
		"e.status IN ('published', 'active')",
	}
	args := make([]any, 0)

	if query.City != "" {
		normalizedCity := strings.ToLower(strings.TrimSpace(query.City))
		if normalizedCity == "moscow" || normalizedCity == "москва" {
			conditions = append(conditions, "LOWER(COALESCE(v.city, '')) IN ('moscow', 'москва')")
		} else {
			args = append(args, normalizedCity)
			conditions = append(conditions, fmt.Sprintf("LOWER(COALESCE(v.city, '')) = $%d", len(args)))
		}
	}
	if query.Source != "" {
		source := query.Source
		if source == "reserveflow" {
			source = domain.SourceManual
		}
		args = append(args, source)
		conditions = append(conditions, fmt.Sprintf("e.source = $%d", len(args)))
	}
	if query.Category != "" {
		// "lectures" and "lecture" are the same category (data inconsistency)
		if query.Category == "lectures" || query.Category == "lecture" {
			conditions = append(conditions, "e.category IN ('lectures', 'lecture')")
		} else {
			args = append(args, query.Category)
			conditions = append(conditions, fmt.Sprintf("e.category = $%d", len(args)))
		}
	}
	if query.BookingMode != "" {
		if query.BookingMode == "bookable" {
			conditions = append(conditions, "e.booking_mode = 'reserveflow_managed'")
		} else {
			args = append(args, query.BookingMode)
			conditions = append(conditions, fmt.Sprintf("e.booking_mode = $%d", len(args)))
		}
	}
	if query.From != nil {
		args = append(args, query.From.UTC())
		conditions = append(conditions, fmt.Sprintf("COALESCE(e.ends_at, e.starts_at) >= $%d", len(args)))
	}
	if query.To != nil {
		args = append(args, query.To.UTC())
		conditions = append(conditions, fmt.Sprintf("COALESCE(e.starts_at, e.ends_at) <= $%d", len(args)))
	}
	if query.OnlyActual {
		args = append(args, time.Now().UTC())
		conditions = append(conditions, fmt.Sprintf("COALESCE(e.ends_at, e.starts_at) >= $%d", len(args)))
	}
	if requireCoordinates {
		conditions = append(conditions, "v.latitude IS NOT NULL AND v.longitude IS NOT NULL")
	}
	return conditions, args
}

// pgxRows satisfies eventScanner for pgx.Rows.
var _ eventScanner = (pgx.Rows)(nil)
