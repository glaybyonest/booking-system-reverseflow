# Booking Flow

## Hold Flow

`POST /api/v1/bookings/hold`

1. The API gets `userId` from JWT claims.
2. The application validates `sessionId` and `seatId`.
3. PostgreSQL transaction starts.
4. Repository runs:

```sql
SELECT *
FROM session_seats
WHERE session_id = $1 AND seat_id = $2
FOR UPDATE;
```

5. If the row is not `available`, the request returns `409 SEAT_NOT_AVAILABLE`.
6. Booking expiration is calculated as `now + HOLD_TTL`.
7. `session_seats` becomes `held`.
8. A `pending` booking and `held` booking item are inserted.
9. `seat.held` and `booking.created` are written to `outbox_events`.
10. Transaction commits.
11. Redis hold key is written with TTL.
12. Seat map cache is invalidated.

Only one concurrent transaction can hold the same `session_seats` row. The other requests wait, see `held`, and return conflict.

## Payment Success

`POST /api/v1/payments`

The payment use case locks the booking, checks ownership, verifies `pending` status and non-expired `expires_at`, then creates a mock payment.

On success:

- payment becomes `succeeded`
- booking becomes `confirmed`
- booking item becomes `booked`
- session seat becomes `booked`
- hold key is removed
- seat map cache is invalidated
- `payment.succeeded` and `booking.confirmed` events are written

## Payment Failure

On forced or mock failure:

- payment becomes `failed`
- booking becomes `payment_failed`
- booking item becomes `payment_failed`
- session seat becomes `available`
- hold key is removed
- seat map cache is invalidated
- `payment.failed` event is written

## Expiration

The worker runs every 15 seconds:

```sql
SELECT id
FROM bookings
WHERE status = 'pending' AND expires_at < now()
ORDER BY expires_at
LIMIT 100
FOR UPDATE SKIP LOCKED;
```

For each expired booking it marks the booking and item as `expired`, releases the session seat, writes `booking.expired`, removes Redis hold key, and invalidates seat map cache.

The job is idempotent because it only selects `pending` expired bookings.

The worker first collects locked booking IDs, closes the result set, and only then performs follow-up queries in the same transaction. This avoids `pgx` connection-busy errors while keeping `FOR UPDATE SKIP LOCKED` semantics.
