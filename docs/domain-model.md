# Domain Model

## Entities

- `users`: registered users with bcrypt password hashes and roles.
- `events`: public event catalog records.
- `venues`: physical venue.
- `halls`: rooms inside venues.
- `sessions`: scheduled event occurrence in a hall.
- `seats`: physical seats in a hall.
- `session_seats`: per-session seat state.
- `bookings`: user reservation aggregate.
- `booking_items`: seat-level booking item. MVP supports one seat per booking.
- `payments`: mock payment records with optional idempotency key.
- `refresh_tokens`: hashed refresh tokens with revoke metadata.
- `notifications`: user-visible notification history.
- `outbox_events`: durable domain events waiting for Kafka publish.
- `processed_events`: idempotency table for consumers.

## Statuses

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

## Why `session_seats`

`seats` are physical seats in a hall. A seat can be available for one session and booked for another, so mutable availability must not live on `seats`. `session_seats` is the per-session inventory row and the row locked during hold/payment transitions.
