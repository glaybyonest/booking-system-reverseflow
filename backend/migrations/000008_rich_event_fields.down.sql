DROP INDEX IF EXISTS idx_venues_venue_type_code;
DROP INDEX IF EXISTS idx_events_age_restriction;
DROP INDEX IF EXISTS idx_events_price_min;
DROP INDEX IF EXISTS idx_events_category;

ALTER TABLE venues
    DROP COLUMN IF EXISTS metro_stations,
    DROP COLUMN IF EXISTS venue_type_code,
    DROP COLUMN IF EXISTS venue_type_name;

ALTER TABLE events
    DROP COLUMN IF EXISTS age_restriction,
    DROP COLUMN IF EXISTS price_min,
    DROP COLUMN IF EXISTS price_max,
    DROP COLUMN IF EXISTS long_description,
    DROP COLUMN IF EXISTS tags,
    DROP COLUMN IF EXISTS rating_count;
