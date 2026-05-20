INSERT INTO venues (id, name, address, city, latitude, longitude, seat_map_provider, created_at, updated_at)
VALUES (
    '20000000-0000-0000-0000-000000000001',
    'Seed Venue',
    'Seed Address',
    'Moscow',
    55.757780,
    37.615149,
    'internal_grid',
    now(),
    now()
)
ON CONFLICT (id) DO NOTHING;

INSERT INTO halls (id, venue_id, name, rows_count, seats_per_row, created_at, updated_at)
VALUES
    ('30000000-0000-0000-0000-000000000001', '20000000-0000-0000-0000-000000000001', 'Seed Hall 1', 0, 0, now(), now()),
    ('30000000-0000-0000-0000-000000000002', '20000000-0000-0000-0000-000000000001', 'Seed Hall 2', 0, 0, now(), now())
ON CONFLICT (id) DO NOTHING;
