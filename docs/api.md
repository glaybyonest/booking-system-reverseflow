# API

All business endpoints are under `/api/v1`. All responses are JSON.

## Health

- `GET /health`
- `GET /ready`
- `GET /metrics`

## Error Format

```json
{
  "error": {
    "code": "SEAT_NOT_AVAILABLE",
    "message": "Seat is not available",
    "details": {}
  }
}
```

Common codes: `VALIDATION_ERROR`, `UNAUTHORIZED`, `FORBIDDEN`, `NOT_FOUND`, `SEAT_NOT_AVAILABLE`, `BOOKING_NOT_FOUND`, `BOOKING_EXPIRED`, `BOOKING_NOT_PENDING`, `PAYMENT_ALREADY_PROCESSED`, `IDEMPOTENCY_CONFLICT`, `INTERNAL_ERROR`.

## Auth

`POST /api/v1/auth/register`

```json
{
  "email": "demo@example.com",
  "password": "Password123!",
  "name": "Demo User"
}
```

`POST /api/v1/auth/login`

```json
{
  "email": "demo@example.com",
  "password": "Password123!"
}
```

Both return user data plus `accessToken` and `refreshToken`.

`POST /api/v1/auth/refresh`

```json
{
  "refreshToken": "jwt-refresh-token"
}
```

`POST /api/v1/auth/logout`

```json
{
  "refreshToken": "jwt-refresh-token"
}
```

`GET /api/v1/auth/me` requires `Authorization: Bearer <accessToken>`.

## Events

- `GET /api/v1/events`
- `GET /api/v1/events/{eventId}`
- `GET /api/v1/events/{eventId}/sessions`

## Sessions And Seats

- `GET /api/v1/sessions/{sessionId}`
- `GET /api/v1/sessions/{sessionId}/seats`

Seat map response:

```json
{
  "sessionId": "uuid",
  "event": {
    "id": "uuid",
    "title": "Jazz Night"
  },
  "hall": {
    "id": "uuid",
    "name": "Main Hall"
  },
  "seats": [
    {
      "seatId": "uuid",
      "row": "A",
      "number": 1,
      "status": "available",
      "price": 700
    }
  ]
}
```

## Bookings

All booking endpoints require auth.

`POST /api/v1/bookings/hold`

```json
{
  "sessionId": "uuid",
  "seatId": "uuid"
}
```

Response:

```json
{
  "bookingId": "uuid",
  "status": "pending",
  "expiresAt": "2026-05-04T12:10:00Z",
  "seat": {
    "seatId": "uuid",
    "row": "A",
    "number": 7
  },
  "totalPrice": 700
}
```

- `GET /api/v1/bookings/{bookingId}`
- `GET /api/v1/bookings/me`
- `POST /api/v1/bookings/{bookingId}/cancel`

## Payments

All payment endpoints require auth.

`POST /api/v1/payments`

```json
{
  "bookingId": "uuid",
  "idempotencyKey": "unique-key",
  "forceStatus": "succeeded"
}
```

`forceStatus` can be `succeeded` or `failed`. If omitted, mock payment succeeds.

`GET /api/v1/payments/{paymentId}`

Payment access is owner-scoped. A user cannot fetch or process another user's payment.

Idempotency rules:

- Repeating the same `idempotencyKey` for the same booking returns the existing payment.
- Reusing the same `idempotencyKey` for another booking returns `409 IDEMPOTENCY_CONFLICT`.
- Replaying the same `idempotencyKey` with a different `forceStatus` returns `409 IDEMPOTENCY_CONFLICT`.

## Notifications

All notification endpoints require auth.

- `GET /api/v1/notifications`
- `POST /api/v1/notifications/{id}/read`
