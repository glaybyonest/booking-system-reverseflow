# Deployment

## Локальный Docker Compose

Рекомендуемый сценарий на Windows:

```powershell
cd C:\development\booking-system\reserveflow
powershell -ExecutionPolicy Bypass -File scripts\dev-up.ps1
```

Скрипт:

- поднимает `frontend`, `backend-api`, `backend-worker`, `postgres`, `redis`, `kafka`, `prometheus`, `grafana`
- применяет миграции
- загружает demo seed

## Порты по умолчанию

| Сервис | Адрес |
| --- | --- |
| Frontend | `http://localhost:3000` |
| Backend API | `http://localhost:18080/api/v1` |
| Backend health | `http://localhost:18080/health` |
| Postgres | `localhost:5432` |
| Redis | `localhost:6379` |
| Kafka | `localhost:19092` |
| Prometheus | `http://localhost:9090` |
| Grafana | `http://localhost:3001` |

Порт `18080` — канонический host-порт Docker Compose. Порт `8080` сохраняется для direct standalone backend без Compose.

## Moscow sync env vars

Ключевые переменные для внешнего импорта:

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
```

Дополнительные флаги, оставленные выключенными:

```env
TICKETMASTER_ENABLED=false
PRO_CULTURE_ENABLED=false
DATA_MOS_ENABLED=false
```

## Включение scheduled worker sync

По умолчанию worker sync выключен.

Для периодического импорта Москвы:

```env
EXTERNAL_SYNC_ENABLED=true
EXTERNAL_SYNC_INTERVAL=6h
```

Worker:

- запускает sync сразу после старта
- затем повторяет его по таймеру
- синхронизирует KudaGo и Timepad независимо
- не падает при ошибке одного провайдера
- записывает failed import runs

## Manual sync

Нужен `admin` JWT.

```bash
curl -X POST http://localhost:18080/api/v1/admin/integrations/sync/moscow \
  -H "Authorization: Bearer <admin-access-token>" \
  -H "Content-Type: application/json" \
  -d '{"providers":["kudago","timepad"],"daysAhead":180,"lookbackDays":14}'
```

## Compose вручную

```powershell
docker compose --env-file .env.example -f deploy\docker\docker-compose.yml up -d --build
docker run --rm --network reserveflow_reserveflow -v "${PWD}\backend\migrations:/migrations:ro" migrate/migrate:v4.18.2 -path=/migrations -database "postgres://reserveflow:reserveflow@postgres:5432/reserveflow?sslmode=disable" up
docker compose --env-file .env.example -f deploy\docker\docker-compose.yml cp backend\seeds\dev-users.sql postgres:/tmp/reserveflow-dev-users.sql
docker compose --env-file .env.example -f deploy\docker\docker-compose.yml exec -T postgres psql -U reserveflow -d reserveflow -v ON_ERROR_STOP=1 -f /tmp/reserveflow-dev-users.sql
docker compose --env-file .env.example -f deploy\docker\docker-compose.yml cp backend\seeds\dev-catalog.sql postgres:/tmp/reserveflow-dev-catalog.sql
docker compose --env-file .env.example -f deploy\docker\docker-compose.yml exec -T postgres psql -U reserveflow -d reserveflow -v ON_ERROR_STOP=1 -f /tmp/reserveflow-dev-catalog.sql
```

## Kubernetes

Манифесты находятся в `deploy/k8s`.

```powershell
kubectl apply -f deploy/k8s/namespace.yaml
kubectl apply -f deploy/k8s/secret.example.yaml
kubectl apply -f deploy/k8s/
```

Перед публикацией в общее окружение:

- замените все секреты
- не коммитьте реальные API keys и JWT secrets
- оставляйте внешние Moscow sync значения явными, чтобы избежать cross-city imports

## Healthchecks

- API liveness: `/health`
- API readiness: `/ready`
- API metrics: `/metrics`
- Postgres: `pg_isready`
- Redis: `redis-cli ping`
- Redpanda: `rpk cluster health`
