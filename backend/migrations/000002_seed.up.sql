INSERT INTO users (id, email, password_hash, name, role, created_at, updated_at)
VALUES (
    '10000000-0000-0000-0000-000000000001',
    'demo@example.com',
    '$2a$10$o5VAvU9pBmagDcUdtsfUROJV1fgAJth2MGCirwjmYqXfWoYfxA.B6',
    'Demo User',
    'user',
    now(),
    now()
)
ON CONFLICT (email) DO NOTHING;

INSERT INTO venues (id, name, address, city, created_at, updated_at)
VALUES ('20000000-0000-0000-0000-000000000001', 'ReserveFlow Center', '1 Booking Avenue', 'Moscow', now(), now())
ON CONFLICT (id) DO NOTHING;

INSERT INTO halls (id, venue_id, name, rows_count, seats_per_row, created_at, updated_at)
VALUES
    ('30000000-0000-0000-0000-000000000001', '20000000-0000-0000-0000-000000000001', 'Main Hall', 4, 10, now(), now()),
    ('30000000-0000-0000-0000-000000000002', '20000000-0000-0000-0000-000000000001', 'Blue Hall', 4, 10, now(), now())
ON CONFLICT (id) DO NOTHING;

INSERT INTO events (id, title, description, category, poster_url, status, created_at, updated_at)
VALUES
    ('40000000-0000-0000-0000-000000000001', 'Jazz Night', 'Live jazz evening with local musicians.', 'music', NULL, 'published', now(), now()),
    ('40000000-0000-0000-0000-000000000002', 'Backend Architecture Meetup', 'Talks about distributed systems and pragmatic monoliths.', 'tech', NULL, 'published', now(), now()),
    ('40000000-0000-0000-0000-000000000003', 'Modern Theatre', 'A small-stage theatre performance.', 'theatre', NULL, 'published', now(), now())
ON CONFLICT (id) DO NOTHING;

INSERT INTO sessions (id, event_id, hall_id, starts_at, ends_at, status, created_at, updated_at)
VALUES
    ('50000000-0000-0000-0000-000000000001', '40000000-0000-0000-0000-000000000001', '30000000-0000-0000-0000-000000000001', now() + interval '2 days', now() + interval '2 days 2 hours', 'scheduled', now(), now()),
    ('50000000-0000-0000-0000-000000000002', '40000000-0000-0000-0000-000000000001', '30000000-0000-0000-0000-000000000002', now() + interval '3 days', now() + interval '3 days 2 hours', 'scheduled', now(), now()),
    ('50000000-0000-0000-0000-000000000003', '40000000-0000-0000-0000-000000000002', '30000000-0000-0000-0000-000000000001', now() + interval '4 days', now() + interval '4 days 3 hours', 'scheduled', now(), now()),
    ('50000000-0000-0000-0000-000000000004', '40000000-0000-0000-0000-000000000003', '30000000-0000-0000-0000-000000000002', now() + interval '5 days', now() + interval '5 days 2 hours', 'scheduled', now(), now())
ON CONFLICT (id) DO NOTHING;

WITH rows(row_label) AS (
    VALUES ('A'), ('B'), ('C'), ('D')
),
numbers(seat_number) AS (
    SELECT generate_series(1, 10)
),
h AS (
    SELECT id AS hall_id FROM halls WHERE id IN ('30000000-0000-0000-0000-000000000001', '30000000-0000-0000-0000-000000000002')
)
INSERT INTO seats (id, hall_id, row_label, seat_number, seat_type, base_price, created_at, updated_at)
SELECT
    md5('seat-' || h.hall_id::text || '-' || rows.row_label || '-' || numbers.seat_number::text)::uuid,
    h.hall_id,
    rows.row_label,
    numbers.seat_number,
    'standard',
    CASE rows.row_label WHEN 'A' THEN 700 WHEN 'B' THEN 600 ELSE 500 END,
    now(),
    now()
FROM h
CROSS JOIN rows
CROSS JOIN numbers
ON CONFLICT (hall_id, row_label, seat_number) DO NOTHING;

INSERT INTO session_seats (id, session_id, seat_id, status, hold_expires_at, version, updated_at)
SELECT
    md5('session-seat-' || sessions.id::text || '-' || seats.id::text)::uuid,
    sessions.id,
    seats.id,
    'available',
    NULL,
    1,
    now()
FROM sessions
JOIN seats ON seats.hall_id = sessions.hall_id
ON CONFLICT (session_id, seat_id) DO NOTHING;
