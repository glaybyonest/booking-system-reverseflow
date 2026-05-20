INSERT INTO venues (id, name, address, city, latitude, longitude, seat_map_provider, created_at, updated_at)
VALUES (
    '20000000-0000-0000-0000-000000000001',
    'Культурный центр «Север»',
    'ул. Тверская, 7',
    'Москва',
    55.757780,
    37.615149,
    'internal_grid',
    now(),
    now()
)
ON CONFLICT (id) DO UPDATE SET
    name = EXCLUDED.name,
    address = EXCLUDED.address,
    city = EXCLUDED.city,
    latitude = EXCLUDED.latitude,
    longitude = EXCLUDED.longitude,
    seat_map_provider = EXCLUDED.seat_map_provider,
    updated_at = now();

INSERT INTO halls (id, venue_id, name, rows_count, seats_per_row, created_at, updated_at)
VALUES
    ('30000000-0000-0000-0000-000000000001', '20000000-0000-0000-0000-000000000001', 'Большой зал', 4, 10, now(), now()),
    ('30000000-0000-0000-0000-000000000002', '20000000-0000-0000-0000-000000000001', 'Камерный зал', 4, 10, now(), now())
ON CONFLICT (id) DO UPDATE SET
    venue_id = EXCLUDED.venue_id,
    name = EXCLUDED.name,
    rows_count = EXCLUDED.rows_count,
    seats_per_row = EXCLUDED.seats_per_row,
    updated_at = now();

INSERT INTO events (
    id,
    title,
    description,
    category,
    poster_url,
    status,
    source,
    booking_mode,
    starts_at,
    ends_at,
    venue_id,
    created_at,
    updated_at
)
VALUES
    (
        '40000000-0000-0000-0000-000000000001',
        'Джаз на Новой сцене',
        'Вечер живого джаза с московскими музыкантами, камерной посадкой и быстрым выбором мест.',
        'Концерт',
        NULL,
        'published',
        'manual',
        'reserveflow_managed',
        now() + interval '2 days',
        now() + interval '2 days 2 hours',
        '20000000-0000-0000-0000-000000000001',
        now(),
        now()
    ),
    (
        '40000000-0000-0000-0000-000000000002',
        'Go в продакшене',
        'Практическая встреча для backend-команд: надежные монолиты, очереди, платежные сценарии и наблюдаемость.',
        'IT-конференция',
        NULL,
        'published',
        'manual',
        'reserveflow_managed',
        now() + interval '4 days',
        now() + interval '4 days 3 hours',
        '20000000-0000-0000-0000-000000000001',
        now(),
        now()
    ),
    (
        '40000000-0000-0000-0000-000000000003',
        'Современный театр: Город',
        'Спектакль малой формы о ритме большого города, личных маршрутах и случайных встречах.',
        'Театр',
        NULL,
        'published',
        'manual',
        'reserveflow_managed',
        now() + interval '5 days',
        now() + interval '5 days 2 hours',
        '20000000-0000-0000-0000-000000000001',
        now(),
        now()
    )
ON CONFLICT (id) DO UPDATE SET
    title = EXCLUDED.title,
    description = EXCLUDED.description,
    category = EXCLUDED.category,
    poster_url = EXCLUDED.poster_url,
    status = EXCLUDED.status,
    source = EXCLUDED.source,
    booking_mode = EXCLUDED.booking_mode,
    starts_at = EXCLUDED.starts_at,
    ends_at = EXCLUDED.ends_at,
    venue_id = EXCLUDED.venue_id,
    updated_at = now();

INSERT INTO sessions (id, event_id, hall_id, starts_at, ends_at, status, is_bookable, created_at, updated_at)
VALUES
    ('50000000-0000-0000-0000-000000000001', '40000000-0000-0000-0000-000000000001', '30000000-0000-0000-0000-000000000001', now() + interval '2 days', now() + interval '2 days 2 hours', 'scheduled', TRUE, now(), now()),
    ('50000000-0000-0000-0000-000000000002', '40000000-0000-0000-0000-000000000001', '30000000-0000-0000-0000-000000000002', now() + interval '3 days', now() + interval '3 days 2 hours', 'scheduled', TRUE, now(), now()),
    ('50000000-0000-0000-0000-000000000003', '40000000-0000-0000-0000-000000000002', '30000000-0000-0000-0000-000000000001', now() + interval '4 days', now() + interval '4 days 3 hours', 'scheduled', TRUE, now(), now()),
    ('50000000-0000-0000-0000-000000000004', '40000000-0000-0000-0000-000000000003', '30000000-0000-0000-0000-000000000002', now() + interval '5 days', now() + interval '5 days 2 hours', 'scheduled', TRUE, now(), now())
ON CONFLICT (id) DO UPDATE SET
    event_id = EXCLUDED.event_id,
    hall_id = EXCLUDED.hall_id,
    starts_at = EXCLUDED.starts_at,
    ends_at = EXCLUDED.ends_at,
    status = EXCLUDED.status,
    is_bookable = EXCLUDED.is_bookable,
    updated_at = now();

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
    CASE rows.row_label WHEN 'A' THEN 2500 WHEN 'B' THEN 1900 ELSE 1400 END,
    now(),
    now()
FROM h
CROSS JOIN rows
CROSS JOIN numbers
ON CONFLICT (hall_id, row_label, seat_number) DO UPDATE SET
    seat_type = EXCLUDED.seat_type,
    base_price = EXCLUDED.base_price,
    updated_at = now();

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
