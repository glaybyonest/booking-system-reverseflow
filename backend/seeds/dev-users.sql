INSERT INTO users (id, email, password_hash, name, role, created_at, updated_at)
VALUES (
    '10000000-0000-0000-0000-000000000001',
    'demo@example.com',
    '$2a$10$o5VAvU9pBmagDcUdtsfUROJV1fgAJth2MGCirwjmYqXfWoYfxA.B6',
    'Демо-покупатель',
    'user',
    now(),
    now()
)
ON CONFLICT (email) DO UPDATE SET
    name = EXCLUDED.name,
    updated_at = now();

INSERT INTO users (id, email, password_hash, name, role, created_at, updated_at)
VALUES (
    '10000000-0000-0000-0000-000000000002',
    'admin@example.com',
    '$2a$10$o5VAvU9pBmagDcUdtsfUROJV1fgAJth2MGCirwjmYqXfWoYfxA.B6',
    'Администратор ReserveFlow',
    'admin',
    now(),
    now()
)
ON CONFLICT (email) DO UPDATE SET
    name = EXCLUDED.name,
    role = EXCLUDED.role,
    updated_at = now();
