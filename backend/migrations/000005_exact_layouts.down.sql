DROP INDEX IF EXISTS idx_session_seats_session_id_is_active;
DROP INDEX IF EXISTS idx_seats_hall_id_layout_key_unique;

ALTER TABLE session_seats
    DROP COLUMN IF EXISTS is_active;

ALTER TABLE seats
    DROP COLUMN IF EXISTS layout_key;

ALTER TABLE sessions
    DROP COLUMN IF EXISTS layout_json;

ALTER TABLE halls
    DROP COLUMN IF EXISTS layout_json;
