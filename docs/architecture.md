# Architecture

ReserveFlow is a backend-only modular monolith. It is one Go deployable codebase with clear domain modules instead of distributed services in the MVP.

## Modules

- `auth`: users, bcrypt, JWT access tokens, hashed refresh tokens
- `events`: published event catalog
- `sessions`: scheduled event sessions and hall metadata
- `seats`: seat map read model and Redis cache
- `bookings`: hold, cancel, expire, confirm, payment-failed transitions
- `payments`: mock provider and idempotency
- `notifications`: notification API and Kafka-driven creation

Each module is split into:

- `domain`: entities, statuses, transitions, domain errors
- `application`: use cases and orchestration
- `transport`: HTTP handlers and DTOs
- `repository`: PostgreSQL persistence

## Source Of Truth

PostgreSQL is the source of truth for booking consistency. Seat availability lives in `session_seats`, and the hold flow locks the exact row with `SELECT ... FOR UPDATE`.

Redis is intentionally not the booking authority. It stores temporary hold TTL keys, seat map cache, idempotency lookup cache, and future rate limit counters. If Redis fails after a successful database commit, the booking remains consistent.

## Kafka And Outbox

Business transactions write rows to `outbox_events`. The worker publishes those rows to Kafka topics and marks them as `published` after Kafka confirms the write.

This avoids losing domain events when a transaction commits but the process crashes before publishing.

## Runtime Modes

- `backend-api`: stateless HTTP server. It can be horizontally scaled.
- `backend-worker`: expiration job, outbox publisher, notification consumers. The expiration job uses `FOR UPDATE SKIP LOCKED`, so multiple worker replicas are safe, though MVP manifests start one replica.

## Scaling Notes

API replicas are stateless and can be scaled behind a Service or ingress. PostgreSQL row locks remain the double-booking protection. Redis improves read performance and short-lived metadata, but losing Redis does not make confirmed booking state inconsistent.
