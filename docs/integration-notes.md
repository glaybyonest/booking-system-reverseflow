# Integration Notes

## Что проверено

- auth через Next.js route handlers и HTTP-only cookies
- backend proxy `/api/backend/*`
- новый public catalog `/events`
- новая карта `/events/map`
- detail page с imported event states
- demo promotion через `/api/v1/admin/events/{eventId}/make-demo-bookable`
- booking flow для `demo_bookable` без замены SeatMap/Hold/Checkout/Payment core

## DTO и contract compatibility

Frontend по-прежнему понимает оба стиля именования:

- `posterUrl` и `poster_url`
- `sourceUrl` и `source_url`
- `startsAt` и `starts_at`
- `endsAt` и `ends_at`
- `bookingMode` и `booking_mode`
- `externalSource` и `external_source`
- `eventId` и `event_id`
- `hallId` и `hall_id`
- `isBookable` и `is_bookable`

Основная нормализация находится в `frontend/src/shared/api/mappers.ts`.

## External providers

### KudaGo

- только `location=msk`
- используется полная пагинация по `next`
- импортируются future и ongoing events
- при наличии expanded `place` сохраняются venue name, address и coords

### Timepad

- только `cities=Москва`
- используется полная пагинация по `skip + limit`
- сохраняются только события с физическим Moscow location
- online-only events без московского адреса отбрасываются

## Dedupe strategy

Идемпотентность sync строится в два слоя:

1. exact duplicate: `events.external_source + events.external_id`
2. soft duplicate: `dedupe_key = normalized_title + date + venue`

Если одно и то же событие пришло из разных провайдеров:

- новый canonical event не создаётся
- в `event_external_links` добавляется связь на второй provider id
- детальная страница показывает merged links

## Imported sessions

Imported occurrences хранятся как `sessions` с:

- `hall_id = NULL`
- `is_bookable = false`

Это позволяет:

- показывать расписание imported event
- не ломать существующий booking core
- не создавать лишние seat inventories до promo в demo booking

## Demo promotion

`make-demo-bookable`:

- переиспользует existing imported event
- создаёт или переиспользует venue
- создаёт hall `Demo Hall`
- создаёт rows `A-D`
- создаёт 10 seats на ряд
- создаёт `session_seats`
- переводит event в `booking_mode=demo_bookable`
- делает сессию bookable без переписывания existing booking logic

## Worker and outbox

- worker запускает `MoscowEventsSyncJob`, если `EXTERNAL_SYNC_ENABLED=true`
- failure одного provider не валит весь worker
- успешный run пишет outbox event `external_events.synced`
- существующие Kafka/outbox сценарии бронирования не менялись

## Ограничения

- imported external events являются discovery-данными, а не реальными ticket sales
- canonical source badge в каталоге один, дополнительные provider links отображаются на detail page
- future sources пока не подключены
