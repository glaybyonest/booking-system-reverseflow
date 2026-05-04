COMPOSE_FILE=deploy/docker/docker-compose.yml
ENV_FILE=.env.example
DATABASE_URL?=postgres://reserveflow:reserveflow@localhost:5432/reserveflow?sslmode=disable

.PHONY: up down logs api worker migrate-up migrate-down seed test test-integration lint docker-build

up:
	docker compose --env-file $(ENV_FILE) -f $(COMPOSE_FILE) up -d --build

down:
	docker compose -f $(COMPOSE_FILE) down

logs:
	docker compose -f $(COMPOSE_FILE) logs -f

api:
	cd backend && go run ./cmd/api

worker:
	cd backend && go run ./cmd/worker

migrate-up:
	migrate -path backend/migrations -database "$(DATABASE_URL)" up

migrate-down:
	migrate -path backend/migrations -database "$(DATABASE_URL)" down 1

seed:
	psql "$(DATABASE_URL)" -f backend/migrations/000002_seed.up.sql

test:
	cd backend && go test ./...

test-integration:
	cd backend && go test -tags=integration ./tests

lint:
	cd backend && go vet ./...

docker-build:
	docker build -t reserveflow-backend:local ./backend
