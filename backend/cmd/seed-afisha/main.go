// seed-afisha imports Yandex Afisha JSON event files into the ReserveFlow database.
//
// Data sources (in results/ folder):
//   - venues.json      — 841 venue objects with metro_stations, type, coordinates
//   - all_events.json  — 5 681 events with occurrences, age_restriction,
//                        long_description, price_min/max, tags, rating_count
//
// Usage:
//
//	DATABASE_URL="postgres://user:pass@localhost:5432/reserveflow?sslmode=disable" \
//	  go run ./cmd/seed-afisha --dir ../../results --limit 30 --max-occ 10
//
// Flags:
//
//	--dir      path to the folder containing venues.json and all_events.json
//	           (default: ../../results)
//	--limit    max events to import per category  (default: 30)
//	--max-occ  max future occurrences to create per event (default: 10)
//	--dsn      PostgreSQL DSN (overrides DATABASE_URL env var)
//
// Date shifting:
//
//	The script shifts all occurrence dates forward so the earliest date becomes
//	tomorrow, preserving relative ordering.  This ensures seeded events appear in
//	the live calendar without any manual date edits.
package main

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"math"
	"os"
	"path/filepath"
	"strings"
	"time"
	"unicode"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
)

// ─── JSON input models ────────────────────────────────────────────────────────

type afishaVenue struct {
	VenueID       string   `json:"venue_id"`
	Title         string   `json:"title"`
	Address       string   `json:"address"`
	URL           string   `json:"url"`
	Logo          string   `json:"logo"`
	TypeCode      string   `json:"type_code"`
	TypeName      string   `json:"type_name"`
	Latitude      float64  `json:"latitude"`
	Longitude     float64  `json:"longitude"`
	Phones        []string `json:"phones"`
	Links         []string `json:"links"`
	MetroStations []string `json:"metro_stations"`
	Tags          []string `json:"tags"`
	TagCodes      []string `json:"tag_codes"`
	Categories    []string `json:"categories"`
}

type afishaOccurrence struct {
	Date       string   `json:"date"`
	PriceMin   *float64 `json:"price_min"`
	PriceMax   *float64 `json:"price_max"`
	Currency   string   `json:"currency"`
	Location   string   `json:"location"`
	VenueID    string   `json:"venue_id"`
	VenueURL   string   `json:"venue_url"`
	BookingURL string   `json:"booking_url"`
}

type afishaEvent struct {
	Category         string             `json:"category"`
	SourceSlug       string             `json:"source_slug"`
	City             string             `json:"city"`
	EventID          string             `json:"event_id"`
	Title            string             `json:"title"`
	URL              string             `json:"url"`
	CanonicalURL     string             `json:"canonical_url"`
	ShortDescription string             `json:"short_description"`
	LongDescription  string             `json:"long_description"`
	AgeRestriction   string             `json:"age_restriction"`
	DurationMinutes  *int               `json:"duration_minutes"`
	EventTypeCode    string             `json:"event_type_code"`
	EventTypeName    string             `json:"event_type_name"`
	RatingCount      int                `json:"rating_count"`
	Photo            string             `json:"photo"`
	PhotoHighres     string             `json:"photo_highres"`
	VenueID          string             `json:"venue_id"`
	Location         string             `json:"location"`
	VenueAddress     string             `json:"venue_address"`
	VenueTypeCode    string             `json:"venue_type_code"`
	VenueTypeName    string             `json:"venue_type_name"`
	MetroStations    []string           `json:"metro_stations"`
	Tags             []string           `json:"tags"`
	TagCodes         []string           `json:"tag_codes"`
	Occurrences      []afishaOccurrence `json:"occurrences"`
	PriceMin         *float64           `json:"price_min"`
	Currency         string             `json:"currency"`
	Date             string             `json:"date"`
}

// ─── Main ────────────────────────────────────────────────────────────────────

func main() {
	dir := flag.String("dir", "../../results", "folder with venues.json and all_events.json")
	limit := flag.Int("limit", 30, "max events per category")
	maxOcc := flag.Int("max-occ", 10, "max future occurrences to create per event")
	dsn := flag.String("dsn", "", "PostgreSQL DSN (falls back to DATABASE_URL env)")
	flag.Parse()

	connStr := *dsn
	if connStr == "" {
		connStr = os.Getenv("DATABASE_URL")
	}
	if connStr == "" {
		log.Fatal("DATABASE_URL env variable (or --dsn flag) is required")
	}

	ctx := context.Background()
	pool, err := pgxpool.New(ctx, connStr)
	if err != nil {
		log.Fatalf("connect to DB: %v", err)
	}
	defer pool.Close()
	if err := pool.Ping(ctx); err != nil {
		log.Fatalf("ping DB: %v", err)
	}
	log.Println("✓ connected to database")

	// ── Phase 1: load venues ─────────────────────────────────────────────────
	venuesPath := filepath.Join(*dir, "venues.json")
	afishaVenues, err := loadVenues(venuesPath)
	if err != nil {
		log.Fatalf("load venues: %v", err)
	}
	log.Printf("loaded %d venues from %s", len(afishaVenues), venuesPath)

	// Upsert all venues; build map afishaVenueID → db UUID
	venueMap := make(map[string]uuid.UUID, len(afishaVenues))
	for _, v := range afishaVenues {
		id, err := upsertVenueFull(ctx, pool, v)
		if err != nil {
			log.Printf("WARN: venue upsert %q: %v", v.Title, err)
			continue
		}
		venueMap[v.VenueID] = id
	}
	log.Printf("✓ upserted %d venues", len(venueMap))

	// ── Phase 2: load events ─────────────────────────────────────────────────
	eventsPath := filepath.Join(*dir, "all_events.json")
	allEvents, err := loadEvents(eventsPath)
	if err != nil {
		log.Fatalf("load events: %v", err)
	}
	log.Printf("loaded %d events from %s", len(allEvents), eventsPath)

	// Group by category and apply per-category limit
	byCat := make(map[string][]afishaEvent)
	for _, ev := range allEvents {
		byCat[ev.Category] = append(byCat[ev.Category], ev)
	}
	var selected []afishaEvent
	for cat, evs := range byCat {
		n := len(evs)
		if *limit > 0 && n > *limit {
			n = *limit
		}
		log.Printf("  category %-15s — using %d / %d events", cat, n, len(evs))
		selected = append(selected, evs[:n]...)
	}
	log.Printf("selected %d events total", len(selected))

	// ── Phase 3: compute date offset ─────────────────────────────────────────
	minDate := findMinOccurrenceDate(selected)
	tomorrow := time.Now().In(moscowTZ()).Truncate(24 * time.Hour).Add(24 * time.Hour)
	offsetDays := int(math.Ceil(tomorrow.Sub(minDate).Hours() / 24))
	if offsetDays < 0 {
		offsetDays = 0
	}
	log.Printf("date offset: +%d days (min occurrence: %s → tomorrow: %s)",
		offsetDays, minDate.Format("2006-01-02"), tomorrow.Format("2006-01-02"))

	// ── Phase 4: upsert events + sessions ────────────────────────────────────
	var importedEvents, importedSessions, skippedEvents int
	for _, ev := range selected {
		if ev.Title == "" {
			skippedEvents++
			continue
		}

		// Resolve venue from pre-loaded map; fall back to inline event data
		venueID, err := resolveOrUpsertVenue(ctx, pool, ev, venueMap)
		if err != nil {
			log.Printf("WARN: venue for %q: %v", ev.Title, err)
			skippedEvents++
			continue
		}

		dbEventID, err := upsertEvent(ctx, pool, ev, venueID, offsetDays)
		if err != nil {
			log.Printf("WARN: event upsert %q: %v", ev.Title, err)
			skippedEvents++
			continue
		}
		importedEvents++

		// Create sessions from occurrences
		n, err := upsertSessions(ctx, pool, ev, dbEventID, offsetDays, *maxOcc)
		if err != nil {
			log.Printf("WARN: sessions for %q: %v", ev.Title, err)
		}
		importedSessions += n
	}

	log.Printf("\n✓ done — events: %d imported, %d skipped | sessions: %d created",
		importedEvents, skippedEvents, importedSessions)
}

// ─── Loaders ─────────────────────────────────────────────────────────────────

func loadVenues(path string) ([]afishaVenue, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer f.Close()
	var venues []afishaVenue
	if err := json.NewDecoder(f).Decode(&venues); err != nil {
		return nil, err
	}
	return venues, nil
}

func loadEvents(path string) ([]afishaEvent, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer f.Close()
	var events []afishaEvent
	if err := json.NewDecoder(f).Decode(&events); err != nil {
		return nil, err
	}
	return events, nil
}

// ─── Date helpers ─────────────────────────────────────────────────────────────

func moscowTZ() *time.Location {
	loc, err := time.LoadLocation("Europe/Moscow")
	if err != nil {
		return time.FixedZone("MSK", 3*3600)
	}
	return loc
}

func parseDate(dateStr string) time.Time {
	if dateStr == "" {
		return time.Time{}
	}
	t, err := time.ParseInLocation("2006-01-02", dateStr, moscowTZ())
	if err != nil {
		return time.Time{}
	}
	return t
}

func shiftDate(dateStr string, offsetDays int) *time.Time {
	d := parseDate(dateStr)
	if d.IsZero() {
		return nil
	}
	t := d.Add(time.Duration(offsetDays) * 24 * time.Hour)
	return &t
}

func findMinOccurrenceDate(events []afishaEvent) time.Time {
	var min time.Time
	for _, ev := range events {
		// Check event-level date
		if d := parseDate(ev.Date); !d.IsZero() {
			if min.IsZero() || d.Before(min) {
				min = d
			}
		}
		// Also scan occurrences
		for _, occ := range ev.Occurrences {
			if d := parseDate(occ.Date); !d.IsZero() {
				if min.IsZero() || d.Before(min) {
					min = d
				}
			}
		}
	}
	if min.IsZero() {
		min = time.Now().In(moscowTZ())
	}
	return min
}

// ─── String helpers ───────────────────────────────────────────────────────────

func normalizeTitle(s string) string {
	var b strings.Builder
	for _, r := range strings.ToLower(s) {
		if unicode.IsLetter(r) || unicode.IsDigit(r) {
			b.WriteRune(r)
		} else {
			b.WriteRune(' ')
		}
	}
	return strings.Join(strings.Fields(b.String()), " ")
}

func posterURL(photo, photoHighres string) string {
	if photoHighres != "" {
		return photoHighres
	}
	return photo
}

func nullableString(s string) *string {
	if s == "" {
		return nil
	}
	return &s
}

func nullableFloat(f *float64) *float64 {
	return f
}

func nullableInt(n int) *int {
	if n == 0 {
		return nil
	}
	return &n
}

func stringSliceToPostgres(ss []string) interface{} {
	if len(ss) == 0 {
		return nil
	}
	return ss
}

// ─── Venue upsert (from venues.json) ─────────────────────────────────────────

func upsertVenueFull(ctx context.Context, pool *pgxpool.Pool, v afishaVenue) (uuid.UUID, error) {
	city := "moscow"
	name := v.Title
	if name == "" {
		name = "Уточняется"
	}

	var id uuid.UUID
	err := pool.QueryRow(ctx, `
		INSERT INTO venues (
			id, name, address, city, latitude, longitude,
			external_source, external_id, source_url,
			metro_stations, venue_type_code, venue_type_name,
			created_at, updated_at
		)
		VALUES ($1, $2, $3, $4, $5, $6, 'yandex_afisha', $7, $8,
		        $9, $10, $11, now(), now())
		ON CONFLICT (external_source, external_id)
		WHERE external_source IS NOT NULL AND external_id IS NOT NULL
		DO UPDATE SET
			name           = EXCLUDED.name,
			address        = EXCLUDED.address,
			latitude       = EXCLUDED.latitude,
			longitude      = EXCLUDED.longitude,
			source_url     = EXCLUDED.source_url,
			metro_stations = EXCLUDED.metro_stations,
			venue_type_code= EXCLUDED.venue_type_code,
			venue_type_name= EXCLUDED.venue_type_name,
			updated_at     = now()
		RETURNING id
	`,
		uuid.New(), name, v.Address, city, v.Latitude, v.Longitude,
		v.VenueID, nullableString(v.URL),
		stringSliceToPostgres(v.MetroStations),
		nullableString(v.TypeCode),
		nullableString(v.TypeName),
	).Scan(&id)
	if err != nil {
		// Fallback: look up by (name, city) if conflict index wasn't hit
		err2 := pool.QueryRow(ctx, `
			SELECT id FROM venues WHERE external_source = 'yandex_afisha' AND external_id = $1 LIMIT 1
		`, v.VenueID).Scan(&id)
		if err2 != nil {
			return uuid.Nil, fmt.Errorf("venue lookup: %w (original: %v)", err2, err)
		}
	}
	return id, nil
}

// resolveOrUpsertVenue resolves the event's venue_id from the pre-loaded map;
// if not found, falls back to upserting using the inline event venue fields.
func resolveOrUpsertVenue(ctx context.Context, pool *pgxpool.Pool, ev afishaEvent, venueMap map[string]uuid.UUID) (uuid.UUID, error) {
	if ev.VenueID != "" {
		if id, ok := venueMap[ev.VenueID]; ok {
			return id, nil
		}
	}

	// Fallback: create a minimal venue from event's inline fields
	name := ev.Location
	if name == "" {
		name = "Уточняется"
	}
	var id uuid.UUID
	err := pool.QueryRow(ctx, `
		INSERT INTO venues (
			id, name, address, city, latitude, longitude,
			external_source, external_id,
			metro_stations, venue_type_code, venue_type_name,
			created_at, updated_at
		)
		VALUES ($1, $2, $3, $4, 0, 0, 'yandex_afisha', $5,
		        $6, $7, $8, now(), now())
		ON CONFLICT (external_source, external_id)
		WHERE external_source IS NOT NULL AND external_id IS NOT NULL
		DO UPDATE SET updated_at = now()
		RETURNING id
	`,
		uuid.New(), name, ev.VenueAddress, "moscow",
		nullableString(ev.VenueID),
		stringSliceToPostgres(ev.MetroStations),
		nullableString(ev.VenueTypeCode),
		nullableString(ev.VenueTypeName),
	).Scan(&id)
	if err != nil {
		// Last resort: find by name+city
		err2 := pool.QueryRow(ctx, `
			SELECT id FROM venues WHERE name = $1 AND city = 'moscow' LIMIT 1
		`, name).Scan(&id)
		if err2 != nil {
			return uuid.Nil, fmt.Errorf("venue fallback lookup: %w", err2)
		}
	}
	return id, nil
}

// ─── Event upsert ─────────────────────────────────────────────────────────────

func upsertEvent(ctx context.Context, pool *pgxpool.Pool, ev afishaEvent, venueID uuid.UUID, offsetDays int) (uuid.UUID, error) {
	startsAt := shiftDate(ev.Date, offsetDays)

	// ends_at: add 1 day so single-day events don't expire at midnight
	var endsAt *time.Time
	if startsAt != nil {
		t := startsAt.Add(23*time.Hour + 59*time.Minute + 59*time.Second)
		endsAt = &t
	}

	normTitle := normalizeTitle(ev.Title)
	dedupeKey := fmt.Sprintf("%s|%s", normTitle, ev.Date)

	externalID := ev.URL
	if externalID == "" {
		externalID = ev.EventID
	}
	if externalID == "" {
		externalID = dedupeKey
	}

	poster := posterURL(ev.Photo, ev.PhotoHighres)
	category := ev.Category
	if category == "" {
		category = ev.SourceSlug
	}

	// Prefer long_description, fall back to short
	description := ev.LongDescription
	if description == "" {
		description = ev.ShortDescription
	}

	var ageRestriction *string
	if ev.AgeRestriction != "" {
		ageRestriction = &ev.AgeRestriction
	}

	var ratingCount *int
	if ev.RatingCount > 0 {
		ratingCount = &ev.RatingCount
	}

	var tags interface{}
	if len(ev.Tags) > 0 {
		tags = ev.Tags
	}

	var dbID uuid.UUID
	err := pool.QueryRow(ctx, `
		INSERT INTO events (
			id, title, description, long_description, category,
			poster_url, status, source,
			external_source, external_id, source_url,
			booking_mode,
			age_restriction, price_min, tags, rating_count,
			starts_at, ends_at,
			normalized_title, dedupe_key, venue_id,
			created_at, updated_at
		) VALUES (
			$1, $2, $3, $4, $5,
			$6, 'published', 'yandex_afisha',
			'yandex_afisha', $7, $8,
			'reserveflow_managed',
			$9, $10, $11, $12,
			$13, $14,
			$15, $16, $17,
			now(), now()
		)
		ON CONFLICT (external_source, external_id)
		WHERE external_source IS NOT NULL
		DO UPDATE SET
			poster_url      = EXCLUDED.poster_url,
			description     = EXCLUDED.description,
			long_description= EXCLUDED.long_description,
			age_restriction = EXCLUDED.age_restriction,
			price_min       = EXCLUDED.price_min,
			tags            = EXCLUDED.tags,
			rating_count    = EXCLUDED.rating_count,
			starts_at       = EXCLUDED.starts_at,
			ends_at         = EXCLUDED.ends_at,
			venue_id        = EXCLUDED.venue_id,
			updated_at      = now()
		RETURNING id
	`,
		uuid.New(), ev.Title, nullableString(description), nullableString(ev.LongDescription), category,
		nullableString(poster),
		externalID, nullableString(ev.URL),
		ageRestriction, nullableFloat(ev.PriceMin), tags, ratingCount,
		startsAt, endsAt,
		normTitle, dedupeKey, venueID,
	).Scan(&dbID)
	if err != nil {
		// ON CONFLICT DO UPDATE always returns a row, so this shouldn't happen.
		// Try to look up if something went wrong.
		err2 := pool.QueryRow(ctx, `
			SELECT id FROM events
			WHERE external_source = 'yandex_afisha' AND external_id = $1
		`, externalID).Scan(&dbID)
		if err2 != nil {
			return uuid.Nil, fmt.Errorf("event upsert: %w (lookup: %v)", err, err2)
		}
	}
	return dbID, nil
}

// ─── Session upsert ───────────────────────────────────────────────────────────

func upsertSessions(
	ctx context.Context, pool *pgxpool.Pool,
	ev afishaEvent, eventID uuid.UUID,
	offsetDays, maxOcc int,
) (int, error) {
	now := time.Now()
	count := 0

	for i, occ := range ev.Occurrences {
		if maxOcc > 0 && i >= maxOcc {
			break
		}

		shifted := shiftDate(occ.Date, offsetDays)
		if shifted == nil {
			continue
		}

		// Only create future sessions (after shift they should all be future,
		// but guard against edge cases).
		if shifted.Before(now) {
			continue
		}

		// external_id = event_id + "|" + original_date (stable, unique per occurrence)
		externalID := fmt.Sprintf("%s|%s", ev.EventID, occ.Date)
		if ev.EventID == "" {
			externalID = fmt.Sprintf("%s|%s", ev.URL, occ.Date)
		}

		// ends_at: starts_at + duration if available, else +3 hours
		endsAt := shifted.Add(3 * time.Hour)
		if ev.DurationMinutes != nil && *ev.DurationMinutes > 0 {
			endsAt = shifted.Add(time.Duration(*ev.DurationMinutes) * time.Minute)
		}

		bookingURL := occ.BookingURL
		if bookingURL == "" {
			bookingURL = ev.URL
		}

		_, err := pool.Exec(ctx, `
			INSERT INTO sessions (
				id, event_id, hall_id,
				starts_at, ends_at, status,
				external_source, external_id, source_url,
				is_bookable,
				created_at, updated_at
			) VALUES (
				$1, $2, NULL,
				$3, $4, 'scheduled',
				'yandex_afisha', $5, $6,
				true,
				now(), now()
			)
			ON CONFLICT (external_source, external_id)
			WHERE external_source IS NOT NULL AND external_id IS NOT NULL
			DO UPDATE SET
				starts_at  = EXCLUDED.starts_at,
				ends_at    = EXCLUDED.ends_at,
				source_url = EXCLUDED.source_url,
				updated_at = now()
		`,
			uuid.New(), eventID,
			shifted, endsAt,
			externalID, nullableString(bookingURL),
		)
		if err != nil {
			return count, fmt.Errorf("session %s: %w", externalID, err)
		}
		count++
	}
	return count, nil
}
