# Testing

## Unit Tests

Unit tests cover:

- booking status transitions
- `CanConfirm`
- `CanCancel`
- `CanExpire`
- hold duration calculation
- domain errors
- payment forced status validation

Run:

```sh
cd reserveflow
make test
```

## Integration Tests

Integration tests are behind the `integration` build tag and use testcontainers.

They cover:

- register/login/me
- event list
- event sessions
- session details
- seat map
- hold seat
- payment success
- payment idempotent replay
- idempotency conflict
- payment ownership guard
- payment failure and seat release
- expiration job release
- critical concurrent hold

Run:

```sh
cd reserveflow
make test-integration
```

## Critical Concurrency Test

`backend/tests/booking_concurrency_integration_test.go` starts PostgreSQL, applies migrations and seed data, then runs 20 parallel hold attempts for the same `session_seat`.

Expected result:

- 1 successful hold
- 19 conflicts
- exactly one `pending` booking
- `session_seat.status = held`

This verifies that PostgreSQL row locking is the double-booking guard.
