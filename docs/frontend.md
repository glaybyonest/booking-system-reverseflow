# Frontend

Frontend ReserveFlow находится в `frontend/` и построен на Next.js App Router.

## Технологии

- Next.js 16
- React 19
- TypeScript
- Tailwind CSS
- TanStack Query
- Leaflet + `react-leaflet`
- React Hook Form
- Zod

## Структура

- `src/app`: страницы и route handlers
- `src/features`: feature-компоненты и hooks
- `src/entities`: нормализованные TS-модели
- `src/widgets`: переиспользуемые UI-блоки
- `src/shared`: API-клиенты, утилиты, конфиг, UI primitives

## Auth flow

Frontend не хранит backend токены в `localStorage`.

1. Пользователь логинится через `/api/auth/login` или `/api/auth/register`.
2. Next.js route handler обращается к Go backend.
3. `accessToken` и `refreshToken` сохраняются в HTTP-only cookies.
4. Клиентские запросы идут в same-origin proxy `/api/backend/*`.
5. Proxy пробрасывает `Authorization: Bearer <access_token>` и при `401` делает одну попытку refresh.

## Moscow-only pages

### `/events`

По умолчанию каталог показывает актуальные события Москвы.

UI:

- заголовок `Мероприятия в Москве`
- subtitle про внешние источники и ReserveFlow events
- source filters: `Все`, `KudaGo`, `Timepad`, `ReserveFlow`
- booking mode filters: `Все`, `Можно забронировать`, `Внешняя ссылка`, `Demo booking`
- date filters: `Сегодня`, `Завтра`, `Выходные`, `7 дней`, `30 дней`, `180 дней`
- счётчик загруженных событий
- кнопка перехода на `/events/map`
- source badge и booking mode badge на карточке события

### `/events/map`

Карта реализована как client-only page с Leaflet.

- центр по умолчанию: Москва `55.751244, 37.618423`
- zoom по умолчанию: `11`
- используются те же фильтры, что и в каталоге
- показываются только события с координатами
- маркеры кластеризуются
- popup показывает title, date, venue, source badge и кнопку `Открыть`

### `/events/[eventId]`

Поведение зависит от `bookingMode`.

`external_link_only`:

- source badge
- кнопка `Открыть у организатора`
- сообщение о том, что бронирование внутри ReserveFlow недоступно
- адрес площадки и mini-map при наличии координат
- нет CTA на выбор мест

`demo_bookable`:

- сообщение о demo booking
- список sessions
- `Выбрать места` для bookable session

`reserveflow_managed`:

- стандартный existing ReserveFlow flow без изменений дизайна SeatMap

## Важные компоненты

- `EventList` и `EventFiltersBar`
- `EventCard`
- `EventsMapPage` и `MapCanvas`
- `EventDetails`
- `SessionCard`
- `SeatMap`
- `HoldTimer`

## API proxy

Frontend обращается не напрямую к backend, а через:

- `/api/auth/*` для auth flow
- `/api/backend/*` для бизнес-endpoint'ов

Proxy сохраняет query params, JSON body и backend error format.

## Переменные окружения

`frontend/.env.example`:

```env
BACKEND_API_URL=http://localhost:18080
NEXT_PUBLIC_APP_NAME=ReserveFlow
NEXT_PUBLIC_MAP_DEFAULT_LAT=55.751244
NEXT_PUBLIC_MAP_DEFAULT_LON=37.618423
NEXT_PUBLIC_MAP_DEFAULT_ZOOM=11
NODE_ENV=development
```

`BACKEND_API_URL` используется только на серверной стороне Next.js.
Если `frontend/.env.local` не задан, frontend в локальной среде сначала пробует `http://localhost:18080`, затем `http://localhost:8080`, и выбирает первый healthy backend ReserveFlow.

## Локальный запуск

```powershell
cd frontend
npm install
npm run dev
```

`Copy-Item .env.example .env.local` нужен только если вы хотите явно зафиксировать backend URL. По умолчанию frontend доступен на `http://localhost:3000`.

## Проверки

```powershell
npm run typecheck
npm run lint
npm run build
npm run test
```
