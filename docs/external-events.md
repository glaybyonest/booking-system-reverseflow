# External Events

## Назначение

External events расширяют каталог ReserveFlow discovery-данными из легальных API, не заменяя существующий booking engine.

На текущем этапе включён только Moscow-only import.

## Источники

- KudaGo с `location=msk`
- Timepad с `cities=Москва`

Scraping афиш не используется. Yandex Afisha scraping запрещён.

## Data model

### Events

Для imported events используются дополнительные поля:

- `source`
- `external_source`
- `external_id`
- `source_url`
- `imported_at`
- `last_synced_at`
- `raw_payload`
- `booking_mode`
- `starts_at`
- `ends_at`
- `normalized_title`
- `dedupe_key`
- `venue_id`

### Venues

Venue может приходить из внешнего provider и хранит:

- `external_source`
- `external_id`
- `latitude`
- `longitude`
- `source_url`
- `seat_map_provider`
- `raw_payload`

### Sessions

Imported occurrences создаются как `sessions` и по умолчанию имеют:

- `hall_id = NULL`
- `is_bookable = false`
- `external_source`
- `external_id`
- `source_url`

## Booking modes

### `reserveflow_managed`

Внутреннее событие ReserveFlow. Seat inventory уже существует и бронирование доступно сразу.

### `external_link_only`

Imported event. Пользователь может:

- посмотреть карточку события
- перейти по `source_url`
- увидеть площадку, адрес и карту

Пользователь не может:

- открыть seat selection
- создать hold
- оформить оплату внутри ReserveFlow

### `demo_bookable`

Imported event, который администратор продвинул в demo flow. После этого:

- создаётся demo hall
- создаются seats и `session_seats`
- становится доступен существующий ReserveFlow booking flow

### `general_admission`

Поле зарезервировано под future work и пока не используется.

## Canonical event и merged providers

Если одно событие пришло из двух источников, ReserveFlow старается не плодить дубль:

- exact match определяется по `external_source + external_id`
- soft match определяется по `dedupe_key`

При soft merge:

- canonical event остаётся один
- второй provider сохраняется в `event_external_links`
- detail page показывает дополнительные source links

## Почему imported events не продают билеты

Imported events в текущем релизе нужны для discovery и навигации по актуальной афише Москвы.

Это значит:

- imported данные не считаются real ticket inventory
- цены и availability не считаются коммерчески точными
- переход к продаже реальных билетов должен интегрироваться отдельно

## Future Moscow event sources

- PRO.Культура.РФ может дать большой пласт культурных событий, но для production-интеграции нужен API key.
- `data.mos.ru` стоит исследовать как источник открытых московских культурных датасетов.
- Ticketmaster не является приоритетом для московского релиза.
- Yandex Afisha не должна подключаться через scraping.
