#!/usr/bin/env sh
set -eu

: "${DATABASE_URL:=postgres://reserveflow:reserveflow@localhost:5432/reserveflow?sslmode=disable}"

for seed_file in backend/seeds/dev-users.sql; do
    psql "$DATABASE_URL" -f "$seed_file"
done
