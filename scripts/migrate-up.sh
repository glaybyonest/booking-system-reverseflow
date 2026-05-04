#!/usr/bin/env sh
set -eu

: "${DATABASE_URL:=postgres://reserveflow:reserveflow@localhost:5432/reserveflow?sslmode=disable}"
migrate -path backend/migrations -database "$DATABASE_URL" up
