-- Seed real Moscow venues with GPS coordinates so the map page shows content
-- and event-detail mini-maps render even without an active KudaGo/TimePad sync.

INSERT INTO venues (id, name, address, city, latitude, longitude, created_at, updated_at)
VALUES
    ('70000000-0000-0000-0000-000000000001', 'Большой театр',            'Театральная пл., 1',      'Москва', 55.760255, 37.618697, now(), now()),
    ('70000000-0000-0000-0000-000000000002', 'Крокус Сити Холл',         '65–66 км МКАД',           'Москва', 55.825540, 37.375026, now(), now()),
    ('70000000-0000-0000-0000-000000000003', 'Стадион Лужники',          'Лужнецкая наб., 24',      'Москва', 55.716456, 37.555099, now(), now()),
    ('70000000-0000-0000-0000-000000000004', 'Цирк на Цветном бульваре', 'Цветной б-р, 13',         'Москва', 55.772869, 37.616497, now(), now()),
    ('70000000-0000-0000-0000-000000000005', 'Клуб Москва',              'Охотный Ряд, пр. 2',      'Москва', 55.756840, 37.613255, now(), now())
ON CONFLICT (id) DO NOTHING;

-- Demo events with external_link_only booking mode (no sessions required)
-- so they appear in the event list and on the map immediately after migration.
INSERT INTO events (
    id, title, description, category, status, source, booking_mode,
    venue_id, starts_at, ends_at, created_at, updated_at
)
VALUES
    (
        '80000000-0000-0000-0000-000000000001',
        'Лебединое озеро — Большой театр',
        'Великий балет Чайковского в исполнении труппы Государственного академического Большого театра России.',
        'Театр',
        'published', 'manual', 'external_link_only',
        '70000000-0000-0000-0000-000000000001',
        now() + interval '7 days',
        now() + interval '7 days'  + interval '3 hours',
        now(), now()
    ),
    (
        '80000000-0000-0000-0000-000000000002',
        'Международный джазовый фестиваль',
        'Ежегодный фестиваль с участием мировых звёзд джаза. Три дня, три сцены, более 30 артистов.',
        'Фестиваль',
        'published', 'manual', 'external_link_only',
        '70000000-0000-0000-0000-000000000002',
        now() + interval '14 days',
        now() + interval '14 days' + interval '6 hours',
        now(), now()
    ),
    (
        '80000000-0000-0000-0000-000000000003',
        'Симфония под открытым небом',
        'Грандиозное летнее шоу с симфоническим оркестром, лазерным шоу и живым пиротехническим аккомпанементом.',
        'Концерт',
        'published', 'manual', 'external_link_only',
        '70000000-0000-0000-0000-000000000003',
        now() + interval '21 days',
        now() + interval '21 days' + interval '4 hours',
        now(), now()
    ),
    (
        '80000000-0000-0000-0000-000000000004',
        'Цирковое шоу «Магия»',
        'Захватывающее цирковое представление с акробатами, иллюзионистами и дрессированными животными.',
        'Развлечения',
        'published', 'manual', 'external_link_only',
        '70000000-0000-0000-0000-000000000004',
        now() + interval '10 days',
        now() + interval '10 days' + interval '2 hours',
        now(), now()
    )
ON CONFLICT (id) DO NOTHING;
