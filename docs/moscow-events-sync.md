# Moscow Events Sync

## Цель

Синхронизация загружает как можно больше актуальных мероприятий Москвы из легальных внешних API и сохраняет их в PostgreSQL для каталога `/events` и карты `/events/map`.

## Scope

Текущий scope жёстко ограничен Москвой:

- KudaGo: `location=msk`
- Timepad: `cities=Москва`

По умолчанию не импортируются:

- Санкт-Петербург
- другие города РФ
- другие страны
- country-wide и international events

## Defaults

```env
EXTERNAL_SYNC_ENABLED=false
EXTERNAL_SYNC_INTERVAL=6h
EXTERNAL_SYNC_CITY=moscow
EXTERNAL_SYNC_KUDAGO_LOCATION=msk
EXTERNAL_SYNC_TIMEPAD_CITY=Москва
EXTERNAL_SYNC_DAYS_AHEAD=180
EXTERNAL_SYNC_LOOKBACK_DAYS=14
EXTERNAL_SYNC_MAX_PAGES=500
EXTERNAL_SYNC_PAGE_SIZE=100
```

## Manual sync

Для Docker Compose используйте host-порт `18080`. Direct standalone backend без Compose по-прежнему может слушать `8080`.

### Sync оба провайдера

```bash
curl -X POST http://localhost:18080/api/v1/admin/integrations/sync/moscow \
  -H "Authorization: Bearer <admin-access-token>" \
  -H "Content-Type: application/json" \
  -d '{"providers":["kudago","timepad"],"daysAhead":180,"lookbackDays":14}'
```

### Sync только KudaGo

```bash
curl -X POST http://localhost:18080/api/v1/admin/integrations/sync/kudago \
  -H "Authorization: Bearer <admin-access-token>" \
  -H "Content-Type: application/json" \
  -d '{"location":"msk","daysAhead":180,"lookbackDays":14}'
```

### Sync только Timepad

```bash
curl -X POST http://localhost:18080/api/v1/admin/integrations/sync/timepad \
  -H "Authorization: Bearer <admin-access-token>" \
  -H "Content-Type: application/json" \
  -d '{"city":"Москва","daysAhead":180,"lookbackDays":14}'
```

### История запусков

```bash
curl http://localhost:18080/api/v1/admin/integrations/runs \
  -H "Authorization: Bearer <admin-access-token>"
```

## Worker sync

Для фоновой синхронизации:

```env
EXTERNAL_SYNC_ENABLED=true
EXTERNAL_SYNC_INTERVAL=6h
```

Worker:

- запускает sync сразу после старта
- затем повторяет каждые `EXTERNAL_SYNC_INTERVAL`
- пишет статистику по каждому provider
- не падает из-за отказа одного provider

## Pagination

Полная пагинация обязательна.

### KudaGo

- start page: `1`
- `page_size=100`
- продолжаем, пока `response.next != null`
- ограничение сверху: `EXTERNAL_SYNC_MAX_PAGES`

### Timepad

- start `skip=0`
- `limit=100`
- `skip += limit`
- продолжаем, пока `skip < total` и `values` не пуст
- ограничение сверху: `EXTERNAL_SYNC_MAX_PAGES`

## Dedupe

Повторный sync не должен создавать дубли.

Используются:

- exact match по provider id
- soft match по `normalized_title + date + venue`

Если событие совпало мягко:

- новый event не создаётся
- сохраняется `event_external_links`

## make-demo-bookable

Imported event можно перевести в demo booking:

```bash
curl -X POST http://localhost:18080/api/v1/admin/events/<eventId>/make-demo-bookable \
  -H "Authorization: Bearer <admin-access-token>"
```

Результат:

- `booking_mode=demo_bookable`
- создаётся `Demo Hall`
- создаются seats и `session_seats`
- появляется bookable session для existing ReserveFlow flow

## Ограничения

- imported external events не являются реальными билетными продажами
- imported sessions без promo не имеют seat inventory
- источники с scraping не используются
