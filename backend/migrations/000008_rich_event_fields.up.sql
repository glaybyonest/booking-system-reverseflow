-- 000008: rich event fields — age restriction, pricing, extended description,
-- tags, rating, and venue metro / type metadata.

ALTER TABLE events
    ADD COLUMN IF NOT EXISTS age_restriction VARCHAR(16)  NULL,
    ADD COLUMN IF NOT EXISTS price_min        NUMERIC(10,2) NULL,
    ADD COLUMN IF NOT EXISTS price_max        NUMERIC(10,2) NULL,
    ADD COLUMN IF NOT EXISTS long_description TEXT         NULL,
    ADD COLUMN IF NOT EXISTS tags             TEXT[]       NULL,
    ADD COLUMN IF NOT EXISTS rating_count     INT          NULL;

ALTER TABLE venues
    ADD COLUMN IF NOT EXISTS metro_stations  TEXT[]        NULL,
    ADD COLUMN IF NOT EXISTS venue_type_code VARCHAR(64)   NULL,
    ADD COLUMN IF NOT EXISTS venue_type_name VARCHAR(128)  NULL;

CREATE INDEX IF NOT EXISTS idx_events_category        ON events (category);
CREATE INDEX IF NOT EXISTS idx_events_price_min       ON events (price_min);
CREATE INDEX IF NOT EXISTS idx_events_age_restriction ON events (age_restriction);
CREATE INDEX IF NOT EXISTS idx_venues_venue_type_code ON venues (venue_type_code);
