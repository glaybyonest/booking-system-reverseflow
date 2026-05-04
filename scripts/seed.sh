#!/usr/bin/env sh
set -eu

: "${DATABASE_URL:=postgres://reserveflow:reserveflow@localhost:5432/reserveflow?sslmode=disable}"
psql "$DATABASE_URL" -f backend/migrations/000002_seed.up.sql
