# ReserveFlow

ReserveFlow — полнофункциональная система бронирования мест на события. Проект включает Go backend с JSON REST API и Next.js frontend, который покрывает полный пользовательский путь: регистрация, вход, каталог событий, выбор мест, удержание, оплата, история броней и уведомления.

Ключевая инженерная цель — безопасное конкурентное бронирование. Источник истины — PostgreSQL: в сценарии hold используется транзакция и блокировка `SELECT ... FOR UPDATE` по `session_seats`, чтобы только один параллельный запрос мог удержать конкретное место.

## Технологии

- Go + HTTP-роутер chi
- PostgreSQL
- Redis (TTL hold-ключей, кэш карты мест, кэш идемпотентности платежей)
- Kafka (совместимый с Redpanda)
- Outbox pattern
- Метрики Prometheus
- Структурированные логи zerolog
- Docker Compose и Kubernetes-манифесты
- Next.js frontend с auth-прокси через HTTP-only cookie

## Архитектура

Backend реализован как модульный монолит в `backend/internal/modules`:

- `auth`
- `events`
- `sessions`
- `seats`
- `bookings`
- `payments`
- `notifications`

Каждый модуль следует слоям: `domain`, `application`, `transport`, `repository`.

У Go-приложения два режима запуска:

- `backend-api`: HTTP API
- `backend-worker`: expire job, outbox publisher, consumers уведомлений

## Быстрый старт

```sh
cd reserveflow
make up
make migrate-up
make seed
```

После запуска:

- frontend: `http://localhost:3000`
- backend API: `http://localhost:8080`
- Prometheus: `http://localhost:9090`
- Grafana: `http://localhost:3000`

## Локальный запуск backend

```sh
cd reserveflow
cp .env.example .env
make api
make worker
```

Для локального запуска должны быть доступны PostgreSQL, Redis и Kafka/Redpanda из параметров `.env.example`.

## Локальный запуск frontend

```sh
cd reserveflow/frontend
cp .env.example .env.local
npm install
npm run dev
```

Схема запросов:

```text
Браузер -> Next.js (/api/auth, /api/backend) -> Go backend (/api/v1)
```

JWT хранится только в HTTP-only cookie, не в `localStorage`.

## Тестовый пользователь

- Email: `demo@example.com`
- Пароль: `Password123!`

## Основные endpoint-ы

- `POST /api/v1/auth/register`
- `POST /api/v1/auth/login`
- `GET /api/v1/events`
- `GET /api/v1/events/{eventId}/sessions`
- `GET /api/v1/sessions/{sessionId}/seats`
- `POST /api/v1/bookings/hold`
- `POST /api/v1/payments`
- `GET /api/v1/bookings/me`
- `GET /api/v1/notifications`

## Демо-сценарий

1. Откройте `/login` или `/register`.
2. Авторизуйтесь как `demo@example.com` / `Password123!`.
3. Откройте `/events`.
4. Выберите событие.
5. Выберите сеанс.
6. Выберите свободное место на `/sessions/{sessionId}`.
7. Нажмите «Удержать место».
8. Перейдите на `/checkout/{bookingId}`.
9. Выполните успешную оплату или симулируйте неуспешную.
10. Проверьте `/bookings` и `/notifications`.

Пример hold-запроса:

```json
{
  "sessionId": "50000000-0000-0000-0000-000000000001",
  "seatId": "seat-uuid"
}
```

Пример запроса на оплату:

```json
{
  "bookingId": "booking-uuid",
  "idempotencyKey": "demo-payment-1",
  "forceStatus": "succeeded"
}
```

## Проверки и тесты

```sh
cd reserveflow
make test
make test-integration
make frontend-typecheck
make frontend-lint
make frontend-build
```

Интеграционные тесты используют testcontainers и требуют Docker.

Покрыты auth, каталог, карта мест, hold, успешная/неуспешная оплата, идемпотентность, проверка владельца платежа, освобождение по expire и критичный конкурентный сценарий hold.

Frontend-тесты:

```sh
cd reserveflow/frontend
npm run test
```
