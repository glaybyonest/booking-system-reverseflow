ALTER TABLE halls
    ADD COLUMN IF NOT EXISTS layout_json JSONB NULL;

ALTER TABLE sessions
    ADD COLUMN IF NOT EXISTS layout_json JSONB NULL;

ALTER TABLE seats
    ADD COLUMN IF NOT EXISTS layout_key VARCHAR NULL;

ALTER TABLE session_seats
    ADD COLUMN IF NOT EXISTS is_active BOOLEAN NOT NULL DEFAULT TRUE;

UPDATE seats
SET layout_key = row_label || '-' || seat_number::text
WHERE layout_key IS NULL OR layout_key = '';

CREATE UNIQUE INDEX IF NOT EXISTS idx_seats_hall_id_layout_key_unique
    ON seats (hall_id, layout_key)
    WHERE layout_key IS NOT NULL;

CREATE INDEX IF NOT EXISTS idx_session_seats_session_id_is_active
    ON session_seats (session_id, is_active);
