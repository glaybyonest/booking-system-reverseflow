# ReserveFlow

ReserveFlow is a backend-only reservation system for event seats. It exposes a JSON REST API for registration, login, event browsing, seat map lookup, temporary seat holds, mock payments, booking history, and notifications.

The primary engineering goal is safe concurrent booking. PostgreSQL is the source of truth, and the hold flow uses a database transaction with `SELECT ... FOR UPDATE` on `session_seats`, so only one concurrent request can hold a seat.

## Stack

- Go, chi HTTP router
- PostgreSQL
- Redis for hold TTL keys, seat map cache, payment idempotency cache, and future rate limit metadata
- Redpanda-compatible Kafka
- Outbox pattern
- Prometheus metrics
- Structured zerolog logs
- Docker Compose and Kubernetes manifests

No frontend, React, Next.js, CSS, or real payment/email provider is included.

## Architecture

The backend is a modular monolith under `backend/internal/modules`:

- `auth`
- `events`
- `sessions`
- `seats`
- `bookings`
- `payments`
- `notifications`

Each module follows layered architecture: `domain`, `application`, `transport`, and `repository`.

The same Go codebase has two runtime modes:

- `backend-api`: HTTP API
- `backend-worker`: expiration job, outbox publisher, notification consumers

## Quick Start

```sh
cd reserveflow
make up
make migrate-up
make seed
```

API is available at `http://localhost:8080`.

Prometheus is available at `http://localhost:9090`.
Grafana is available at `http://localhost:3000`.

## Local Go Run

```sh
cd reserveflow
cp .env.example .env
make api
make worker
```

The local commands expect PostgreSQL, Redis, and Redpanda/Kafka to be reachable from `.env.example` values.

## Test User

- Email: `demo@example.com`
- Password: `Password123!`

## Main Endpoints

- `POST /api/v1/auth/register`
- `POST /api/v1/auth/login`
- `GET /api/v1/events`
- `GET /api/v1/events/{eventId}/sessions`
- `GET /api/v1/sessions/{sessionId}/seats`
- `POST /api/v1/bookings/hold`
- `POST /api/v1/payments`
- `GET /api/v1/bookings/me`
- `GET /api/v1/notifications`

## Demo Flow

1. Login with `demo@example.com` and `Password123!`.
2. Call `GET /api/v1/events`.
3. Pick an event and call `GET /api/v1/events/{eventId}/sessions`.
4. Pick a session and call `GET /api/v1/sessions/{sessionId}/seats`.
5. Hold one available seat:

```json
{
  "sessionId": "50000000-0000-0000-0000-000000000001",
  "seatId": "seat-uuid"
}
```

6. Pay with mock payment:

```json
{
  "bookingId": "booking-uuid",
  "idempotencyKey": "demo-payment-1",
  "forceStatus": "succeeded"
}
```

## Tests

```sh
cd reserveflow
make test
make test-integration
```

Integration tests use testcontainers and require Docker.

The integration suite covers auth, catalog reads, seat map, hold, payment success/failure, idempotency, payment ownership, expiration release, and the critical concurrent hold scenario.
