# ReserveFlow

ReserveFlow объединяет два сценария в одном каталоге событий:

- внутренние события ReserveFlow с полноценным SeatMap -> Hold -> Checkout -> Mock Payment;
- импортированные Moscow-only события из легальных API KudaGo и Timepad для discovery-каталога и карты.

На первом этапе внешняя синхронизация работает только для Москвы. Санкт-Петербург, другие города, страны и международные события по умолчанию не импортируются.

## Что реализовано

- `/events` показывает внутренние и импортированные московские события в одном каталоге.
- `/events/map` показывает актуальные события Москвы на карте Leaflet.
- Imported events по умолчанию имеют `booking_mode=external_link_only`.
- `POST /api/v1/admin/events/{eventId}/make-demo-bookable` переводит imported event в demo booking без замены существующего booking core.
- Worker умеет запускать периодический Moscow sync по `EXTERNAL_SYNC_*` настройкам.

## Booking modes

- `reserveflow_managed`: обычное внутреннее событие ReserveFlow, бронирование доступно.
- `external_link_only`: imported event, доступен просмотр и переход к организатору, бронирование внутри ReserveFlow отключено.
- `demo_bookable`: imported event подключён к demo hall и проходит через текущий SeatMap/Hold/Checkout/Payment flow.
- `general_admission`: зарезервировано на будущее, сейчас не используется.

## Быстрый старт

Требуется Docker Desktop.

```powershell
cd C:\development\booking-system\reserveflow
powershell -ExecutionPolicy Bypass -File scripts\dev-up.ps1
```

После старта доступны:

- frontend: `http://localhost:3000`
- backend API: `http://localhost:18080/api/v1`
- backend health: `http://localhost:18080/health`
- Prometheus: `http://localhost:9090`
- Grafana: `http://localhost:3001`

Локальный frontend без `frontend/.env.local` сначала ищет ReserveFlow backend на `http://localhost:18080`, а затем на `http://localhost:8080`. Порт `8080` остаётся direct-режимом для standalone backend без Docker Compose.

Если нужно поднять окружение без seed:

```powershell
powershell -ExecutionPolicy Bypass -File scripts\dev-up.ps1 -SkipSeed
```

## Demo пользователи

- `demo@example.com` / `Password123!`
- `admin@example.com` / `Password123!`

Админ нужен для ручного sync и `make-demo-bookable`.

## Moscow-only sync

`.env.example` уже содержит безопасные дефолты:

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
KUDAGO_BASE_URL=https://kudago.com/public-api/v1.4
TIMEPAD_BASE_URL=https://api.timepad.ru
NEXT_PUBLIC_MAP_DEFAULT_LAT=55.751244
NEXT_PUBLIC_MAP_DEFAULT_LON=37.618423
NEXT_PUBLIC_MAP_DEFAULT_ZOOM=11
```

Ручной sync:

```bash
curl -X POST http://localhost:18080/api/v1/admin/integrations/sync/moscow \
  -H "Authorization: Bearer <admin-access-token>" \
  -H "Content-Type: application/json" \
  -d '{"providers":["kudago","timepad"],"daysAhead":180,"lookbackDays":14}'
```

Promotion imported event в demo booking:

```bash
curl -X POST http://localhost:18080/api/v1/admin/events/<eventId>/make-demo-bookable \
  -H "Authorization: Bearer <admin-access-token>"
```

## Команды качества

Backend:

```powershell
cd backend
go test ./...
```

Frontend:

```powershell
cd frontend
npm install
npm run typecheck
npm run lint
npm run build
npm run test
```

Integration tests используют `testcontainers` и требуют запущенный Docker daemon:

```powershell
cd backend
go test -tags=integration ./tests
```

## Ограничения текущего релиза

- Imported external events используются как discovery-данные, а не как реальные продажи билетов.
- Без `make-demo-bookable` imported sessions не имеют seat inventory и не участвуют в бронировании.
- Yandex Afisha scraping и другой scraping афиш не используются.
- Future sources вроде PRO.Культура.РФ и `data.mos.ru` пока только задокументированы.

## Документация

- `docs/api.md`
- `docs/frontend.md`
- `docs/deployment.md`
- `docs/integration-notes.md`
- `docs/external-events.md`
- `docs/moscow-events-sync.md`
