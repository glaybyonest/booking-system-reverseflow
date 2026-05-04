# Deployment

## Docker Compose

```sh
cd reserveflow
make up
make migrate-up
make seed
```

Compose starts:

- `backend-api`
- `backend-worker`
- `postgres`
- `redis`
- `kafka` as Redpanda
- `prometheus`
- `grafana`

API health:

```sh
curl http://localhost:8080/health
curl http://localhost:8080/ready
```

## Kubernetes

Manifests live in `deploy/k8s`.

```sh
kubectl apply -f deploy/k8s/namespace.yaml
kubectl apply -f deploy/k8s/secret.example.yaml
kubectl apply -f deploy/k8s/
```

Replace `secret.example.yaml` values before using a shared environment.

`backend-api` runs with 2 replicas and is stateless. `backend-worker` runs with 1 replica in MVP, but the expiration job uses `FOR UPDATE SKIP LOCKED` and can be scaled if needed.

## Environment Variables

See `.env.example`.

Important values:

- `DATABASE_URL`
- `REDIS_ADDR`
- `KAFKA_BROKERS`
- `JWT_ACCESS_SECRET`
- `JWT_REFRESH_SECRET`
- `HOLD_TTL`
- `SEATMAP_CACHE_TTL`

## Healthchecks

- API liveness: `/health`
- API readiness: `/ready`
- Metrics: `/metrics`
- Postgres: `pg_isready`
- Redis: `redis-cli ping`
- Redpanda: admin readiness endpoint or `rpk cluster health`
