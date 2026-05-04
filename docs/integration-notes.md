# Integration Notes

## Что проверено end-to-end

- Сверка frontend API-вызовов с реальными backend route-ами (`/api/v1/*`).
- Сверка TypeScript DTO с реальными JSON-ответами backend.
- Проверка auth flow: `login`, `register`, `refresh`, `logout`, `protected routes`.
- Проверка booking flow: события, детали, сеансы, polling карты мест, hold, 409 conflict, checkout timer, payment idempotency, мои брони, уведомления.
- Прогон `npm run lint`, `npm run typecheck`, `npm run build`.

Все перечисленные проверки проходят, блокирующих несовпадений не найдено.

## Реально используемые endpoint-ы

Auth:

- `POST /api/v1/auth/register`
- `POST /api/v1/auth/login`
- `POST /api/v1/auth/refresh`
- `POST /api/v1/auth/logout`
- `GET /api/v1/auth/me`

Каталог и сеансы:

- `GET /api/v1/events`
- `GET /api/v1/events/{eventId}`
- `GET /api/v1/events/{eventId}/sessions`
- `GET /api/v1/sessions/{sessionId}`
- `GET /api/v1/sessions/{sessionId}/seats`

Брони:

- `POST /api/v1/bookings/hold`
- `GET /api/v1/bookings/{bookingId}`
- `GET /api/v1/bookings/me`
- `POST /api/v1/bookings/{bookingId}/cancel`

Платежи:

- `POST /api/v1/payments`
- `GET /api/v1/payments/{paymentId}`

Уведомления:

- `GET /api/v1/notifications`
- `POST /api/v1/notifications/{id}/read`

## Какие DTO были адаптированы

- `frontend/src/shared/api/mappers.ts`:
  - Поддержка и `camelCase`, и `snake_case` для ключевых сущностей (`event`, `session`, `seat map`, `booking`, `payment`, `notification`).
  - Добавлен `normalizeHoldSeatResponse()` для нормализации ответа `POST /bookings/hold` и выравнивания полей `bookingId/expiresAt/totalPrice/seat`.
- `frontend/src/shared/api/bookings.api.ts`:
  - `holdSeat()` теперь использует mapper (`normalizeHoldSeatResponse`) вместо предположения о «идеальном» формате payload.

## Backend assumptions

- Все бизнес-endpoint-ы доступны под префиксом `/api/v1`.
- Защищённые route-ы backend требуют `Authorization: Bearer <accessToken>`.
- Refresh делается через `POST /auth/refresh` с телом `{ refreshToken }`.
- 409 в сценарии удержания места возвращается в формате API-ошибки и обрабатывается фронтом через единый `friendlyApiError`.
- Идемпотентность платежей обеспечивается backend (`idempotencyKey`) и может возвращать `409 IDEMPOTENCY_CONFLICT`.
- Уведомления появляются асинхронно (через worker + outbox + Kafka), поэтому возможна естественная задержка между действием пользователя и появлением уведомления.

## Что отложено на v2

- Расширить DTO брони данными для UX (название события, зал, читаемое время сеанса) прямо в `GET /bookings/*`, чтобы убрать зависимость от fallback-полей в UI.
- Добавить endpoint со сводкой непрочитанных уведомлений (например, `GET /notifications/unread-count`) для более дешёвого обновления индикатора в header.
- Добавить интеграционные e2e-тесты фронта против живого backend в CI (с поднятием полного окружения), чтобы автоматически проверять пользовательские сценарии целиком.
