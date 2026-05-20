ALTER TABLE events
    ADD COLUMN IF NOT EXISTS source VARCHAR NOT NULL DEFAULT 'manual',
    ADD COLUMN IF NOT EXISTS external_source VARCHAR NULL,
    ADD COLUMN IF NOT EXISTS external_id VARCHAR NULL,
    ADD COLUMN IF NOT EXISTS source_url TEXT NULL,
    ADD COLUMN IF NOT EXISTS imported_at TIMESTAMPTZ NULL,
    ADD COLUMN IF NOT EXISTS last_synced_at TIMESTAMPTZ NULL,
    ADD COLUMN IF NOT EXISTS raw_payload JSONB NULL,
    ADD COLUMN IF NOT EXISTS booking_mode VARCHAR NOT NULL DEFAULT 'reserveflow_managed',
    ADD COLUMN IF NOT EXISTS starts_at TIMESTAMPTZ NULL,
    ADD COLUMN IF NOT EXISTS ends_at TIMESTAMPTZ NULL,
    ADD COLUMN IF NOT EXISTS normalized_title VARCHAR NULL,
    ADD COLUMN IF NOT EXISTS dedupe_key VARCHAR NULL,
    ADD COLUMN IF NOT EXISTS venue_id UUID NULL;

DO $$
BEGIN
    IF NOT EXISTS (
        SELECT 1
        FROM pg_constraint
        WHERE conname = 'fk_events_venue'
    ) THEN
        ALTER TABLE events
            ADD CONSTRAINT fk_events_venue
            FOREIGN KEY (venue_id) REFERENCES venues(id);
    END IF;
END
$$;

ALTER TABLE venues
    ADD COLUMN IF NOT EXISTS external_source VARCHAR NULL,
    ADD COLUMN IF NOT EXISTS external_id VARCHAR NULL,
    ADD COLUMN IF NOT EXISTS latitude DOUBLE PRECISION NULL,
    ADD COLUMN IF NOT EXISTS longitude DOUBLE PRECISION NULL,
    ADD COLUMN IF NOT EXISTS source_url TEXT NULL,
    ADD COLUMN IF NOT EXISTS seat_map_provider VARCHAR NOT NULL DEFAULT 'internal_grid',
    ADD COLUMN IF NOT EXISTS seatsio_chart_key VARCHAR NULL,
    ADD COLUMN IF NOT EXISTS raw_payload JSONB NULL;

ALTER TABLE sessions
    ALTER COLUMN hall_id DROP NOT NULL;

ALTER TABLE sessions
    ADD COLUMN IF NOT EXISTS external_source VARCHAR NULL,
    ADD COLUMN IF NOT EXISTS external_id VARCHAR NULL,
    ADD COLUMN IF NOT EXISTS source_url TEXT NULL,
    ADD COLUMN IF NOT EXISTS ticketmaster_seatmap_static_url TEXT NULL,
    ADD COLUMN IF NOT EXISTS seatsio_event_key VARCHAR NULL,
    ADD COLUMN IF NOT EXISTS is_bookable BOOLEAN NOT NULL DEFAULT TRUE;

CREATE TABLE IF NOT EXISTS event_import_runs (
    id UUID PRIMARY KEY,
    provider VARCHAR NOT NULL,
    city VARCHAR NULL,
    status VARCHAR NOT NULL,
    started_at TIMESTAMPTZ NOT NULL,
    finished_at TIMESTAMPTZ NULL,
    fetched_count INT NOT NULL DEFAULT 0,
    imported_count INT NOT NULL DEFAULT 0,
    updated_count INT NOT NULL DEFAULT 0,
    skipped_count INT NOT NULL DEFAULT 0,
    duplicate_count INT NOT NULL DEFAULT 0,
    page_count INT NOT NULL DEFAULT 0,
    error_message TEXT NULL
);

CREATE TABLE IF NOT EXISTS event_external_links (
    id UUID PRIMARY KEY,
    event_id UUID NOT NULL REFERENCES events(id),
    external_source VARCHAR NOT NULL,
    external_id VARCHAR NOT NULL,
    source_url TEXT NULL,
    raw_payload JSONB NULL,
    imported_at TIMESTAMPTZ NOT NULL,
    UNIQUE (external_source, external_id)
);

CREATE UNIQUE INDEX IF NOT EXISTS idx_events_external_source_external_id_unique
    ON events (external_source, external_id)
    WHERE external_source IS NOT NULL;

CREATE INDEX IF NOT EXISTS idx_events_starts_at ON events (starts_at);
CREATE INDEX IF NOT EXISTS idx_events_ends_at ON events (ends_at);
CREATE INDEX IF NOT EXISTS idx_events_booking_mode ON events (booking_mode);
CREATE INDEX IF NOT EXISTS idx_events_source ON events (source);
CREATE INDEX IF NOT EXISTS idx_events_dedupe_key ON events (dedupe_key);
CREATE INDEX IF NOT EXISTS idx_venues_latitude_longitude ON venues (latitude, longitude);
CREATE INDEX IF NOT EXISTS idx_event_import_runs_provider_started_at ON event_import_runs (provider, started_at DESC);

CREATE UNIQUE INDEX IF NOT EXISTS idx_venues_external_source_external_id_unique
    ON venues (external_source, external_id)
    WHERE external_source IS NOT NULL AND external_id IS NOT NULL;

CREATE UNIQUE INDEX IF NOT EXISTS idx_sessions_external_source_external_id_unique
    ON sessions (external_source, external_id)
    WHERE external_source IS NOT NULL AND external_id IS NOT NULL;

UPDATE venues
SET seat_map_provider = 'internal_grid'
WHERE seat_map_provider IS NULL OR seat_map_provider = '';

UPDATE sessions
SET is_bookable = TRUE
WHERE is_bookable IS DISTINCT FROM TRUE;

WITH event_rollup AS (
    SELECT
        s.event_id,
        MIN(s.starts_at) AS starts_at,
        MAX(s.ends_at) AS ends_at,
        MIN(v.id::text) AS venue_id
    FROM sessions s
    LEFT JOIN halls h ON h.id = s.hall_id
    LEFT JOIN venues v ON v.id = h.venue_id
    GROUP BY s.event_id
)
UPDATE events e
SET
    source = COALESCE(NULLIF(e.source, ''), 'manual'),
    booking_mode = COALESCE(NULLIF(e.booking_mode, ''), 'reserveflow_managed'),
    starts_at = COALESCE(e.starts_at, er.starts_at),
    ends_at = COALESCE(e.ends_at, er.ends_at),
    venue_id = COALESCE(e.venue_id, er.venue_id::uuid),
    normalized_title = COALESCE(e.normalized_title, LOWER(e.title))
FROM event_rollup er
WHERE e.id = er.event_id;
