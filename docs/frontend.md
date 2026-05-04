# Frontend

Frontend ReserveFlow — приложение на Next.js App Router в директории `frontend/`.

## Технологии

- Next.js 16 App Router
- React 19
- TypeScript
- Tailwind CSS
- TanStack Query
- React Hook Form
- Zod
- Auth через HTTP-only cookie и Next.js route handlers

## Структура

- `src/app`: страницы, layout, route handlers
- `src/features`: feature-компоненты и hooks
- `src/entities`: нормализованные TypeScript-модели
- `src/widgets`: переиспользуемые продуктовые виджеты
- `src/shared`: API-клиенты, UI-примитивы, конфиг, утилиты

## Auth Flow

Браузер не хранит backend-токены в `localStorage`.

1. Браузер отправляет credentials в `/api/auth/login` или `/api/auth/register`.
2. Next.js route handler обращается к Go backend.
3. Next.js извлекает `accessToken` и `refreshToken`.
4. Токены сохраняются в HTTP-only cookie:
   - `access_token`
   - `refresh_token`
5. Клиентские компоненты работают через same-origin Next.js endpoint-ы.

Защищённые frontend route-ы:

- `/checkout/:path*`
- `/bookings/:path*`
- `/notifications/:path*`

Middleware проверяет наличие cookie; финальная валидация прав выполняется на backend.

## API Proxy

Клиент вызывает, например:

```text
/api/backend/events
/api/backend/bookings/hold
```

Прокси перенаправляет в:

```text
{BACKEND_API_URL}/api/v1/events
{BACKEND_API_URL}/api/v1/bookings/hold
```

Прокси:

- прокидывает query params
- прокидывает JSON body
- добавляет `Authorization: Bearer <access_token>`
- при `401` делает одну попытку refresh
- сохраняет формат backend-ошибок

## Страницы

- `/`: главная страница
- `/login`: вход
- `/register`: регистрация
- `/events`: каталог событий
- `/events/[eventId]`: детали события и список сеансов
- `/sessions/[sessionId]`: выбор места и hold
- `/checkout/[bookingId]`: таймер удержания и mock-оплата
- `/bookings`: история броней
- `/notifications`: список уведомлений

## Важные компоненты

- `Header`: публичный и авторизованный варианты
- `SeatMap`: группировка по рядам, статусы available/held/booked/selected
- `HoldTimer`: обратный отсчёт до истечения hold
- `BookingSummary`: унифицированный вывод брони
- базовый UI: `Button`, `Card`, `Badge`, `Input`, `Alert`, `Spinner`, `EmptyState`

## Локальный запуск

Запустите backend и зависимости:

```sh
cd reserveflow
make up
make migrate-up
make seed
```

Запустите frontend:

```sh
cd frontend
cp .env.example .env.local
npm install
npm run dev
```

Откройте `http://localhost:3000`.

Если порт `3000` занят, используйте `npm run dev -- -p 3001`.

## Переменные окружения

`frontend/.env.example`:

```env
BACKEND_API_URL=http://localhost:8080
NEXT_PUBLIC_APP_NAME=ReserveFlow
NODE_ENV=development
```

`BACKEND_API_URL` используется только на серверной стороне и не должен содержать секреты.

## Команды качества

```sh
npm run lint
npm run typecheck
npm run build
npm run test
```
