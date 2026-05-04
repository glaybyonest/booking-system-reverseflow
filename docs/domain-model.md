# Domain Model

## Сущности

- `users`: зарегистрированные пользователи с bcrypt-хешем пароля и ролью
- `events`: публичный каталог событий
- `venues`: площадки
- `halls`: залы внутри площадок
- `sessions`: конкретные сеансы событий в залах
- `seats`: физические места в зале
- `session_seats`: состояние места в рамках конкретного сеанса
- `bookings`: агрегат пользовательской брони
- `booking_items`: позиции брони по местам (в MVP — одно место на бронь)
- `payments`: записи mock-платежей с опциональным `idempotencyKey`
- `refresh_tokens`: хешированные refresh-токены с метаданными revoke
- `notifications`: история пользовательских уведомлений
- `outbox_events`: гарантированно сохранённые доменные события до публикации в Kafka
- `processed_events`: таблица идемпотентности consumers

## Статусы

Booking:

- `pending`
- `confirmed`
- `cancelled`
- `expired`
- `payment_failed`

Session seat:

- `available`
- `held`
- `booked`

Payment:

- `pending`
- `succeeded`
- `failed`

Event:

- `draft`
- `published`
- `cancelled`
- `archived`

Session:

- `scheduled`
- `cancelled`
- `finished`

Notification:

- `booking_confirmed`
- `booking_expired`
- `payment_failed`
- `booking_cancelled`

## Почему `session_seats`

`seats` — это физические места, общие для зала. Одно и то же место может быть свободно в одном сеансе и занято в другом, поэтому изменяемое состояние доступности не должно храниться в `seats`.

`session_seats` — это инвентарь мест на уровне конкретного сеанса и именно эта строка блокируется при hold/payment переходах.
