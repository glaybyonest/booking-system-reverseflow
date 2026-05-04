# Architecture

ReserveFlow — модульный монолит на Go (backend-first). В MVP это единый deployable без разделения на микросервисы, но с чёткими границами доменных модулей.

## Модули

- `auth`: пользователи, bcrypt, access/refresh JWT
- `events`: каталог опубликованных событий
- `sessions`: расписание сеансов и данные залов
- `seats`: read model карты мест и кэш в Redis
- `bookings`: hold/cancel/expire/confirm/payment_failed переходы
- `payments`: mock-провайдер и идемпотентность
- `notifications`: API уведомлений и генерация через Kafka события

Структура каждого модуля:

- `domain`: сущности, статусы, переходы, доменные ошибки
- `application`: use case-ы и оркестрация
- `transport`: HTTP handlers и DTO
- `repository`: работа с PostgreSQL

## Источник истины

Источник истины для консистентности бронирований — PostgreSQL. Состояние мест хранится в `session_seats`, а hold-сценарий блокирует конкретную строку через `SELECT ... FOR UPDATE`.

Redis не является authority для брони. Он используется для:

- TTL-ключей удержания
- кэша карты мест
- кэша lookup по идемпотентности платежей
- будущих счётчиков rate-limit

Если Redis недоступен после успешного DB commit, данные брони остаются консистентными.

## Kafka и Outbox

Доменные события сначала пишутся в `outbox_events` в рамках бизнес-транзакции. Затем worker публикует их в Kafka и помечает как `published`.

Это предотвращает потерю событий в случае, когда транзакция уже зафиксирована, а процесс упал до отправки в Kafka.

## Режимы запуска

- `backend-api`: stateless HTTP сервер, горизонтально масштабируется
- `backend-worker`: expire job, outbox publisher, consumers уведомлений

Expire job использует `FOR UPDATE SKIP LOCKED`, поэтому допускает несколько реплик worker-а.

## Масштабирование

API-реплики stateless и масштабируются за ingress/service. Защита от double-booking обеспечивается row lock-ами PostgreSQL. Redis ускоряет чтение и хранит краткоживущие метаданные, но не влияет на консистентность подтверждённых броней.
