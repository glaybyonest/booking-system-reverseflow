DELETE FROM event_external_links;
DROP TABLE IF EXISTS event_external_links;
DROP TABLE IF EXISTS event_import_runs;

DROP INDEX IF EXISTS idx_sessions_external_source_external_id_unique;
DROP INDEX IF EXISTS idx_venues_external_source_external_id_unique;
DROP INDEX IF EXISTS idx_event_import_runs_provider_started_at;
DROP INDEX IF EXISTS idx_venues_latitude_longitude;
DROP INDEX IF EXISTS idx_events_dedupe_key;
DROP INDEX IF EXISTS idx_events_source;
DROP INDEX IF EXISTS idx_events_booking_mode;
DROP INDEX IF EXISTS idx_events_ends_at;
DROP INDEX IF EXISTS idx_events_starts_at;
DROP INDEX IF EXISTS idx_events_external_source_external_id_unique;

DELETE FROM sessions WHERE hall_id IS NULL;

ALTER TABLE events DROP CONSTRAINT IF EXISTS fk_events_venue;

ALTER TABLE sessions
    DROP COLUMN IF EXISTS is_bookable,
    DROP COLUMN IF EXISTS seatsio_event_key,
    DROP COLUMN IF EXISTS ticketmaster_seatmap_static_url,
    DROP COLUMN IF EXISTS source_url,
    DROP COLUMN IF EXISTS external_id,
    DROP COLUMN IF EXISTS external_source;

ALTER TABLE sessions
    ALTER COLUMN hall_id SET NOT NULL;

ALTER TABLE venues
    DROP COLUMN IF EXISTS raw_payload,
    DROP COLUMN IF EXISTS seatsio_chart_key,
    DROP COLUMN IF EXISTS seat_map_provider,
    DROP COLUMN IF EXISTS source_url,
    DROP COLUMN IF EXISTS longitude,
    DROP COLUMN IF EXISTS latitude,
    DROP COLUMN IF EXISTS external_id,
    DROP COLUMN IF EXISTS external_source;

ALTER TABLE events
    DROP COLUMN IF EXISTS venue_id,
    DROP COLUMN IF EXISTS dedupe_key,
    DROP COLUMN IF EXISTS normalized_title,
    DROP COLUMN IF EXISTS ends_at,
    DROP COLUMN IF EXISTS starts_at,
    DROP COLUMN IF EXISTS booking_mode,
    DROP COLUMN IF EXISTS raw_payload,
    DROP COLUMN IF EXISTS last_synced_at,
    DROP COLUMN IF EXISTS imported_at,
    DROP COLUMN IF EXISTS source_url,
    DROP COLUMN IF EXISTS external_id,
    DROP COLUMN IF EXISTS external_source,
    DROP COLUMN IF EXISTS source;
