# API

Все бизнес-endpoint-ы находятся под `/api/v1`. Все ответы возвращаются в формате JSON.

## Health

- `GET /health`
- `GET /ready`
- `GET /metrics`

## Формат ошибки

```json
{
  "error": {
    "code": "SEAT_NOT_AVAILABLE",
    "message": "Seat is not available",
    "details": {}
  }
}
```

Типовые коды: `VALIDATION_ERROR`, `UNAUTHORIZED`, `FORBIDDEN`, `NOT_FOUND`, `SEAT_NOT_AVAILABLE`, `BOOKING_NOT_FOUND`, `BOOKING_EXPIRED`, `BOOKING_NOT_PENDING`, `PAYMENT_ALREADY_PROCESSED`, `IDEMPOTENCY_CONFLICT`, `INTERNAL_ERROR`.

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

Оба endpoint-а возвращают данные пользователя и токены (`accessToken`, `refreshToken`).

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

`GET /api/v1/auth/me` требует заголовок `Authorization: Bearer <accessToken>`.

## События

- `GET /api/v1/events`
- `GET /api/v1/events/{eventId}`
- `GET /api/v1/events/{eventId}/sessions`

## Сеансы и места

- `GET /api/v1/sessions/{sessionId}`
- `GET /api/v1/sessions/{sessionId}/seats`

Пример ответа карты мест:

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

## Брони

Все endpoint-ы броней требуют авторизацию.

`POST /api/v1/bookings/hold`

```json
{
  "sessionId": "uuid",
  "seatId": "uuid"
}
```

Пример ответа:

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

## Платежи

Все endpoint-ы платежей требуют авторизацию.

`POST /api/v1/payments`

```json
{
  "bookingId": "uuid",
  "idempotencyKey": "unique-key",
  "forceStatus": "succeeded"
}
```

`forceStatus` может быть `succeeded` или `failed`. Если поле пропущено, mock-платёж считается успешным.

`GET /api/v1/payments/{paymentId}`

Доступ к платежам ограничен владельцем: пользователь не может получить или обработать чужой платёж.

Правила идемпотентности:

- Повтор с тем же `idempotencyKey` и тем же `bookingId` возвращает существующий платёж.
- Использование того же `idempotencyKey` для другой брони возвращает `409 IDEMPOTENCY_CONFLICT`.
- Повтор с другим `forceStatus` для того же ключа возвращает `409 IDEMPOTENCY_CONFLICT`.

## Уведомления

Все endpoint-ы уведомлений требуют авторизацию.

- `GET /api/v1/notifications`
- `POST /api/v1/notifications/{id}/read`
