# Deployment

## Docker Compose

```sh
cd reserveflow
make up
make migrate-up
make seed
```

Compose поднимает:

- `frontend`
- `backend-api`
- `backend-worker`
- `postgres`
- `redis`
- `kafka` (Redpanda)
- `prometheus`
- `grafana`

Проверка API:

```sh
curl http://localhost:8080/health
curl http://localhost:8080/ready
```

Frontend:

```sh
open http://localhost:3000
```

## Kubernetes

Манифесты находятся в `deploy/k8s`.

```sh
kubectl apply -f deploy/k8s/namespace.yaml
kubectl apply -f deploy/k8s/secret.example.yaml
kubectl apply -f deploy/k8s/
```

Перед использованием в общем окружении обновите значения в `secret.example.yaml`.

- `frontend`: 2 реплики, проксирует backend-запросы в сервис `backend-api`
- `backend-api`: 2 stateless реплики
- `backend-worker`: 1 реплика в MVP (при необходимости масштабируется, т.к. expire job использует `FOR UPDATE SKIP LOCKED`)

Ingress:

- `reserveflow.local/` -> frontend
- `reserveflow.local/api/v1/*` -> backend API

## Переменные окружения

Смотрите `.env.example`.

Ключевые параметры:

- `DATABASE_URL`
- `REDIS_ADDR`
- `KAFKA_BROKERS`
- `JWT_ACCESS_SECRET`
- `JWT_REFRESH_SECRET`
- `HOLD_TTL`
- `SEATMAP_CACHE_TTL`
- frontend: `BACKEND_API_URL`

## Healthchecks

- liveness API: `/health`
- readiness API: `/ready`
- метрики: `/metrics`
- Postgres: `pg_isready`
- Redis: `redis-cli ping`
- Redpanda: readiness endpoint admin API или `rpk cluster health`
