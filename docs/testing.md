# Testing

## Backend

Если установлен `make`:

```powershell
make test
make lint
make test-integration
```

Прямые команды, которые выполняют эти targets:

```powershell
cd backend
go test ./...
go vet ./...
go test -tags=integration ./tests
```

Интеграционные тесты используют testcontainers, поэтому нужен запущенный Docker Desktop.

## Frontend

Если установлен `make`:

```powershell
make frontend-typecheck
make frontend-lint
make frontend-build
```

Прямые команды:

```powershell
cd frontend
npm install
npm run typecheck
npm run lint
npm run build
npm run test
```

Docker image frontend собирается через `npm ci`, поэтому `package-lock.json` должен быть синхронизирован с `package.json`.

## Критичный Тест Конкуренции

`backend/tests/booking_concurrency_integration_test.go` поднимает PostgreSQL, применяет миграции и seed, затем запускает 20 параллельных попыток hold для одного `session_seat`.

Ожидаемый результат:

- 1 успешное удержание
- остальные попытки получают conflict
- ровно одна `pending` бронь
- `session_seat.status = held`

Этот тест подтверждает, что row lock в PostgreSQL защищает от double booking.

## Что Проверялось В Последнем Аудите

На Windows-окружении от 2026-05-05 напрямую прошли:

- `go test ./...`
- `go vet ./...`
- `go test -tags=integration ./tests`
- `npm install`
- `npm run typecheck`
- `npm run lint`
- `npm run build`
- `npm run test`
- `npm ci --dry-run`
- `docker compose --env-file .env.example -f deploy\docker\docker-compose.yml up -d --build`

Literal `make test`, `make test-integration`, `make frontend-typecheck`, `make frontend-lint` и `make frontend-build` в этом Windows-окружении не запускались, потому что `make` не установлен. Эквивалентные команды из targets выше проходят.

При проверке Compose локальные `8080` и `5432` были заняты сторонними процессами, поэтому использовались host overrides:

```powershell
$env:BACKEND_PORT="18080"
$env:POSTGRES_PORT="15432"
docker compose --env-file .env.example -f deploy\docker\docker-compose.yml up -d --build
```

Это не меняет значения по умолчанию: frontend остается на `3000`, Grafana на `3001`.
