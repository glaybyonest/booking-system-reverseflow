# API

Все бизнес-endpoint'ы находятся под `/api/v1` и возвращают JSON.

## Health

- `GET /health`
- `GET /ready`
- `GET /metrics`

## Формат ошибки

```json
{
  "error": {
    "code": "SESSION_NOT_BOOKABLE",
    "message": "Session is not bookable",
    "details": {}
  }
}
```

Типовые коды: `VALIDATION_ERROR`, `UNAUTHORIZED`, `FORBIDDEN`, `NOT_FOUND`, `SEAT_NOT_AVAILABLE`, `SEAT_ALREADY_HELD`, `SESSION_NOT_BOOKABLE`, `BOOKING_NOT_FOUND`, `BOOKING_EXPIRED`, `BOOKING_NOT_PENDING`, `PAYMENT_ALREADY_PROCESSED`, `IDEMPOTENCY_CONFLICT`, `INTERNAL_ERROR`.

## Auth

- `POST /api/v1/auth/register`
- `POST /api/v1/auth/login`
- `POST /api/v1/auth/refresh`
- `POST /api/v1/auth/logout`
- `GET /api/v1/auth/me`

Логин и регистрация возвращают пользователя и токены `accessToken`/`refreshToken`.

## Public events API

### `GET /api/v1/events`

Query params:

- `city`, по умолчанию `Moscow`
- `source`
- `category`
- `from`
- `to`
- `bookingMode`
- `onlyActual`, по умолчанию `true`
- `limit`, по умолчанию `24`
- `offset`, по умолчанию `0`

Поддерживаемые значения `source`:

- `manual`
- `reserveflow`
- `kudago`
- `timepad`

Поддерживаемые значения `bookingMode`:

- `reserveflow_managed`
- `external_link_only`
- `demo_bookable`
- `general_admission`
- `bookable` как alias для `reserveflow_managed + demo_bookable`

Пример ответа:

```json
{
  "items": [
    {
      "id": "uuid",
      "title": "Ночной концерт в Москве",
      "description": "Живое выступление в центре города",
      "category": "concert",
      "posterUrl": "https://images.example.com/poster.jpg",
      "status": "published",
      "source": "kudago",
      "externalSource": "kudago",
      "sourceUrl": "https://kudago.com/msk/event/night-concert/",
      "bookingMode": "external_link_only",
      "startsAt": "2030-01-01T18:00:00+03:00",
      "endsAt": "2030-01-01T20:00:00+03:00",
      "isImported": true,
      "venue": {
        "id": "uuid",
        "name": "Дом музыки",
        "address": "Москва, Космодамианская набережная, 52",
        "city": "Москва",
        "latitude": 55.7343,
        "longitude": 37.6467
      }
    }
  ],
  "total": 1
}
```

### `GET /api/v1/events/map`

Использует те же фильтры, что и `/events`, но возвращает только актуальные события с координатами.

```json
{
  "events": [
    {
      "id": "uuid",
      "title": "Ночной концерт в Москве",
      "category": "concert",
      "posterUrl": "https://images.example.com/poster.jpg",
      "source": "kudago",
      "bookingMode": "external_link_only",
      "startsAt": "2030-01-01T18:00:00+03:00",
      "endsAt": "2030-01-01T20:00:00+03:00",
      "venue": {
        "id": "uuid",
        "name": "Дом музыки",
        "address": "Москва, Космодамианская набережная, 52",
        "city": "Москва",
        "latitude": 55.7343,
        "longitude": 37.6467
      }
    }
  ]
}
```

### `GET /api/v1/events/{eventId}`

Возвращает полную карточку события:

- `source`
- `externalSource`
- `sourceUrl`
- `bookingMode`
- `startsAt`
- `endsAt`
- `venue` с координатами
- `externalLinks`
- `sessions`

### `GET /api/v1/events/{eventId}/sessions`

Возвращает `items[]` с сессиями события. Для imported occurrences без demo promotion у session:

- `hallId=null`
- `isBookable=false`

## Sessions and seats

- `GET /api/v1/sessions/{sessionId}`
- `GET /api/v1/sessions/{sessionId}/seats`

`GET /api/v1/sessions/{sessionId}` работает и для imported sessions с nullable hall.

`GET /api/v1/sessions/{sessionId}/seats` и последующий booking flow предназначены только для `isBookable=true`.

## Bookings

Все booking endpoint'ы требуют `Authorization: Bearer <accessToken>`.

### `POST /api/v1/bookings/hold`

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

Дополнительно:

- `GET /api/v1/bookings/{bookingId}`
- `GET /api/v1/bookings/me`
- `POST /api/v1/bookings/{bookingId}/cancel`

## Payments

Все payment endpoint'ы требуют авторизацию.

### `POST /api/v1/payments`

```json
{
  "bookingId": "uuid",
  "idempotencyKey": "unique-key",
  "forceStatus": "succeeded"
}
```

`forceStatus` может быть `succeeded` или `failed`. Если поле опущено, mock-платёж считается успешным.

### `GET /api/v1/payments/{paymentId}`

Платеж доступен только владельцу брони.

## Notifications

- `GET /api/v1/notifications`
- `POST /api/v1/notifications/{id}/read`

## Admin integrations API

Все admin endpoint'ы требуют JWT пользователя с `role=admin`.

### `POST /api/v1/admin/integrations/sync/moscow`

```json
{
  "providers": ["kudago", "timepad"],
  "daysAhead": 180,
  "lookbackDays": 14
}
```

### `POST /api/v1/admin/integrations/sync/kudago`

```json
{
  "location": "msk",
  "daysAhead": 180,
  "lookbackDays": 14
}
```

### `POST /api/v1/admin/integrations/sync/timepad`

```json
{
  "city": "Москва",
  "daysAhead": 180,
  "lookbackDays": 14
}
```

### `GET /api/v1/admin/integrations/runs`

Возвращает import run'ы со статусом, счётчиками, `pageCount` и возможным `errorMessage`.

### `POST /api/v1/admin/events/{eventId}/make-demo-bookable`

Переключает imported event в `demo_bookable`, создаёт или переиспользует:

- venue
- hall `Demo Hall`
- rows `A-D`
- 10 seats в каждом ряду
- session
- session_seats

Ответ:

```json
{
  "sessionId": "uuid"
}
```
